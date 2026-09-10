// Package clog: source-fixed Rails provenance for LAS-3488 fixture tests.
package clog

func validProvenance(provenance map[string]any) bool {
	if provenance["rails_repository"] != "LastBotInc/lastbot" || provenance["rails_commit"] != "062dfd6975512604eb3a88a3e3c3b099a4e8368b" {
		return false
	}
	fixtures := objectNoTest(provenance["fixtures"])
	if fixtures == nil || len(fixtures) != 2 {
		return false
	}
	want := map[string]map[string]any{
		"golden_corpus":    {"path": "test/fixtures/files/redaction/golden_corpus.json", "version": "2026-09-10.2", "sha256": "d9d5a426043099e78630c798f7893648c4c23b9d101becbed59bf80a71f82f6d"},
		"capture_contract": {"path": "test/fixtures/files/redaction/capture_contract.json", "version": "2026-09-10.4", "sha256": "c45694d7fb89cc1c4f0248e5e8df6584d973dbe7d6c04a632308df975f010216"},
	}
	for name, entry := range want {
		actual := objectNoTest(fixtures[name])
		if actual == nil || !equalMap(actual, entry) {
			return false
		}
	}
	return true
}
