// Package clog: tests for the centralized PII redactor.
package clog

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestRedactTruePositives verifies PII is scrubbed in the forms LAS-1488 targets.
func TestRedactTruePositives(t *testing.T) {
	r := NewDefaultRedactor()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"email_simple", "a@b.com", "<email>"},
		{"email_in_sentence", "user alice@example.co.uk failed login", "user <email> failed login"},
		{"email_with_digits", "u123.foo+tag@mail-server.example.com", "<email>"},
		{"phone_e164", "+358401234567", "<phone>"},
		{"phone_e164_in_sentence", "caller +358401234567 connected", "caller <phone> connected"},
		{"phone_at_ip", "358401234567@10.0.0.19", "<phone>@<ip>"},
		{
			"participant_id_with_suffix",
			"participant_id=358401234567@10.0.0.19-abc",
			"participant_id=<phone>@<ip>-abc",
		},
		{
			"conference_id",
			"conference_id=4915112345678@10.0.0.14:5066 joined",
			"conference_id=<phone>@<ip> joined",
		},
		{"ipv4_plain", "RTP dest=10.0.0.5 ready", "RTP dest=<ip> ready"},
		{"ipv4_with_port", "RTP dest=10.0.0.5:4000", "RTP dest=<ip>"},
		{
			"realistic_participant_line",
			"[INFO][SIP] participant_id=358401234567@10.0.0.19 media=10.0.0.5:4000 codec=opus",
			"[INFO][SIP] participant_id=<phone>@<ip> media=<ip> codec=opus",
		},
		{
			"realistic_conference_line",
			"conference_id=358501119999@192.168.1.7 participant_id=358401234567@10.0.0.19",
			"conference_id=<phone>@<ip> participant_id=<phone>@<ip>",
		},
		{"email_and_phone_and_ip", "from a@b.com via +123456789 at 8.8.8.8", "from <email> via <phone> at <ip>"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := r.Redact(c.in); got != c.want {
				t.Errorf("Redact(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestRedactFalsePositiveGuards proves common non-PII numbers in logs are NOT
// touched. These are the values the phone rule must never clobber.
func TestRedactFalsePositiveGuards(t *testing.T) {
	r := NewDefaultRedactor()
	cases := []struct {
		name string
		in   string // must be returned unchanged
	}{
		{"epoch_millis_13", "ts=1717589445123 event=ringing"},
		{"epoch_millis_bare", "1717589445123"},
		{"port", "port=5060"},
		{"samples", "samples=480"},
		{"bytes", "bytes=1048576"},
		{"frame_counter", "frame=1234567"},
		{"uuid", "id=550e8400-e29b-41d4-a716-446655440000"},
		{"version_v", "v1.26.4"},
		{"version_go", "go 1.26.4"},
		{"short_number", "count=42"},
		{"sequence", "seq=123456 of 999999"},
		{"plain_sentence", "no PII here at all"},
		{"big_bare_number", "total=999999999999"}, // 12 bare digits, no '+' / '@'
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := r.Redact(c.in); got != c.in {
				t.Errorf("Redact(%q) = %q, want unchanged (false positive)", c.in, got)
			}
		})
	}
}

// TestRedactIdempotent verifies redacting twice equals redacting once.
func TestRedactIdempotent(t *testing.T) {
	r := NewDefaultRedactor()
	inputs := []string{
		"participant_id=358401234567@10.0.0.19 from a@b.com via +358401234567 at 10.0.0.5:4000",
		"plain line with no pii",
		"port=5060 samples=480 ts=1717589445123",
	}
	for _, in := range inputs {
		once := r.Redact(in)
		twice := r.Redact(once)
		if once != twice {
			t.Errorf("not idempotent: once=%q twice=%q", once, twice)
		}
	}
}

// TestRedactEmptyString verifies the empty-string fast path.
func TestRedactEmptyString(t *testing.T) {
	if got := NewDefaultRedactor().Redact(""); got != "" {
		t.Errorf("Redact(\"\") = %q, want \"\"", got)
	}
}

// TestRedactToggleDisabled verifies the package-level helper passes through when
// redaction is disabled, and restores afterward.
func TestRedactToggleDisabled(t *testing.T) {
	prev := RedactionEnabled()
	defer SetRedactionEnabled(prev)

	in := "participant_id=358401234567@10.0.0.19"

	SetRedactionEnabled(true)
	if got := Redact(in); got != "participant_id=<phone>@<ip>" {
		t.Fatalf("enabled: Redact(%q) = %q, want redacted", in, got)
	}

	SetRedactionEnabled(false)
	if got := Redact(in); got != in {
		t.Fatalf("disabled: Redact(%q) = %q, want passthrough", in, got)
	}
}

// TestRedactEnabledFromEnv covers env-var parsing for the toggle default.
func TestRedactEnabledFromEnv(t *testing.T) {
	cases := map[string]bool{
		"":       true,
		"1":      true,
		"true":   true,
		"yes":    true,
		"on":     true,
		"random": true,
		"0":      false,
		"false":  false,
		"FALSE":  false,
		"No":     false,
		"off":    false,
		" off ":  false,
	}
	for v, want := range cases {
		t.Setenv(redactEnvVar, v)
		if got := redactEnabledFromEnv(); got != want {
			t.Errorf("redactEnabledFromEnv() with %q = %v, want %v", v, got, want)
		}
	}
}

// TestSetRedactor verifies the default redactor can be replaced and that nil is
// ignored.
func TestSetRedactor(t *testing.T) {
	prevEnabled := RedactionEnabled()
	defer SetRedactionEnabled(prevEnabled)
	SetRedactionEnabled(true)

	defaultRedactMu.RLock()
	orig := defaultRedactor
	defaultRedactMu.RUnlock()
	defer SetRedactor(orig)

	custom := NewRedactor([]RedactPattern{
		{Name: "secret", Regex: `SEKRET-\d+`, Replacement: "<secret>"},
	})
	SetRedactor(custom)
	if got := Redact("token SEKRET-12345 here"); got != "token <secret> here" {
		t.Errorf("custom redactor not applied: got %q", got)
	}
	// Default email rule is gone now (custom set has only the secret rule).
	if got := Redact("a@b.com"); got != "a@b.com" {
		t.Errorf("custom redactor should not redact email: got %q", got)
	}

	// nil is ignored: custom remains in effect.
	SetRedactor(nil)
	if got := Redact("SEKRET-9"); got != "<secret>" {
		t.Errorf("SetRedactor(nil) should be ignored, got %q", got)
	}
}

// captureSink records every formatted string written to it, proving what the
// sink layer actually receives.
type captureSink struct {
	mu   sync.Mutex
	msgs []string
}

func (c *captureSink) Write(level Level, iface, formatted string) {
	c.mu.Lock()
	c.msgs = append(c.msgs, formatted)
	c.mu.Unlock()
}
func (c *captureSink) Flush() {}
func (c *captureSink) Close() {}

func (c *captureSink) snapshot() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.msgs))
	copy(out, c.msgs)
	return out
}

