// Package clog: mutation guards for LAS-3488 capture-contract test fixtures.
package clog

import "testing"

func TestCaptureContractGuardsRejectSecurityRegressions(t *testing.T) {
	var contract map[string]any
	readFixture(t, "capture_contract.json", &contract)
	vectors := vectorMap(t, array(t, contract["vectors"]))
	for _, test := range []struct {
		name, id string
		edit     func(map[string]any)
	}{
		{"missing permission evidence", "contract.valid_redacted", func(v map[string]any) {
			delete(object(t, object(t, v["context"])["contract_permission"]), "evidence_reference")
		}},
		{"missing permission expiry", "contract.valid_redacted", func(v map[string]any) {
			delete(object(t, object(t, v["context"])["contract_permission"]), "expires_at")
		}},
		{"valid redacted upgraded", "contract.valid_redacted", func(v map[string]any) {
			object(t, v["expected"])["action"] = "allow_full_payload"
		}},
		{"full grant audit changed", "contract.valid_full", func(v map[string]any) {
			object(t, object(t, v["context"])["superadmin_full_grant"])["audit_event_id"] = "wrong-audit"
		}},
		{"half-open expiry relaxed", "contract.expired_at_boundary", func(v map[string]any) {
			object(t, object(t, v["context"])["contract_permission"])["expires_at"] = "2030-01-02T00:00:00Z"
		}},
		{"scope mismatch removed", "contract.scope_mismatch.account", func(v map[string]any) {
			object(t, object(t, object(t, v["context"])["contract_permission"])["scope"])["account"] = "acct-a"
		}},
		{"partner precedence weakened", "contract.partner_restricts_account", func(v map[string]any) { object(t, v["expected"])["effective_level"] = "full" }},
		{"synthetic ceiling raised", "provenance.synthetic_metadata_system_ceiling", func(v map[string]any) { object(t, v["expected"])["effective_level"] = "full" }},
		{"raw payload relabeled metadata", "shape.metadata_stamped_raw_payload", func(v map[string]any) { object(t, v["expected"])["action"] = "allow_metadata" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			clone := clonedJSON(t, vectors[test.id])
			test.edit(clone)
			if validateSourceFixedVector(test.id, clone) == nil {
				t.Fatal("mutated vector was accepted")
			}
		})
	}
}

func TestCaptureBoundaryGuardRejectsFalsifiedMeasurement(t *testing.T) {
	var contract map[string]any
	readFixture(t, "capture_contract.json", &contract)
	vector := clonedJSON(t, vectorMap(t, array(t, contract["vectors"]))["bounds.field.at"])
	part := object(t, array(t, object(t, object(t, vector["input"])["recipe"])["parts"])[0])
	part["count"] = float64(8193)
	payload := materializePayload(t, object(t, vector["input"])["recipe"])
	if boundaryMeasurementMatches("bounds.field.at", payload) {
		t.Fatal("falsified field-byte recipe was accepted")
	}
}

func TestCaptureBoundaryGuardRejectsSwappedReason(t *testing.T) {
	var contract map[string]any
	readFixture(t, "capture_contract.json", &contract)
	vector := clonedJSON(t, vectorMap(t, array(t, contract["vectors"]))["bounds.field.over"])
	object(t, vector["expected"])["error_code"] = "depth_exceeded"
	if validateBoundaryReason("bounds.field.over", object(t, vector["expected"])) == nil {
		t.Fatal("swapped boundary reason was accepted")
	}
}

func TestCaptureContractGuardRejectsNonStringOwnerTicket(t *testing.T) {
	var contract map[string]any
	readFixture(t, "capture_contract.json", &contract)
	vector := clonedJSON(t, vectorMap(t, array(t, contract["vectors"]))["bounds.field.at"])
	vector["owner_ticket"] = float64(3489)
	if validateCaptureVectorShape("bounds.field.at", vector) == nil {
		t.Fatal("non-string owner ticket was accepted")
	}
}

func TestCaptureBoundaryGuardsRejectSourceFixedDrift(t *testing.T) {
	var contract map[string]any
	readFixture(t, "capture_contract.json", &contract)
	vectors := vectorMap(t, array(t, contract["vectors"]))
	for _, test := range []struct {
		name string
		edit func(map[string]any)
	}{
		{"requested level", func(v map[string]any) { object(t, v["context"])["requested_level"] = "full" }},
		{"authorization", func(v map[string]any) { object(t, v["context"])["contract_authorized"] = false }},
		{"effective level", func(v map[string]any) { object(t, v["expected"])["effective_level"] = "full" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			vector := clonedJSON(t, vectors["bounds.field.at"])
			test.edit(vector)
			if validateSourceFixedVector("bounds.field.at", vector) == nil {
				t.Fatal("source-fixed boundary drift was accepted")
			}
		})
	}
}

func boundaryMeasurementMatches(id string, payload any) bool {
	if id != "bounds.field.at" {
		return false
	}
	return maxStringBytes(payload) == 32768 && payloadDepth(payload) == 1 && maxMembers(payload) == 1
}
