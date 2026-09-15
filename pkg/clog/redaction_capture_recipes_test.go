// Package clog: deterministic boundary recipe materialization for LAS-3488.
package clog

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func materializePayload(t *testing.T, value any) any {
	t.Helper()
	recipe := object(t, value)
	switch recipe["kind"] {
	case "field_utf8":
		var out strings.Builder
		for _, value := range array(t, recipe["parts"]) {
			part := object(t, value)
			text, _ := part["repeat"].(string)
			if text == "" {
				text = stringValue(t, part["literal"])
			}
			count := 1
			if value, ok := part["count"].(float64); ok {
				count = int(value)
			}
			out.WriteString(strings.Repeat(text, count))
		}
		return map[string]any{"field": out.String()}
	case "record_ascii":
		payload := map[string]any{}
		for index, value := range array(t, recipe["field_bytes"]) {
			payload[fmt.Sprintf("p%d", index)] = strings.Repeat("a", int(value.(float64)))
		}
		return payload
	case "nested_object":
		var value any = "x"
		for range int(recipe["containers"].(float64)) {
			value = map[string]any{"node": value}
		}
		return value
	case "object_members":
		payload := map[string]any{}
		for index := range int(recipe["members"].(float64)) {
			payload[fmt.Sprintf("k%d", index)] = "x"
		}
		return payload
	case "array_members":
		payload := make([]any, int(recipe["members"].(float64)))
		for index := range payload {
			payload[index] = "x"
		}
		return payload
	default:
		t.Fatalf("unknown recipe %v", recipe["kind"])
		return nil
	}
}

func assertBoundaryResult(t *testing.T, id string, expected map[string]any, payload any, limits map[string]any) {
	t.Helper()
	output := object(t, expected["output"])
	retained := strings.HasSuffix(id, ".at")
	if retained {
		if expected["action"] != "retain_payload" || expected["error_code"] != nil {
			t.Fatalf("%s retained result changed: %v", id, expected)
		}
		if id == "bounds.record.at" {
			if !equalMap(output, completeEnvelope("$materialized", "redacted")) {
				t.Fatalf("%s envelope stamps changed: %v", id, output)
			}
		} else if !equalMap(output, map[string]any{"capture_level": "redacted", "payload": "$materialized"}) {
			t.Fatalf("%s retained output changed: %v", id, output)
		}
		return
	}
	if strings.HasPrefix(id, "bounds.record.") {
		if expected["action"] != "drop_record_payload" || expected["error_code"] != "record_bytes_exceeded" || !equalMap(output, completeEnvelope(nil, "metadata")) {
			t.Fatalf("%s record drop changed: %v", id, expected)
		}
		return
	}
	if expected["action"] != "drop_payload_field" || validateBoundaryReason(id, expected) != nil {
		t.Fatalf("%s field drop changed: %v", id, expected)
	}
	if !equalMap(output, map[string]any{"capture_level": "redacted", "payload": map[string]any{}}) {
		t.Fatalf("%s field drop output changed: %v", id, output)
	}
	_ = payload
	_ = limits
}

func completeEnvelope(payload any, captureLevel string) map[string]any {
	return map[string]any{
		"capture_level": captureLevel, "policy_version": "capture-policy-2026-09-15.1", "schema_version": "capture-contract-v2",
		"traffic_classification": "synthetic", "correlation_ids": map[string]any{"trace_id": "trace-fixture-opaque", "turn_id": "turn-fixture-opaque"},
		"captured_at": "2030-01-01T00:00:00Z", "expires_at": "2030-01-01T01:00:00Z", "payload": payload,
	}
}

func maxStringBytes(value any) int {
	switch value := value.(type) {
	case string:
		return len(value)
	case map[string]any:
		maximum := 0
		for _, child := range value {
			maximum = max(maximum, maxStringBytes(child))
		}
		return maximum
	case []any:
		maximum := 0
		for _, child := range value {
			maximum = max(maximum, maxStringBytes(child))
		}
		return maximum
	default:
		return 0
	}
}

func payloadDepth(value any) int {
	switch value := value.(type) {
	case map[string]any:
		depth := 0
		for _, child := range value {
			depth = max(depth, payloadDepth(child))
		}
		return 1 + depth
	case []any:
		depth := 0
		for _, child := range value {
			depth = max(depth, payloadDepth(child))
		}
		return 1 + depth
	default:
		return 0
	}
}

func maxMembers(value any) int {
	switch value := value.(type) {
	case map[string]any:
		maximum := len(value)
		for _, child := range value {
			maximum = max(maximum, maxMembers(child))
		}
		return maximum
	case []any:
		maximum := len(value)
		for _, child := range value {
			maximum = max(maximum, maxMembers(child))
		}
		return maximum
	default:
		return 0
	}
}

func equalMap(left, right map[string]any) bool {
	leftKeys, rightKeys := mapKeys(left), mapKeys(right)
	if !equalStrings(leftKeys, rightKeys) {
		return false
	}
	for _, key := range leftKeys {
		if fmt.Sprint(left[key]) != fmt.Sprint(right[key]) {
			return false
		}
	}
	return true
}

func mapKeys(values map[string]any) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
