// Package clog: strict LAS-3488 capture-contract fixture validation.
package clog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

var expectedContractIDs = []string{
	"shape.unknown_email_key", "shape.unknown_token_key", "shape.nested_unknown_fields", "shape.malformed_json", "shape.wrong_type", "shape.unknown_secret", "shape.metadata_stamped_raw_payload",
	"provenance.synthetic_metadata_system_ceiling", "provenance.synthetic_real_connector", "provenance.unknown_classification", "provenance.replay_cannot_raise", "provenance.billing_usage_without_payload", "provenance.external_session_id_omitted",
	"contract.unknown", "contract.denied", "contract.not_yet_valid", "contract.expired_at_boundary", "contract.scope_mismatch.account", "contract.scope_mismatch.partner", "contract.scope_mismatch.purpose", "contract.scope_mismatch.environment", "contract.scope_mismatch.store", "contract.valid_redacted", "contract.valid_full", "contract.partner_restricts_account", "contract.export_after_expiry",
	"bounds.field.at", "bounds.field.over", "bounds.record.at", "bounds.record.over", "bounds.depth.at", "bounds.depth.over", "bounds.object_members.at", "bounds.object_members.over", "bounds.array_members.at", "bounds.array_members.over",
}

func TestCaptureContractSchemaAndPolicyPins(t *testing.T) {
	var contract map[string]any
	readFixture(t, "capture_contract.json", &contract)
	assertKeys(t, contract, "contract_version", "limits", "measurement", "purpose", "schema_version", "ticket", "vectors")
	if contract["schema_version"] != float64(2) || contract["contract_version"] != "2026-09-10.4" {
		t.Fatal("capture contract version drifted")
	}
	limits := object(t, contract["limits"])
	if !equalMap(limits, map[string]any{"payload_field_utf8_bytes": float64(32768), "record_compact_json_utf8_bytes": float64(65536), "max_nesting": float64(8), "max_collection_members": float64(128)}) {
		t.Fatalf("limits changed: %#v", limits)
	}
	values := array(t, contract["vectors"])
	ids := make([]string, 0, len(values))
	for _, value := range values {
		ids = append(ids, stringValue(t, object(t, value)["id"]))
	}
	if !equalStrings(ids, expectedContractIDs) {
		t.Fatalf("vector ids changed: %v", ids)
	}
	vectors := vectorMap(t, values)
	for id, vector := range vectors {
		if err := validateCaptureVectorShape(id, vector); err != nil {
			t.Fatal(err)
		}
	}
	assertSecurityVectors(t, vectors)
	assertPermissionVectors(t, vectors)
}

func validateCaptureVectorShape(id string, vector map[string]any) error {
	if !sameKeys(vector, []string{"context", "expected", "id", "input", "owner_ticket"}) {
		return fmt.Errorf("%s has invalid keys", id)
	}
	if _, ok := vector["owner_ticket"].(string); !ok {
		return fmt.Errorf("%s owner_ticket is not a string", id)
	}
	expected := objectNoTest(vector["expected"])
	if expected == nil || !sameKeys(expected, []string{"action", "effective_level", "error_code", "output"}) {
		return fmt.Errorf("%s has invalid expected shape", id)
	}
	if _, ok := objectNoTest(expected["output"])["capture_level"]; !ok {
		return fmt.Errorf("%s omits capture level", id)
	}
	return nil
}

func TestCaptureBoundaryRecipes(t *testing.T) {
	var contract map[string]any
	readFixture(t, "capture_contract.json", &contract)
	vectors := vectorMap(t, array(t, contract["vectors"]))
	limits := object(t, contract["limits"])
	want := map[string][4]int{
		"bounds.field.at": {32768, 33101, 1, 1}, "bounds.field.over": {32769, 33102, 1, 1},
		"bounds.record.at": {21730, 65536, 1, 3}, "bounds.record.over": {21731, 65537, 1, 3},
		"bounds.depth.at": {1, 396, 8, 1}, "bounds.depth.over": {1, 405, 9, 1},
		"bounds.object_members.at": {1, 1620, 1, 128}, "bounds.object_members.over": {1, 1631, 1, 129},
		"bounds.array_members.at": {1, 834, 1, 128}, "bounds.array_members.over": {1, 838, 1, 129},
	}
	for id, measured := range want {
		vector := vectors[id]
		payload := materializePayload(t, object(t, vector["input"])["recipe"])
		encoded, err := compactJSON(completeEnvelope(payload, "redacted"))
		if err != nil {
			t.Fatal(err)
		}
		got := [4]int{maxStringBytes(payload), len(encoded), payloadDepth(payload), maxMembers(payload)}
		if got != measured {
			t.Fatalf("%s measurements = %v, want %v", id, got, measured)
		}
		assertBoundaryResult(t, id, object(t, vector["expected"]), payload, limits)
	}
}

func assertSecurityVectors(t *testing.T, vectors map[string]map[string]any) {
	t.Helper()
	for id := range sourceSecurityVectors() {
		if err := validateSourceFixedVector(id, vectors[id]); err != nil {
			t.Fatal(err)
		}
	}
}

func assertPermissionVectors(t *testing.T, vectors map[string]map[string]any) {
	t.Helper()
	for id := range sourceContractVectors() {
		if err := validateSourceFixedVector(id, vectors[id]); err != nil {
			t.Fatal(err)
		}
	}
}

func vectorMap(t *testing.T, values []any) map[string]map[string]any {
	t.Helper()
	result := make(map[string]map[string]any, len(values))
	for _, value := range values {
		vector := object(t, value)
		id := stringValue(t, vector["id"])
		if _, exists := result[id]; exists {
			t.Fatalf("duplicate vector %s", id)
		}
		result[id] = vector
	}
	return result
}

func objectNoTest(value any) map[string]any {
	result, _ := value.(map[string]any)
	return result
}

func compactJSON(value any) ([]byte, error) {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(out.Bytes(), []byte("\n")), nil
}