// TestRedactionThroughSinkPath wires a sink into a real agent and proves the
// emitted (formatted) string is redacted by the time it reaches the sink. This
// proves the hook in processEvent, not just the helper. It exercises the same
// path every sink (incl. BetterStack/Postgres) consumes.
func TestRedactionThroughSinkPath(t *testing.T) {
	prevEnabled := RedactionEnabled()
	defer SetRedactionEnabled(prevEnabled)
	SetRedactionEnabled(true)

	sink := &captureSink{}
	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.Dedupe.Enabled = false // avoid suppression collapsing our two lines

	a, err := newAgent(cfg)
	if err != nil {
		t.Fatalf("newAgent: %v", err)
	}
	a.sinks = append(a.sinks, sink)
	defer a.stop(context.Background())

	a.enqueue(Event{
		Level:   LevelInfo,
		Iface:   "SIP",
		Message: "participant_id=%s media=%s",
		Params:  []interface{}{"358401234567@10.0.0.19", "10.0.0.5:4000"},
	})
	a.enqueue(Event{
		Level:   LevelInfo,
		Iface:   "App",
		Message: "contact a@b.com phone +358401234567",
	})

	// Poll for delivery (async agent).
	deadline := time.Now().Add(2 * time.Second)
	var got []string
	for time.Now().Before(deadline) {
		got = sink.snapshot()
		if len(got) >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(got) < 2 {
		t.Fatalf("expected 2 messages at sink, got %d: %v", len(got), got)
	}

	want0 := "participant_id=<phone>@<ip> media=<ip>"
	want1 := "contact <email> phone <phone>"
	if got[0] != want0 {
		t.Errorf("sink msg[0] = %q, want %q", got[0], want0)
	}
	if got[1] != want1 {
		t.Errorf("sink msg[1] = %q, want %q", got[1], want1)
	}
	// Hard guarantee: no raw MSISDN/email/IP reaches the sink.
	for _, m := range got {
		for _, leak := range []string{"358401234567", "a@b.com", "10.0.0.19", "10.0.0.5"} {
			if strings.Contains(m, leak) {
				t.Errorf("raw PII %q leaked to sink in %q", leak, m)
			}
		}
	}
}

// TestRedactEventClonesAndScrubs verifies RedactEvent scrubs the Message template
// and every string-valued param, leaves non-string params untouched, and never
// mutates the caller's Event or Params slice.
func TestRedactEventClonesAndScrubs(t *testing.T) {
	r := NewDefaultRedactor()

	origParams := []interface{}{
		"participant_id=358401234567@10.0.0.19", // string with PII -> redacted
		42,                                      // int -> untouched
		"contact alice@example.com",             // string with PII -> redacted
		struct{ N int }{N: 7},                   // struct -> untouched
	}
	e := Event{
		Level:   LevelInfo,
		Iface:   "SIP",
		Message: "caller +358401234567 via %s n=%d",
		Params:  origParams,
	}

	out := r.RedactEvent(e)

	if out.Message != "caller <phone> via %s n=%d" {
		t.Errorf("Message = %q, want redacted", out.Message)
	}
	if out.Params[0] != "participant_id=<phone>@<ip>" {
		t.Errorf("Params[0] = %v, want redacted", out.Params[0])
	}
	if out.Params[1] != 42 {
		t.Errorf("Params[1] = %v, want 42 (untouched int)", out.Params[1])
	}
	if out.Params[2] != "contact <email>" {
		t.Errorf("Params[2] = %v, want redacted email", out.Params[2])
	}
	if out.Params[3] != (struct{ N int }{N: 7}) {
		t.Errorf("Params[3] = %v, want untouched struct", out.Params[3])
	}

	// Caller's Event and slice must be untouched (clone semantics).
	if e.Message != "caller +358401234567 via %s n=%d" {
		t.Errorf("caller Message mutated: %q", e.Message)
	}
	if origParams[0] != "participant_id=358401234567@10.0.0.19" {
		t.Errorf("caller Params[0] mutated: %v", origParams[0])
	}
	if &out.Params[0] == &origParams[0] {
		t.Error("RedactEvent must allocate a new Params slice, not alias the caller's")
	}
}

// TestRedactEventToggleDisabled verifies the package-level RedactEvent passes the
// Event through unchanged when redaction is disabled.
func TestRedactEventToggleDisabled(t *testing.T) {
	prev := RedactionEnabled()
	defer SetRedactionEnabled(prev)

	e := Event{
		Level:   LevelInfo,
		Iface:   "SIP",
		Message: "participant_id=%s",
		Params:  []interface{}{"358401234567@10.0.0.19"},
	}

	SetRedactionEnabled(false)
	out := RedactEvent(e)
	if out.Message != e.Message || out.Params[0] != "358401234567@10.0.0.19" {
		t.Errorf("disabled RedactEvent should pass through, got %+v", out)
	}

	SetRedactionEnabled(true)
	out = RedactEvent(e)
	if out.Params[0] != "<phone>@<ip>" {
		t.Errorf("enabled RedactEvent param = %v, want redacted", out.Params[0])
	}
}

// captureHook records every Event handed to OnLog, proving exactly what a hook
// (which may forward to Sentry/Datadog/webhooks) actually receives.
type captureHook struct {
	mu     sync.Mutex
	events []Event
}

func (h *captureHook) OnLog(e Event) {
	h.mu.Lock()
	h.events = append(h.events, e)
	h.mu.Unlock()
}

func (h *captureHook) snapshot() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]Event, len(h.events))
	copy(out, h.events)
	return out
}

