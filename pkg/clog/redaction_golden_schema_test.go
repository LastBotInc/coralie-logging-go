// Package clog: strict corpus-schema and mutation guards for LAS-3488.
package clog

import (
	"fmt"
	"testing"
)

var corpusCategories = map[string]bool{
	"finnish_hetu": true, "swedish_personnummer": true, "estonian_isikukood": true,
	"polish_pesel": true, "danish_cpr": true, "iban": true, "payment_card": true,
	"email": true, "phone": true, "ip_address": true, "person_name": true,
	"street_address": true, "secret": true,
}

var corpusCapabilities = map[string]bool{
	"regex_reference": true, "structural_only": true,
	"free_text_research": true, "secret_shape_reference": true,
}

func TestGoldenCorpusGuardsRejectSchemaMutations(t *testing.T) {
	var corpus map[string]any
	readFixture(t, "golden_corpus.json", &corpus)
	for _, mutation := range []struct {
		name string
		edit func(map[string]any)
	}{
		{"missing id", func(c map[string]any) { c["cases"] = array(t, c["cases"])[:len(expectedCorpusIDs)-1] }},
		{"extra id", func(c map[string]any) {
			cases := array(t, c["cases"])
			extra := clonedJSON(t, object(t, cases[0]))
			extra["id"] = "extra.case"
			c["cases"] = append(cases, extra)
		}},
		{"duplicate id", func(c map[string]any) { object(t, array(t, c["cases"])[1])["id"] = expectedCorpusIDs[0] }},
		{"unknown kind", func(c map[string]any) { object(t, array(t, c["cases"])[0])["kind"] = "unexpected" }},
		{"negative redaction", func(c map[string]any) {
			object(t, object(t, array(t, c["cases"])[len(expectedCorpusIDs)-1])["expected"])["redacted"] = "changed"
		}},
		{"notes number", func(c map[string]any) { object(t, array(t, c["cases"])[0])["notes"] = float64(1) }},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			clone := clonedJSON(t, corpus)
			mutation.edit(clone)
			if validateCorpusCases(array(t, clone["cases"])) == nil {
				t.Fatal("mutated corpus was accepted")
			}
		})
	}
}

func validateCorpusCases(cases []any) error {
	if len(cases) != len(expectedCorpusIDs) {
		return fmt.Errorf("case count = %d", len(cases))
	}
	seen := map[string]bool{}
	for index, value := range cases {
		kase := objectNoTest(value)
		if kase == nil {
			return fmt.Errorf("case %d is not an object", index)
		}
		id, ok := kase["id"].(string)
		if !ok || id != expectedCorpusIDs[index] || seen[id] {
			return fmt.Errorf("case id %d is invalid", index)
		}
		seen[id] = true
		kind, ok := kase["kind"].(string)
		if !ok || !contains([]string{"positive", "negative", "ambiguous"}, kind) {
			return fmt.Errorf("%s has invalid kind", id)
		}
		if err := validateCaseShape(kase, kind); err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
	}
	return nil
}

func validateCaseShape(kase map[string]any, kind string) error {
	want := []string{"capabilities", "categories", "expected", "id", "input", "kind", "locale", "notes"}
	if kind != "positive" {
		want = append(want, "near_miss_of")
	}
	if !sameKeys(kase, want) {
		return fmt.Errorf("unexpected keys")
	}
	if _, ok := kase["input"].(string); !ok {
		return fmt.Errorf("input is not a string")
	}
	if notes := kase["notes"]; notes != nil {
		if _, ok := notes.(string); !ok {
			return fmt.Errorf("notes is not a string or null")
		}
	}
	if locale, ok := kase["locale"].(string); !ok || !contains([]string{"fi", "sv", "en", "da", "et", "pl", "mixed"}, locale) {
		return fmt.Errorf("invalid locale")
	}
	categories, err := validStringSet(kase["categories"], corpusCategories, false)
	if err != nil {
		return err
	}
	if _, err := validStringSet(kase["capabilities"], corpusCapabilities, true); err != nil {
		return err
	}
	expected := objectNoTest(kase["expected"])
	if expected == nil || !sameKeys(expected, []string{"full", "metadata", "redacted"}) || expected["metadata"] != nil {
		return fmt.Errorf("invalid expected shape")
	}
	input, full, redacted := kase["input"], expected["full"], expected["redacted"]
	if _, ok := full.(string); !ok {
		return fmt.Errorf("full is not a string")
	}
	if _, ok := redacted.(string); !ok {
		return fmt.Errorf("redacted is not a string")
	}
	switch kind {
	case "positive":
		if len(categories) == 0 || input == redacted {
			return fmt.Errorf("positive invariant")
		}
	case "negative":
		near, err := validStringSet(kase["near_miss_of"], corpusCategories, true)
		if err != nil || len(categories) != 0 || len(near) == 0 || input != redacted {
			return fmt.Errorf("negative invariant")
		}
	case "ambiguous":
		if kase["near_miss_of"] != nil || !contains(categories, "phone") || input == redacted {
			return fmt.Errorf("ambiguous invariant")
		}
	}
	return nil
}

func validStringSet(value any, allowed map[string]bool, nonempty bool) ([]string, error) {
	values, ok := value.([]any)
	if !ok || (nonempty && len(values) == 0) {
		return nil, fmt.Errorf("invalid string set")
	}
	seen, result := map[string]bool{}, make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok || !allowed[text] || seen[text] {
			return nil, fmt.Errorf("invalid set member")
		}
		seen[text] = true
		result = append(result, text)
	}
	return result, nil
}

func sameKeys(value map[string]any, want []string) bool {
	actual := mapKeys(value)
	if len(actual) != len(want) {
		return false
	}
	for _, key := range want {
		if !contains(actual, key) {
			return false
		}
	}
	return true
}
