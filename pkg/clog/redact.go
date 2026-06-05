// Package clog: centralized PII redaction for log messages.
//
// LAS-1488 layer #1: every log line flows through processEvent, which calls
// Redact on the final formatted message string exactly once before it fans out
// to dedupe and all sinks (console, file, BetterStack/Postgres). Scrubbing here
// catches caller PII -- email addresses, IPv4 (incl. media IPs), and the caller
// phone embedded in participant_id=<CID>@<ip> / conference_id=<DID>@<ip> -- in a
// single place, including future call sites.
//
// Redaction is ON by default. Disable for local development with the env var
// CORALIE_LOG_REDACT=0 (or "false"/"no"/"off"), or programmatically via
// SetRedactionEnabled(false). The default redactor and its pattern set can be
// replaced wholesale with SetRedactor for ops tuning.
package clog

import (
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

// redactEnvVar is the environment variable that toggles redaction. It is read
// once at package init. Any of "0", "false", "no", "off" (case-insensitive)
// disables redaction; anything else (including unset) leaves it enabled.
const redactEnvVar = "CORALIE_LOG_REDACT"

// pattern is one ordered redaction rule: a precompiled regex and the literal
// string that replaces every match. Compiled once at package init.
type pattern struct {
	name        string
	re          *regexp.Regexp
	replacement string
}

// Redactor scrubs PII from a string by applying an ordered list of precompiled
// patterns in a single pass each. Order matters (see defaultPatterns): email and
// IPv4 are redacted before phone so that digits living inside an email or IP are
// not partially consumed by the phone rule. A Redactor is immutable after
// construction and therefore safe for concurrent use.
type Redactor struct {
	patterns []pattern
}

var (
	// defaultRedactor is the package-level redactor used by the Redact helper
	// and by the logging hot path (processEvent). Built once at init.
	defaultRedactor *Redactor
	defaultRedactMu sync.RWMutex

	// redactEnabled gates redaction globally. Read once from the environment at
	// init, then mutable via SetRedactionEnabled. Stored as int32 (0/1) for
	// lock-free reads on the hot path.
	redactEnabled atomic.Bool
)

func init() {
	defaultRedactor = NewDefaultRedactor()
	redactEnabled.Store(redactEnabledFromEnv())
}

// redactEnabledFromEnv returns whether redaction should be enabled based on the
// CORALIE_LOG_REDACT environment variable. Default (unset or unrecognized) is
// enabled; only explicit falsey values disable it.
func redactEnabledFromEnv() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(redactEnvVar))) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// defaultPatterns returns the ordered default PII pattern set. Ordering rationale:
//
//  1. email first  -- a conservative addr@host.tld match. Redacting it before the
//     phone rule prevents the local-part / domain digits of an email from being
//     half-eaten by the phone rule.
//  2. IPv4 (optional :port) second -- four dotted octets. Done before phone so an
//     IP's octets are never mistaken for phone digits. Intentionally covers media
//     IPs (e.g. RTP dest=10.0.0.5:4000) per LAS-1488. The :port suffix is dropped
//     into the <ip> token so "10.0.0.5:4000" -> "<ip>".
//  3. phone E.164 ("+" prefixed) third -- conservative: a leading "+" then 7-15
//     digits. This is the canonical carrier-supplied caller number form.
//  4. phone digits@ fourth -- a run of 7-15 digits immediately followed by "@".
//     This is exactly the participant_id=<CID>@<ip> / conference_id=<DID>@<ip>
//     form that leaks the caller MSISDN, and is the actual measured leak. Go's
//     RE2 engine has no lookahead, so the "@" is matched and re-emitted in the
//     replacement ("<phone>@") rather than asserted with (?=@).
//
// Bare in-sentence digit runs (no "+" and no trailing "@") are deliberately NOT
// redacted: doing so clobbers common non-PII numbers in logs (epoch-millis
// timestamps, port=5060, samples=480, byte counts, frame counters, version
// strings, UUID segments). The structural fix for caller-ID leakage is
// CID-hashing (LAS-1482, a separate follow-up). Correctness over breadth.
func defaultPatterns() []pattern {
	return []pattern{
		{
			name:        "email",
			re:          regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`),
			replacement: "<email>",
		},
		{
			name:        "ipv4",
			re:          regexp.MustCompile(`\b(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}(?::\d{1,5})?\b`),
			replacement: "<ip>",
		},
		{
			name:        "phone_e164",
			re:          regexp.MustCompile(`\+\d{7,15}\b`),
			replacement: "<phone>",
		},
		{
			name:        "phone_at",
			re:          regexp.MustCompile(`\b\d{7,15}@`),
			replacement: "<phone>@",
		},
	}
}

// NewDefaultRedactor returns a Redactor preloaded with the default PII pattern
// set (email, IPv4, +E.164 phone, digits@ phone) in the documented order.
func NewDefaultRedactor() *Redactor {
	return &Redactor{patterns: defaultPatterns()}
}

// NewRedactor returns a Redactor with a caller-supplied ordered pattern set,
// each entry being a regex source string and its replacement. Patterns are
// applied in the given order; compile errors panic (call sites supply static
// regexes). Use this to fully customize redaction for ops tuning.
func NewRedactor(patterns []RedactPattern) *Redactor {
	compiled := make([]pattern, 0, len(patterns))
	for _, p := range patterns {
		compiled = append(compiled, pattern{
			name:        p.Name,
			re:          regexp.MustCompile(p.Regex),
			replacement: p.Replacement,
		})
	}
	return &Redactor{patterns: compiled}
}

// RedactPattern is an exported, ordered redaction rule used to build a custom
// Redactor via NewRedactor. Regex is a Go regexp source string; Replacement is
// the literal text substituted for each match.
type RedactPattern struct {
	Name        string
	Regex       string
	Replacement string
}

// Redact applies every pattern in order and returns the scrubbed string. It is
// safe for concurrent use. The redaction is idempotent for the default pattern
// set: the replacement tokens (<email>, <ip>, <phone>) contain no characters
// that any pattern matches, so redacting an already-redacted string is a no-op.
func (r *Redactor) Redact(s string) string {
	if s == "" {
		return s
	}
	for i := range r.patterns {
		p := &r.patterns[i]
		// ReplaceAllString returns the input unchanged (no new allocation) when
		// there is no match, keeping the no-PII common case cheap.
		s = p.re.ReplaceAllString(s, p.replacement)
	}
	return s
}

// Redact scrubs PII from s using the package-level default redactor, honoring
// the global enable toggle. When redaction is disabled it returns s unchanged.
// This is the helper used at the logging choke point (processEvent).
func Redact(s string) string {
	if !redactEnabled.Load() {
		return s
	}
	defaultRedactMu.RLock()
	r := defaultRedactor
	defaultRedactMu.RUnlock()
	return r.Redact(s)
}

// SetRedactionEnabled turns global PII redaction on or off at runtime. Redaction
// is enabled by default (see CORALIE_LOG_REDACT). Disabling is intended for
// local development only.
func SetRedactionEnabled(enabled bool) {
	redactEnabled.Store(enabled)
}

// RedactionEnabled reports whether global PII redaction is currently enabled.
func RedactionEnabled() bool {
	return redactEnabled.Load()
}

// SetRedactor replaces the package-level default redactor used by Redact. Pass a
// custom Redactor (e.g. from NewRedactor) to tune patterns at runtime. A nil
// argument is ignored.
func SetRedactor(r *Redactor) {
	if r == nil {
		return
	}
	defaultRedactMu.Lock()
	defaultRedactor = r
	defaultRedactMu.Unlock()
}
