// Package clog: mutation guards for LAS-3488 native redactor baseline tests.
package clog

import (
	"fmt"
	"strings"
	"testing"
)

func TestGoldenGuardsRejectWeakenedPartitions(t *testing.T) {
	outputs := map[string]string{}
	for _, id := range expectedCorpusIDs {
		outputs[id] = id
	}
	for _, mutation := range []struct {
		name string
		ok   map[string]string
		sup  []string
		gap  []string
	}{
		{"deleted partition member", outputs, expectedCorpusIDs[:69], []string{}},
		{"duplicate partition member", outputs, append(append([]string{}, expectedCorpusIDs...), "fi.hetu.free_text"), []string{}},
		{"extra baseline", map[string]string{"extra": "x"}, expectedCorpusIDs, []string{}},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			if validatePartition(mutation.ok, mutation.sup, mutation.gap) == nil {
				t.Fatal("mutation was accepted")
			}
		})
	}
}

func TestGoldenGuardsRejectWrongProvenancePath(t *testing.T) {
	var provenance map[string]any
	readFixture(t, "provenance.json", &provenance)
	clone := clonedJSON(t, provenance)
	object(t, object(t, clone["fixtures"])["golden_corpus"])["path"] = "test/fixtures/files/redaction/not-the-source.json"
	if validProvenance(clone) {
		t.Fatal("wrong Rails provenance path was accepted")
	}
}

func TestGoldenGuardsRejectSecretFullInput(t *testing.T) {
	var corpus map[string]any
	readFixture(t, "golden_corpus.json", &corpus)
	for index, value := range array(t, corpus["cases"]) {
		kase := object(t, value)
		if !contains(stringArray(t, array(t, kase["categories"])), "secret") {
			continue
		}
		clone := clonedJSON(t, corpus)
		cloneCase := object(t, array(t, clone["cases"])[index])
		object(t, cloneCase["expected"])["full"] = cloneCase["input"]
		if validateSecretFull(array(t, clone["cases"])) == nil {
			t.Fatalf("%s secret full=input was accepted", kase["id"])
		}
	}
}

func TestGoldenGuardsRejectSecretSpecificRegressions(t *testing.T) {
	var corpus map[string]any
	readFixture(t, "golden_corpus.json", &corpus)
	for _, test := range []struct {
		name, want string
		edit       func(map[string]any)
	}{
		{"mixed URL/email full redacts ordinary email", "secret full output changed", func(c map[string]any) {
			for _, value := range array(t, c["cases"]) {
				caseValue := object(t, value)
				if caseValue["id"] == "net.url_with_credentials_and_query_email" {
					object(t, caseValue["expected"])["full"] = "Callback https://[SECRET]@api.example.com/v1?email=[EMAIL] failed"
				}
			}
		}},
		{"secret category removed", "secret classification changed", func(c map[string]any) {
			for _, value := range array(t, c["cases"]) {
				caseValue := object(t, value)
				if caseValue["id"] == "secret.api_key" {
					caseValue["categories"] = []any{"email"}
				}
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			clone := clonedJSON(t, corpus)
			test.edit(clone)
			err := validateSecretFull(array(t, clone["cases"]))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("secret regression error = %v, want %q", err, test.want)
			}
		})
	}
}

func validateSecretFull(cases []any) error {
	want := map[string]string{ // #nosec G101 -- Expected redaction placeholders, not credentials.
		"secret.api_key": "The integration key is [SECRET]", "secret.bearer_token": "Authorization: Bearer [SECRET]",
		"secret.password_in_free_text": "Salasanani on [SECRET] jos se auttaa", "net.url_with_credentials_and_query_email": "Callback https://[SECRET]@api.example.com/v1?email=fixture@example.com failed",
	}
	seen := map[string]bool{}
	for _, value := range cases {
		kase := objectNoTest(value)
		if kase == nil {
			return fmt.Errorf("case is not an object")
		}
		id, input := kase["id"].(string), kase["input"].(string)
		full := objectNoTest(kase["expected"])["full"]
		wantFull, isSecret := want[id]
		if isSecret {
			if !hasSecretCategory(kase["categories"]) {
				return fmt.Errorf("%s secret classification changed", id)
			}
			if wantFull != full || input == full {
				return fmt.Errorf("%s secret full output changed", id)
			}
			seen[id] = true
		} else if hasSecretCategory(kase["categories"]) {
			return fmt.Errorf("%s secret classification changed", id)
		} else if input != full {
			return fmt.Errorf("%s non-secret full output changed", id)
		}
	}
	if len(seen) != len(want) {
		return fmt.Errorf("secret classification changed")
	}
	return nil
}

func hasSecretCategory(value any) bool {
	values, ok := value.([]any)
	if !ok {
		return false
	}
	for _, value := range values {
		if value == "secret" {
			return true
		}
	}
	return false
}
