// Package clog: governance checks for the LAS-3488 shared redaction corpus.
package clog

import (
	"reflect"
	"testing"
)

func TestRedactionFixtureProvenance(t *testing.T) {
	var provenance map[string]any
	readFixture(t, "provenance.json", &provenance)
	assertKeys(t, provenance, "fixtures", "rails_commit", "rails_repository")
	if !validProvenance(provenance) {
		t.Fatal("provenance is not source-fixed")
	}
	if got := provenance["rails_repository"]; got != "LastBotInc/lastbot" {
		t.Fatalf("rails_repository = %v", got)
	}
	if got := provenance["rails_commit"]; got != "062dfd6975512604eb3a88a3e3c3b099a4e8368b" {
		t.Fatalf("rails_commit = %v", got)
	}
	fixtures := object(t, provenance["fixtures"])
	assertKeys(t, fixtures, "capture_contract", "golden_corpus")
	for _, fixture := range []struct {
		name, file, path, version, digest string
	}{
		{"golden_corpus", "golden_corpus.json", "test/fixtures/files/redaction/golden_corpus.json", "2026-09-10.2", "d9d5a426043099e78630c798f7893648c4c23b9d101becbed59bf80a71f82f6d"},
		{"capture_contract", "capture_contract.json", "test/fixtures/files/redaction/capture_contract.json", "2026-09-10.4", "c45694d7fb89cc1c4f0248e5e8df6584d973dbe7d6c04a632308df975f010216"},
	} {
		entry := object(t, fixtures[fixture.name])
		assertKeys(t, entry, "path", "sha256", "version")
		if entry["path"] != fixture.path || entry["version"] != fixture.version || entry["sha256"] != fixture.digest {
			t.Fatalf("%s provenance = %v", fixture.name, entry)
		}
		var ignored any
		if got := fixtureDigest(readFixture(t, fixture.file, &ignored)); got != fixture.digest {
			t.Fatalf("%s digest = %s, want %s", fixture.file, got, fixture.digest)
		}
	}
}

func TestGoldenCorpusSchemaAndClassification(t *testing.T) {
	var corpus map[string]any
	readFixture(t, "golden_corpus.json", &corpus)
	assertKeys(t, corpus, "cases", "corpus_version", "description", "governance_doc", "placeholders", "schema_version", "scope", "ticket")
	if corpus["schema_version"] != float64(2) || corpus["corpus_version"] != "2026-09-10.2" {
		t.Fatalf("unexpected corpus version: %v / %v", corpus["schema_version"], corpus["corpus_version"])
	}
	cases := array(t, corpus["cases"])
	if err := validateCorpusCases(cases); err != nil {
		t.Fatal(err)
	}
	ids, categories, locales, capabilities := make([]string, 0, len(cases)), []string{}, []string{}, []string{}
	for _, value := range cases {
		kase := object(t, value)
		id, kind := stringValue(t, kase["id"]), stringValue(t, kase["kind"])
		if kind == "positive" {
			assertKeys(t, kase, "capabilities", "categories", "expected", "id", "input", "kind", "locale", "notes")
		} else {
			assertKeys(t, kase, "capabilities", "categories", "expected", "id", "input", "kind", "locale", "near_miss_of", "notes")
		}
		ids = append(ids, id)
		locales = append(locales, stringValue(t, kase["locale"]))
		categories = append(categories, stringArray(t, array(t, kase["categories"]))...)
		capabilities = append(capabilities, stringArray(t, array(t, kase["capabilities"]))...)
		expected := object(t, kase["expected"])
		assertKeys(t, expected, "full", "metadata", "redacted")
		if expected["metadata"] != nil {
			t.Fatalf("%s metadata must be null", id)
		}
		if kind == "negative" && (len(array(t, kase["categories"])) != 0 || expected["redacted"] != kase["input"]) {
			t.Fatalf("%s negative case was weakened", id)
		}
	}
	if !reflect.DeepEqual(ids, expectedCorpusIDs) {
		t.Fatalf("corpus ids changed: %v", ids)
	}
	assertSet(t, unique(categories), []string{"danish_cpr", "email", "estonian_isikukood", "finnish_hetu", "iban", "ip_address", "payment_card", "person_name", "phone", "polish_pesel", "secret", "street_address", "swedish_personnummer"})
	assertSet(t, unique(locales), []string{"da", "en", "et", "fi", "mixed", "pl", "sv"})
	assertSet(t, unique(capabilities), []string{"free_text_research", "regex_reference", "secret_shape_reference", "structural_only"})
	assertSecretFullOutputs(t, cases)
}

