// Package clog: exact Rails-source semantic pins for capture-contract vectors.
package clog

import "testing"

func TestCaptureContractMatchesSourceFixedVectors(t *testing.T) {
	var contract map[string]any
	readFixture(t, "capture_contract.json", &contract)
	vectors := vectorMap(t, array(t, contract["vectors"]))
	want := sourceSecurityVectors()
	for id, vector := range sourceContractVectors() {
		want[id] = vector
	}
	if len(want) != 26 {
		t.Fatalf("source-fixed vector count = %d", len(want))
	}
	for id := range want {
		if err := validateSourceFixedVector(id, vectors[id]); err != nil {
			t.Fatal(err)
		}
	}
}
