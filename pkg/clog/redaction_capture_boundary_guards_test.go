// Package clog: exact boundary rejection reasons for LAS-3488 fixtures.
package clog

import "fmt"

func validateBoundaryReason(id string, expected map[string]any) error {
	want := map[string]string{
		"bounds.field.over":          "field_bytes_exceeded",
		"bounds.depth.over":          "depth_exceeded",
		"bounds.object_members.over": "members_exceeded",
		"bounds.array_members.over":  "members_exceeded",
	}
	if reason, ok := want[id]; ok && expected["error_code"] != reason {
		return fmt.Errorf("%s error_code = %v, want %s", id, expected["error_code"], reason)
	}
	return nil
}
