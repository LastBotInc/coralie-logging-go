// Package clog: source-fixed Rails provenance for LAS-3488 fixture tests.
package clog

func validProvenance(provenance map[string]any) bool {
	if provenance["rails_repository"] != "LastBotInc/lastbot" || provenance["rails_commit"] != "adddf3495beb3564aeef9adb0890013ff37eb7f6" {
		return false
	}
	fixtures := objectNoTest(provenance["fixtures"])
	if fixtures == nil || len(fixtures) != 2 {
		return false
	}
	want := map[string]map[string]any{
		"golden_corpus":    {"path": "test/fixtures/files/redaction/golden_corpus.json", "version": "2026-09-15.1", "sha256": "397452b288dea786780b1ab092bb39918076afdd90dac6d8bce48bf9e7d434c8"},
		"capture_contract": {"path": "test/fixtures/files/redaction/capture_contract.json", "version": "2026-09-15.1", "sha256": "433fd2a1fa5ea8d080ff684f00650a8e3f8e0ab53b0b1fce2d12e58f2db8b2d4"},
	}
	for name, entry := range want {
		actual := objectNoTest(fixtures[name])
		if actual == nil || !equalMap(actual, entry) {
			return false
		}
	}
	return true
}
