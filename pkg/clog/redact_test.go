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
