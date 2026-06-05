// Package clog: centralized PII redaction for log messages.
//
// LAS-1488 layer #1: every log line flows through processEvent, which redacts the
// final formatted message string exactly once before it fans out to all sinks
// (console, file, BetterStack/Postgres) and to hooks. Note the ORDER: processEvent
// dedupes on the RAW formatted string FIRST (so two distinct callers that differ
// only in PII are not collapsed into one dedupe key), and only redacts after the
// dedupe suppress check passes. Scrubbing here catches caller PII -- email
// addresses, IPv4 (incl. media IPs), and the caller phone embedded in
// participant_id=<CID>@<ip> / conference_id=<DID>@<ip> -- in a single place,
// including future call sites.
//
// Redaction is ON by default. Disable for local development with the env var
// CORALIE_LOG_REDACT=0 (or "false"/"no"/"off"), or programmatically via
// SetRedactionEnabled(false). The default redactor and its pattern set can be
// replaced wholesale with SetRedactor for ops tuning. See redact_config.go for the
// default pattern set, env parsing, and the package-level toggle helpers.
package clog

import (
	"fmt"
	"regexp"
)

// maxRedactLen bounds the input size the redactor will scan in a single call.
// The four PII regexes run over the whole string in the single agent goroutine,
// so an attacker-controlled field (a huge SIP header, a bloated User-Agent, a
// malformed packet dumped on error) could otherwise stall the agent and cause
// queue drops. Inputs longer than this are truncated to the limit (with a marker
// appended) before the patterns run, capping the per-call work. 64 KiB is far
// larger than any legitimate log line yet small enough to keep redaction cheap.
const maxRedactLen = 64 * 1024

// boundString truncates s to maxRedactLen bytes and appends a marker summarizing
// the dropped tail. It is the single helper used both to bound oversized %s params
// before formatting (processEvent.boundParams) and as Redact's own backstop, so
// the truncation form is identical in both places.
func boundString(s string) string {
	if len(s) <= maxRedactLen {
		return s
	}
	dropped := len(s) - maxRedactLen
	return s[:maxRedactLen] + fmt.Sprintf("…<truncated %d bytes>", dropped)
}

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

// RedactPattern is an exported, ordered redaction rule used to build a custom
// Redactor via NewRedactor. Regex is a Go regexp source string; Replacement is
// the literal text substituted for each match.
type RedactPattern struct {
	Name        string
	Regex       string
	Replacement string
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

// Redact applies every pattern in order and returns the scrubbed string. It is
// safe for concurrent use. The redaction is idempotent for the default pattern
// set: the replacement tokens (<email>, <ip>, <phone>) contain no characters
// that any pattern matches, so redacting an already-redacted string is a no-op.
func (r *Redactor) Redact(s string) string {
	if s == "" {
		return s
	}
	// DoS bound (backstop): cap the work the regex engine does per call. The hot
	// path bounds oversized params before formatting (see processEvent), but Redact
	// keeps its own guard so a single attacker-controlled field can never stall the
	// agent goroutine even when called directly. The dropped tail is summarized in
	// a marker; PII near the start (where participant_id/conference_id live) is
	// still redacted because the truncated prefix is what the patterns run over.
	s = boundString(s)
	for i := range r.patterns {
		p := &r.patterns[i]
		// ReplaceAllString returns the input unchanged (no new allocation) when
		// there is no match, keeping the no-PII common case cheap.
		s = p.re.ReplaceAllString(s, p.replacement)
	}
	return s
}
