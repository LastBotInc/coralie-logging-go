// Package clog: mutation guards for LAS-3488 capture-contract test fixtures.
package clog

import (
	"fmt"
	"testing"
)

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
		{"partner precedence weakened", "contract.partner_restricts_account", func(v map[string]any) { object(t, v["expected"])["effective_level"] = "full" }},
		{"synthetic ceiling raised", "provenance.synthetic_metadata_system_ceiling", func(v map[string]any) { object(t, v["expected"])["effective_level"] = "full" }},
		{"raw payload relabeled metadata", "shape.metadata_stamped_raw_payload", func(v map[string]any) { object(t, v["expected"])["action"] = "allow_metadata" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			clone := clonedJSON(t, vectors[test.id])
			test.edit(clone)
			if validatePolicyVector(test.id, clone) == nil {
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

func validatePolicyVector(id string, vector map[string]any) error {
	context, expected := objectNoTest(vector["context"]), objectNoTest(vector["expected"])
	if context == nil || expected == nil {
		return fmt.Errorf("missing context or expected")
	}
	if id == "shape.metadata_stamped_raw_payload" && expected["action"] != "reject_record" {
		return fmt.Errorf("raw payload may not be relabeled metadata")
	}
	if id == "provenance.synthetic_metadata_system_ceiling" && expected["effective_level"] != "metadata" {
		return fmt.Errorf("synthetic metadata ceiling may not be raised")
	}
	if id == "contract.partner_restricts_account" && expected["effective_level"] != "metadata" {
		return fmt.Errorf("partner restriction must win")
	}
	if permission, ok := context["contract_permission"].(map[string]any); ok && permission["status"] == "affirmative" {
		for _, key := range []string{"permitted_level", "valid_from", "expires_at", "scope", "evidence_reference"} {
			if permission[key] == nil || permission[key] == "" {
				return fmt.Errorf("affirmative permission missing %s", key)
			}
		}
	}
	if permissions, ok := context["contract_permissions"].([]any); ok {
		if len(permissions) != 2 || expected["effective_level"] != "metadata" {
			return fmt.Errorf("partner restriction is incomplete")
		}
		for _, value := range permissions {
			permission := objectNoTest(value)
			if permission == nil || permission["owner"] == nil || permission["evidence_reference"] == nil || permission["expires_at"] == nil {
				return fmt.Errorf("partner permission lacks audit fields")
			}
		}
	}
	if id == "contract.valid_full" {
		grant := objectNoTest(context["superadmin_full_grant"])
		for _, key := range []string{"audit_event_id", "basis_reference", "expires_at"} {
			if grant == nil || grant[key] == nil || grant[key] == "" {
				return fmt.Errorf("full grant missing %s", key)
			}
		}
	}
	if id == "contract.expired_at_boundary" || id == "contract.export_after_expiry" {
		if expected["effective_level"] != "metadata" || expected["action"] != "drop_payload" {
			return fmt.Errorf("expiry boundary is not half-open")
		}
	}
	return nil
}

func boundaryMeasurementMatches(id string, payload any) bool {
	if id != "bounds.field.at" {
		return false
	}
	return maxStringBytes(payload) == 32768 && payloadDepth(payload) == 1 && maxMembers(payload) == 1
}
