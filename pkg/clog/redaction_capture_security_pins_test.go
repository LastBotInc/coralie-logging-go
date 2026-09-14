// Package clog: source-fixed shape and provenance vector expectations.
package clog

func sourceSecurityVectors() map[string]captureVectorExpectation {
	redacted := map[string]any{"requested_level": "redacted", "contract_authorized": true}
	return map[string]captureVectorExpectation{
		"shape.unknown_email_key": sourceVector(map[string]any{"payload": map[string]any{"email": "fixture@example.com"}}, redacted,
			captureResult("redacted", "drop_unknown_field", "capture_unknown_schema_field", map[string]any{})),
		"shape.unknown_token_key": sourceVector(map[string]any{"payload": map[string]any{"token": "not-a-credential"}}, redacted, // #nosec G101 -- Synthetic sentinel for unknown-field rejection, not a credential.
			captureResult("redacted", "drop_unknown_field", "capture_unknown_schema_field", map[string]any{})),
		"shape.nested_unknown_fields": sourceVector(map[string]any{"payload": map[string]any{
			"known": map[string]any{"unknown": map[string]any{"email": "fixture@example.com"}}, "items": []any{map[string]any{"unknown": "value"}},
		}}, redacted, captureResult("redacted", "drop_unknown_field", "capture_unknown_schema_field", map[string]any{})),
		"shape.malformed_json": sourceVector(map[string]any{"payload": "{not-json"}, redacted,
			captureResult("redacted", "drop_payload_field", "invalid_structure", map[string]any{})),
		"shape.wrong_type": sourceVector(map[string]any{"payload": map[string]any{"approved_count": "not-an-integer"}}, redacted,
			captureResult("redacted", "drop_payload_field", "invalid_structure", map[string]any{})),
		"shape.unknown_secret": sourceVector(map[string]any{"payload": map[string]any{"approved_text": "secret=not-a-real-secret"}},
			map[string]any{"requested_level": "full", "contract_authorized": true},
			captureResult("full", "drop_payload_field", "unsafe_shape", map[string]any{})),
		"shape.metadata_stamped_raw_payload": sourceVector(map[string]any{"envelope": map[string]any{
			"capture_level": "metadata", "payload": map[string]any{"text": "raw"},
		}}, map[string]any{"requested_level": "metadata", "contract_authorized": true},
			captureResult("metadata", "reject_record", "capture_envelope_level_mismatch", nil)),
		"provenance.synthetic_metadata_system_ceiling": sourceVector(map[string]any{"payload": map[string]any{"approved_text": "synthetic"}},
			map[string]any{"synthetic_provenance": "verified", "requested_level": "full", "system_ceiling": "metadata", "contract_authorized": true},
			captureResult("metadata", "drop_payload", "capture_ceiling_applied", nil)),
		"provenance.synthetic_real_connector": sourceVector(map[string]any{"payload": map[string]any{"approved_text": "synthetic request"}},
			map[string]any{"synthetic_provenance": "real_connector", "requested_level": "full", "system_ceiling": "full", "contract_authorized": true},
			captureResult("metadata", "drop_payload", "capture_untrusted_synthetic_provenance", nil)),
		"provenance.unknown_classification": sourceVector(map[string]any{"payload": map[string]any{"approved_text": "unknown traffic"}},
			map[string]any{"traffic_classification": "unknown", "requested_level": "full", "contract_authorized": true},
			captureResult("metadata", "drop_payload", "capture_unknown_traffic_classification", nil)),
		"provenance.replay_cannot_raise": sourceVector(map[string]any{"payload": map[string]any{"approved_text": "replay"}},
			map[string]any{"source_capture_level": "redacted", "requested_level": "full", "system_ceiling": "full", "contract_authorized": true},
			captureResult("redacted", "allow_redacted_payload", nil, map[string]any{"approved_text": "replay"})),
		"provenance.billing_usage_without_payload": sourceVector(map[string]any{"usage": map[string]any{"input_tokens": float64(12), "output_tokens": float64(4), "provider": "openai"}},
			map[string]any{"requested_level": "metadata", "contract_authorized": true},
			captureResult("metadata", "allow_metadata", nil, nil)),
		"provenance.external_session_id_omitted": sourceVector(map[string]any{"payload": map[string]any{"session_id": "external-session"}},
			redacted, captureResult("redacted", "drop_unknown_field", "capture_external_identifier_omitted", map[string]any{})),
	}
}

func sourceVector(input, context, expected map[string]any) captureVectorExpectation {
	return captureVectorExpectation{input: input, context: context, expected: expected}
}
