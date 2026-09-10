// Package clog: source-fixed capture-contract expectation constructors.
package clog

import (
	"fmt"
	"reflect"
)

type captureVectorExpectation struct {
	input, context, expected map[string]any
}

func sourceFixedVector(id string) (captureVectorExpectation, bool) {
	if vector, ok := sourceSecurityVectors()[id]; ok {
		return vector, true
	}
	vector, ok := sourceContractVectors()[id]
	return vector, ok
}

func validateSourceFixedVector(id string, vector map[string]any) error {
	want, ok := sourceFixedVector(id)
	if !ok {
		return fmt.Errorf("no source-fixed expectation for %s", id)
	}
	if !reflect.DeepEqual(vector["input"], want.input) {
		return fmt.Errorf("%s input changed", id)
	}
	if !reflect.DeepEqual(vector["context"], want.context) {
		return fmt.Errorf("%s context changed", id)
	}
	if !reflect.DeepEqual(vector["expected"], want.expected) {
		return fmt.Errorf("%s result changed", id)
	}
	return nil
}

func captureResult(level, action string, code any, payload any) map[string]any {
	return map[string]any{
		"effective_level": level, "action": action, "error_code": code,
		"output": map[string]any{"capture_level": level, "payload": payload},
	}
}

func affirmativePermission(level, validFrom, expiresAt string, scope map[string]any, evidence string) map[string]any {
	return map[string]any{
		"status": "affirmative", "permitted_level": level, "valid_from": validFrom,
		"expires_at": expiresAt, "scope": scope, "evidence_reference": evidence,
	}
}

func sourceServerScope() map[string]any {
	return map[string]any{
		"account": "acct-a", "partner": "partner-a", "environment": "review",
		"purpose": "support", "store": "trace_backend",
	}
}

func sourceContractContext() map[string]any {
	return map[string]any{
		"requested_level": "full", "other_ceilings": "full",
		"capture_time_utc": "2030-01-01T00:00:00Z", "server_scope": sourceServerScope(),
	}
}

func sourceAffirmative() map[string]any {
	return affirmativePermission("full", "2029-12-31T00:00:00Z", "2030-01-02T00:00:00Z", sourceServerScope(), "fixture://contracts/affirmative-evidence")
}

func sourceDropResult() map[string]any {
	return captureResult("metadata", "drop_payload", "contract_not_authorized", nil)
}
