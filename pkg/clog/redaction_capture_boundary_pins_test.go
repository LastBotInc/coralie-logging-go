// Package clog: source-fixed boundary vectors for LAS-3488 capture contracts.
package clog

func sourceBoundaryVectors() map[string]captureVectorExpectation {
	context := map[string]any{"requested_level": "redacted", "contract_authorized": true}
	fieldAt := map[string]any{"kind": "field_utf8", "parts": []any{map[string]any{"repeat": "😀", "count": float64(8192)}}}
	fieldOver := map[string]any{"kind": "field_utf8", "parts": []any{map[string]any{"repeat": "😀", "count": float64(8192)}, map[string]any{"literal": "x"}}}
	return map[string]captureVectorExpectation{
		"bounds.field.at":            sourceBoundaryVector(fieldAt, context, "redacted", "retain_payload", nil, map[string]any{"capture_level": "redacted", "payload": "$materialized"}),
		"bounds.field.over":          sourceBoundaryVector(fieldOver, context, "redacted", "drop_payload_field", "field_bytes_exceeded", map[string]any{"capture_level": "redacted", "payload": map[string]any{}}),
		"bounds.record.at":           sourceBoundaryVector(map[string]any{"kind": "record_ascii", "field_bytes": []any{float64(21730), float64(21730), float64(21730)}}, context, "redacted", "retain_payload", nil, completeEnvelope("$materialized", "redacted")),
		"bounds.record.over":         sourceBoundaryVector(map[string]any{"kind": "record_ascii", "field_bytes": []any{float64(21731), float64(21730), float64(21730)}}, context, "metadata", "drop_record_payload", "record_bytes_exceeded", completeEnvelope(nil, "metadata")),
		"bounds.depth.at":            sourceBoundaryVector(map[string]any{"kind": "nested_object", "containers": float64(8)}, context, "redacted", "retain_payload", nil, map[string]any{"capture_level": "redacted", "payload": "$materialized"}),
		"bounds.depth.over":          sourceBoundaryVector(map[string]any{"kind": "nested_object", "containers": float64(9)}, context, "redacted", "drop_payload_field", "depth_exceeded", map[string]any{"capture_level": "redacted", "payload": map[string]any{}}),
		"bounds.object_members.at":   sourceBoundaryVector(map[string]any{"kind": "object_members", "members": float64(128)}, context, "redacted", "retain_payload", nil, map[string]any{"capture_level": "redacted", "payload": "$materialized"}),
		"bounds.object_members.over": sourceBoundaryVector(map[string]any{"kind": "object_members", "members": float64(129)}, context, "redacted", "drop_payload_field", "members_exceeded", map[string]any{"capture_level": "redacted", "payload": map[string]any{}}),
		"bounds.array_members.at":    sourceBoundaryVector(map[string]any{"kind": "array_members", "members": float64(128)}, context, "redacted", "retain_payload", nil, map[string]any{"capture_level": "redacted", "payload": "$materialized"}),
		"bounds.array_members.over":  sourceBoundaryVector(map[string]any{"kind": "array_members", "members": float64(129)}, context, "redacted", "drop_payload_field", "members_exceeded", map[string]any{"capture_level": "redacted", "payload": map[string]any{}}),
	}
}

func sourceBoundaryVector(recipe, context map[string]any, level, action string, code any, output map[string]any) captureVectorExpectation {
	return sourceVector(map[string]any{"recipe": recipe}, context, map[string]any{
		"effective_level": level, "action": action, "error_code": code, "output": output,
	})
}
