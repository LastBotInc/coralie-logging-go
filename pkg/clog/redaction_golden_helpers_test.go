// Package clog: shared fixture assertions for LAS-3488 governance tests.
package clog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const redactionFixtureDir = "testdata/redaction"

var expectedCorpusIDs = strings.Fields(`fi.hetu.free_text fi.hetu.century_a fi.hetu.century_plus fi.hetu.lowercase sv.personnummer.separated sv.personnummer.unseparated sv.personnummer.coordination et.isikukood pl.pesel da.cpr da.cpr.ambiguous_with_personnummer fi.iban.spaced fi.iban.compact_lowercase sv.iban en.iban.de en.card.visa16 en.card.amex15 en.card.mastercard16 en.card.diners14 fi.multi_category_one_string multi.repeated_across_lines mixed.emoji_around_identifier mixed.escaped_json_string en.email.simple fi.email.plus_addressed_uppercase en.email.trailing_period sv.email.idn_punycode multi.email_repeated fi.phone.e164_spaced fi.phone.national_hyphen fi.phone.intl_double_zero sv.phone.national_spaced en.phone.with_extension voice.transcript_name_and_phone net.ipv4 net.ipv6 net.url_with_credentials_and_query_email fi.person_name sv.person_name en.person_name_apostrophe mixed.script_name_email_phone fi.street_address sv.street_address en.street_address secret.api_key secret.bearer_token secret.password_in_free_text json.tool_arguments_free_text_spillage json.tool_result_customer_block neg.idempotent_already_redacted neg.order_id_phone_shaped_en neg.order_id_phone_shaped_fi neg.ean13_sku neg.sku_alphanumeric neg.model_name neg.semver neg.four_part_version neg.invalid_ipv4_octet neg.iso_timestamp neg.uuid neg.invalid_hetu_checksum neg.invalid_hetu_century_marker neg.card_luhn_invalid neg.orgnr_separated neg.orgnr_unseparated neg.invoice_twelve_digits neg.loyalty_number neg.tracking_code neg.price_and_quantity neg.booking_reference`)

var supportedGoCaseIDs = strings.Fields(`neg.idempotent_already_redacted neg.ean13_sku neg.sku_alphanumeric neg.model_name neg.semver neg.invalid_ipv4_octet neg.iso_timestamp neg.uuid neg.invalid_hetu_checksum neg.invalid_hetu_century_marker neg.card_luhn_invalid neg.orgnr_separated neg.orgnr_unseparated neg.invoice_twelve_digits neg.tracking_code neg.price_and_quantity neg.booking_reference`)

var gapGoCaseIDs = strings.Fields(`fi.hetu.free_text fi.hetu.century_a fi.hetu.century_plus fi.hetu.lowercase sv.personnummer.separated sv.personnummer.unseparated sv.personnummer.coordination et.isikukood pl.pesel da.cpr da.cpr.ambiguous_with_personnummer fi.iban.spaced fi.iban.compact_lowercase sv.iban en.iban.de en.card.visa16 en.card.amex15 en.card.mastercard16 en.card.diners14 fi.multi_category_one_string multi.repeated_across_lines mixed.emoji_around_identifier mixed.escaped_json_string en.email.simple fi.email.plus_addressed_uppercase en.email.trailing_period sv.email.idn_punycode multi.email_repeated fi.phone.e164_spaced fi.phone.national_hyphen fi.phone.intl_double_zero sv.phone.national_spaced en.phone.with_extension voice.transcript_name_and_phone net.ipv4 net.ipv6 net.url_with_credentials_and_query_email fi.person_name sv.person_name en.person_name_apostrophe mixed.script_name_email_phone fi.street_address sv.street_address en.street_address secret.api_key secret.bearer_token secret.password_in_free_text json.tool_arguments_free_text_spillage json.tool_result_customer_block neg.order_id_phone_shaped_en neg.order_id_phone_shaped_fi neg.four_part_version neg.loyalty_number`)

func readFixture(t *testing.T, name string, target any) []byte {
	t.Helper()
	data, err := os.ReadFile(redactionFixtureDir + "/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatal(err)
	}
	return data
}

func object(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T", value)
	}
	return result
}

func array(t *testing.T, value any) []any {
	t.Helper()
	result, ok := value.([]any)
	if !ok {
		t.Fatalf("expected array, got %T", value)
	}
	return result
}

func stringValue(t *testing.T, value any) string {
	t.Helper()
	result, ok := value.(string)
	if !ok {
		t.Fatalf("expected string, got %T", value)
	}
	return result
}

func assertKeys(t *testing.T, actual map[string]any, want ...string) {
	t.Helper()
	got := make([]string, 0, len(actual))
	for key := range actual {
		got = append(got, key)
	}
	sort.Strings(got)
	want = append([]string(nil), want...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
}

func assertSet(t *testing.T, values []string, want []string) {
	t.Helper()
	got := append([]string(nil), values...)
	sort.Strings(got)
	want = append([]string(nil), want...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("set = %v, want %v", got, want)
	}
}

func stringArray(t *testing.T, values []any) []string {
	t.Helper()
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = stringValue(t, value)
	}
	return result
}

func fixtureDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func clonedJSON(t *testing.T, value any) map[string]any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var clone map[string]any
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func require(condition bool, format string, args ...any) error {
	if !condition {
		return fmt.Errorf(format, args...)
	}
	return nil
}
