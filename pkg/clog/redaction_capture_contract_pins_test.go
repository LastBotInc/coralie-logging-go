// Package clog: source-fixed contract authorization vector expectations.
package clog

func sourceContractVectors() map[string]captureVectorExpectation {
	payload := map[string]any{"payload": map[string]any{"approved_text": "fixture"}}
	result := sourceDropResult()
	vectors := map[string]captureVectorExpectation{}
	unknown := sourceContractContext()
	unknown["contract_permission"] = nil
	vectors["contract.unknown"] = sourceVector(payload, unknown, result)
	denied := sourceContractContext()
	denied["contract_permission"] = map[string]any{"status": "denied"}
	vectors["contract.denied"] = sourceVector(payload, denied, result)
	vectors["contract.not_yet_valid"] = sourceVector(payload, sourceContextWithPermission(
		affirmativePermission("full", "2030-01-01T00:00:01Z", "2030-01-02T00:00:00Z", sourceServerScope(), "fixture://contracts/affirmative-evidence")), result)
	vectors["contract.expired_at_boundary"] = sourceVector(payload, sourceContextWithPermission(
		affirmativePermission("full", "2029-12-31T00:00:00Z", "2030-01-01T00:00:00Z", sourceServerScope(), "fixture://contracts/affirmative-evidence")), result)
	for key, value := range map[string]string{"account": "acct-b", "partner": "partner-b", "purpose": "sales", "environment": "production", "store": "journal"} {
		scope := sourceServerScope()
		scope[key] = value
		vectors["contract.scope_mismatch."+key] = sourceVector(payload, sourceContextWithPermission(
			affirmativePermission("full", "2029-12-31T00:00:00Z", "2030-01-02T00:00:00Z", scope, "fixture://contracts/affirmative-evidence")), result)
	}
	vectors["contract.valid_redacted"] = sourceVector(payload, sourceContextWithPermission(
		affirmativePermission("redacted", "2029-12-31T00:00:00Z", "2030-01-02T00:00:00Z", sourceServerScope(), "fixture://contracts/affirmative-evidence")),
		captureResult("redacted", "allow_redacted_payload", nil, map[string]any{"approved_text": "fixture"}))
	full := sourceContextWithPermission(sourceAffirmative())
	full["superadmin_full_grant"] = map[string]any{
		"audit_event_id": "audit-fixture-001", "basis_reference": "fixture://contracts/full-capture-basis", "expires_at": "2030-01-02T00:00:00Z",
	}
	vectors["contract.valid_full"] = sourceVector(payload, full,
		captureResult("full", "allow_full_payload", nil, map[string]any{"approved_text": "fixture"}))
	partner := sourceContractContext()
	accountPermission := sourceAffirmative()
	accountPermission["owner"] = "account"
	partnerPermission := affirmativePermission("metadata", "2029-12-31T00:00:00Z", "2030-01-02T00:00:00Z", sourceServerScope(), "fixture://contracts/partner-restrictive")
	partnerPermission["owner"], partnerPermission["status"] = "partner", "denied"
	accountPermission["evidence_reference"] = "fixture://contracts/account-affirmative"
	partner["contract_permissions"] = []any{accountPermission, partnerPermission}
	vectors["contract.partner_restricts_account"] = sourceVector(payload, partner, result)
	export := sourceContextWithPermission(affirmativePermission("full", "2029-12-31T00:00:00Z", "2030-01-01T00:00:00Z", sourceServerScope(), "fixture://contracts/affirmative-evidence"))
	export["capture_time_utc"], export["export_time_utc"] = "2029-12-31T23:59:59Z", "2030-01-01T00:00:00Z"
	vectors["contract.export_after_expiry"] = sourceVector(payload, export, result)
	return vectors
}

func sourceContextWithPermission(permission map[string]any) map[string]any {
	context := sourceContractContext()
	context["contract_permission"] = permission
	return context
}