func TestNativeRedactorGoldenBaseline(t *testing.T) {
	var corpus, baseline map[string]any
	readFixture(t, "golden_corpus.json", &corpus)
	readFixture(t, "go_baseline.json", &baseline)
	assertKeys(t, baseline, "corpus_version", "outputs")
	if baseline["corpus_version"] != corpus["corpus_version"] {
		t.Fatal("baseline corpus version drifted")
	}
	outputs := outputByID(t, array(t, baseline["outputs"]))
	if err := validatePartition(outputs, supportedGoCaseIDs, gapGoCaseIDs); err != nil {
		t.Fatal(err)
	}
	redactor := NewDefaultRedactor()
	for _, value := range array(t, corpus["cases"]) {
		kase := object(t, value)
		id, input := stringValue(t, kase["id"]), stringValue(t, kase["input"])
		actual, pinned := redactor.Redact(input), outputs[id]
		if actual != pinned {
			t.Fatalf("%s actual output drifted: %q != %q", id, actual, pinned)
		}
		expected := stringValue(t, object(t, kase["expected"])["redacted"])
		if contains(supportedGoCaseIDs, id) && actual != expected {
			t.Fatalf("%s supported output = %q, want %q", id, actual, expected)
		}
		if contains(gapGoCaseIDs, id) && actual == expected {
			t.Fatalf("%s gap unexpectedly conforms", id)
		}
		if redactor.Redact(actual) != actual {
			t.Fatalf("%s native output is not idempotent", id)
		}
	}
}

func assertSecretFullOutputs(t *testing.T, cases []any) {
	t.Helper()
	if err := validateSecretFull(cases); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{ // #nosec G101 -- Expected redaction placeholders, not credentials.
		"secret.api_key": "The integration key is [SECRET]", "secret.bearer_token": "Authorization: Bearer [SECRET]",
		"secret.password_in_free_text": "Salasanani on [SECRET] jos se auttaa", "net.url_with_credentials_and_query_email": "Callback https://[SECRET]@api.example.com/v1?email=fixture@example.com failed",
	}
	got := map[string]string{}
	capabilityIDs := []string{}
	for _, value := range cases {
		kase, id := object(t, value), stringValue(t, object(t, value)["id"])
		if contains(stringArray(t, array(t, kase["categories"])), "secret") {
			got[id] = stringValue(t, object(t, kase["expected"])["full"])
		}
		if contains(stringArray(t, array(t, kase["capabilities"])), "secret_shape_reference") {
			capabilityIDs = append(capabilityIDs, id)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("secret full outputs = %#v", got)
	}
	assertSet(t, capabilityIDs, []string{"net.url_with_credentials_and_query_email", "secret.api_key", "secret.bearer_token", "secret.password_in_free_text"})
}

func outputByID(t *testing.T, entries []any) map[string]string {
	t.Helper()
	outputs := make(map[string]string, len(entries))
	for _, entry := range entries {
		value := object(t, entry)
		assertKeys(t, value, "id", "output")
		id := stringValue(t, value["id"])
		if _, exists := outputs[id]; exists {
			t.Fatalf("duplicate baseline output %s", id)
		}
		outputs[id] = stringValue(t, value["output"])
	}
	return outputs
}

func validatePartition(outputs map[string]string, supported, gaps []string) error {
	if len(supported) != len(unique(supported)) || len(gaps) != len(unique(gaps)) {
		return require(false, "partition has duplicate members")
	}
	if len(outputs) != len(expectedCorpusIDs) || len(supported)+len(gaps) != len(expectedCorpusIDs) {
		return require(false, "partition is not exhaustive")
	}
	for _, id := range expectedCorpusIDs {
		if _, ok := outputs[id]; !ok || contains(supported, id) == contains(gaps, id) {
			return require(false, "invalid partition member %s", id)
		}
	}
	for id := range outputs {
		if !contains(expectedCorpusIDs, id) {
			return require(false, "extra baseline member %s", id)
		}
	}
	return nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func unique(values []string) []string {
	seen, result := map[string]bool{}, []string{}
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