// TestHookReceivesRedactedEvent proves Fix #1: a Global hook receives an Event
// whose Message AND string params are redacted (no raw PII), with non-string
// params untouched. This guards the hook PII-leak path through the real agent.
func TestHookReceivesRedactedEvent(t *testing.T) {
	prevEnabled := RedactionEnabled()
	defer SetRedactionEnabled(prevEnabled)
	SetRedactionEnabled(true)

	hook := &captureHook{}
	cfg := DefaultConfig()
	cfg.Console.Enabled = false
	cfg.Dedupe.Enabled = false
	cfg.Hooks.Global = []Hook{hook}

	a, err := newAgent(cfg)
	if err != nil {
		t.Fatalf("newAgent: %v", err)
	}
	defer a.stop(context.Background())

	a.enqueue(Event{
		Level:   LevelInfo,
		Iface:   "SIP",
		Message: "caller +358401234567 participant_id=%s email=%s seq=%d",
		Params:  []interface{}{"358401234567@10.0.0.19", "alice@example.com", 7},
	})

	deadline := time.Now().Add(2 * time.Second)
	var got []Event
	for time.Now().Before(deadline) {
		got = hook.snapshot()
		if len(got) >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(got) < 1 {
		t.Fatalf("hook received no events")
	}

	e := got[0]
	if e.Message != "caller <phone> participant_id=%s email=%s seq=%d" {
		t.Errorf("hook Message not redacted: %q", e.Message)
	}
	// Params are the bare PII values (the participant_id=/email= literals live in
	// the Message template), so they redact to the bare tokens.
	if e.Params[0] != "<phone>@<ip>" {
		t.Errorf("hook Params[0] not redacted: %v", e.Params[0])
	}
	if e.Params[1] != "<email>" {
		t.Errorf("hook Params[1] not redacted: %v", e.Params[1])
	}
	if e.Params[2] != 7 {
		t.Errorf("hook Params[2] = %v, want 7 (untouched int)", e.Params[2])
	}

	// Hard guarantee: no raw PII anywhere in the Event handed to the hook.
	hay := e.Message
	for _, p := range e.Params {
		if s, ok := p.(string); ok {
			hay += "\x00" + s
		}
	}
	for _, leak := range []string{"358401234567", "alice@example.com", "10.0.0.19"} {
		if strings.Contains(hay, leak) {
			t.Errorf("raw PII %q leaked to hook", leak)
		}
	}
}

// TestDedupePreservesDistinctCallers proves Fix #2: two events that differ ONLY in
// pre-redaction PII are NOT deduped (both reach the sink, redacted), while two
// truly-identical events ARE deduped. Uses a capturing sink and the real agent.
func TestDedupePreservesDistinctCallers(t *testing.T) {
	prevEnabled := RedactionEnabled()
	defer SetRedactionEnabled(prevEnabled)
	SetRedactionEnabled(true)

	t.Run("distinct_callers_not_deduped", func(t *testing.T) {
		sink := &captureSink{}
		cfg := DefaultConfig()
		cfg.Console.Enabled = false
		cfg.Dedupe.Enabled = true // dedupe ON: the whole point is it must not collapse these
		cfg.Dedupe.SummaryFormat = "last message repeated %d more times"

		a, err := newAgent(cfg)
		if err != nil {
			t.Fatalf("newAgent: %v", err)
		}
		a.sinks = append(a.sinks, sink)
		defer a.stop(context.Background())

		// Same template + iface + level, different caller PID (pre-redaction).
		a.enqueue(Event{Level: LevelInfo, Iface: "SIP", Message: "joined participant_id=%s", Params: []interface{}{"1111111@10.0.0.1"}})
		a.enqueue(Event{Level: LevelInfo, Iface: "SIP", Message: "joined participant_id=%s", Params: []interface{}{"2222222@10.0.0.2"}})

		deadline := time.Now().Add(2 * time.Second)
		var got []string
		for time.Now().Before(deadline) {
			got = sink.snapshot()
			if len(got) >= 2 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if len(got) < 2 {
			t.Fatalf("distinct callers were deduped: got %d msgs %v, want 2", len(got), got)
		}
		// Both redacted at the sink, and neither raw MSISDN leaked.
		for _, m := range got {
			for _, leak := range []string{"1111111", "2222222", "10.0.0.1", "10.0.0.2"} {
				if strings.Contains(m, leak) {
					t.Errorf("raw PII %q leaked to sink: %q", leak, m)
				}
			}
		}
	})

	t.Run("identical_events_deduped", func(t *testing.T) {
		sink := &captureSink{}
		cfg := DefaultConfig()
		cfg.Console.Enabled = false
		cfg.Dedupe.Enabled = true
		cfg.Dedupe.SummaryFormat = "last message repeated %d more times"

		a, err := newAgent(cfg)
		if err != nil {
			t.Fatalf("newAgent: %v", err)
		}
		a.sinks = append(a.sinks, sink)
		defer a.stop(context.Background())

		// Two truly-identical events, then a different one to flush the summary.
		ev := Event{Level: LevelInfo, Iface: "SIP", Message: "heartbeat participant_id=%s", Params: []interface{}{"1111111@10.0.0.1"}}
		a.enqueue(ev)
		a.enqueue(ev)
		a.enqueue(Event{Level: LevelInfo, Iface: "SIP", Message: "done"})

		deadline := time.Now().Add(2 * time.Second)
		var got []string
		for time.Now().Before(deadline) {
			got = sink.snapshot()
			// expect: 1 redacted heartbeat + 1 summary + 1 "done" = 3
			if len(got) >= 3 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		first := ""
		count := 0
		for _, m := range got {
			if m == "heartbeat participant_id=<phone>@<ip>" {
				count++
				first = m
			}
		}
		if count != 1 {
			t.Errorf("identical events not deduped: heartbeat appeared %d times in %v", count, got)
		}
		_ = first
		// The summary must be count-only (no raw message, no PII).
		foundSummary := false
		for _, m := range got {
			if strings.Contains(m, "repeated") {
				foundSummary = true
				for _, leak := range []string{"1111111", "10.0.0.1"} {
					if strings.Contains(m, leak) {
						t.Errorf("dedupe summary leaked PII %q: %q", leak, m)
					}
				}
			}
		}
		if !foundSummary {
			t.Errorf("expected a dedupe summary line in %v", got)
		}
	})
}

// TestRedactDoSBound proves Fix #3: an over-long input is truncated (length capped
// near maxRedactLen, marker present) and PII near the start is still redacted.
func TestRedactDoSBound(t *testing.T) {
	r := NewDefaultRedactor()

	// PII at the very start, then a giant attacker-controlled blob well past 64KiB.
	prefix := "participant_id=358401234567@10.0.0.19 "
	blob := strings.Repeat("A", maxRedactLen+5000)
	in := prefix + blob

	got := r.Redact(in)

	// PII near the start is redacted.
	if !strings.HasPrefix(got, "participant_id=<phone>@<ip> ") {
		t.Errorf("leading PII not redacted; got prefix %q", got[:min(60, len(got))])
	}
	if strings.Contains(got, "358401234567") || strings.Contains(got, "10.0.0.19") {
		t.Errorf("raw PII survived in truncated output")
	}
	// Truncation marker present.
	if !strings.Contains(got, "<truncated ") {
		t.Errorf("expected truncation marker, got tail %q", got[max(0, len(got)-40):])
	}
	// Output length is bounded: at most the limit plus a short marker (and the
	// redaction tokens are shorter than the PII they replaced, so this holds).
	if len(got) > maxRedactLen+64 {
		t.Errorf("output not bounded: len=%d, want <= %d", len(got), maxRedactLen+64)
	}
}

// BenchmarkRedact measures the hot-path cost on a representative log line that
// contains PII (the worst case: every pattern matches).
func BenchmarkRedact(b *testing.B) {
	r := NewDefaultRedactor()
	line := "[INFO][SIP] participant_id=358401234567@10.0.0.19 media=10.0.0.5:4000 " +
		"contact alice@example.com via +358401234567 ts=1717589445123 port=5060 samples=480"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Redact(line)
	}
}

// BenchmarkRedactNoPII measures the common case: a line with no PII (no match,
// no allocation expected).
func BenchmarkRedactNoPII(b *testing.B) {
	r := NewDefaultRedactor()
	line := "[INFO][Mixer] frame=1234567 samples=480 port=5060 bytes=1048576 ts=1717589445123 ok"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Redact(line)
	}
}
