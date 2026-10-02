package shortclaimdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

// Exact public development bytes, not private prompts or held-out examples.
// Labels are the existing Root-qualified finite annotations, not reader labels.
const publicRows = `{"schema":"riido-finite-development-request-row-v1","stable_id":"next60-pflag-native-ipnet","request":"For frozen IPv4 strings, trim CIDRs into masked networks. Preserve the previous network on errors and reject bare addresses.","candidates":[{"metadata_id":"ipNetValue.Set","text":"On frozen inputs, trim CIDRs into masked networks; preserve the previous network on errors and reject bare addresses.","label":true,"sample_weight":1},{"metadata_id":"ipValue.Set","text":"On frozen inputs, parse trimmed plain IP text; preserve the previous IP on errors, reject CIDRs, and accept bare addresses as IP values.","label":false,"sample_weight":1},{"metadata_id":"ipMaskValue.Set","text":"On frozen inputs, parse dotted IPv4 text into mask bytes; preserve the previous mask on errors, reject CIDRs, and accept bare addresses as masks.","label":false,"sample_weight":1}],"role":"development_train","whole_group":77,"source_family":"upstream53-cobra-pflag-connected","source_revision":"c966cfef47379dcb01e7929504d66d94b540945b","finite_scope":{"input_count":5,"candidate_observations":15,"description":"Only the frozen five IPv4 inputs and four integer-hook configurations; no all-input or unseen-source guarantee.","all_input_guarantee":false,"unseen_source_guarantee":false},"text_revision":"native2-finite-v3","feature_policy":"request_and_candidates_text_only"}
{"schema":"riido-finite-development-request-row-v1","stable_id":"next60-mapstructure-native-or","request":"For frozen integer hooks, retry unchanged input until nil error, including nil output. Concatenate each failed message followed by a line feed; zero hooks must return an error.","candidates":[{"metadata_id":"OrComposeDecodeHookFunc","text":"Try frozen hooks on unchanged input; stop on nil error, including nil output. If none succeed, return an error concatenating each message followed by a line feed.","label":true,"sample_weight":1},{"metadata_id":"ComposeDecodeHookFunc","text":"Chain frozen hooks, passing each output onward; stop on the first error. With no hooks, return the original input without error.","label":false,"sample_weight":1}],"role":"development_train","whole_group":78,"source_family":"upstream53-mapstructure","source_revision":"52aa5c6dc1d27226460807054ca2107b2d54fb2d","finite_scope":{"input_count":4,"candidate_observations":8,"description":"Only the frozen five IPv4 inputs and four integer-hook configurations; no all-input or unseen-source guarantee.","all_input_guarantee":false,"unseen_source_guarantee":false},"text_revision":"native2-finite-v3","feature_policy":"request_and_candidates_text_only"}
`

// Exact public INPUTS.v3 bytes provide an independently specified projection
// target. Generic reader provenance is deliberately substituted after parsing.
const publicInputs = `[
  {
    "schema": "riido-short-behavior-claim-v1",
    "request": "For frozen IPv4 strings, trim CIDRs into masked networks. Preserve the previous network on errors and reject bare addresses.",
    "candidates": [
      {"id":"ipNetValue.Set","text":"On frozen inputs, trim CIDRs into masked networks; preserve the previous network on errors and reject bare addresses."},
      {"id":"ipValue.Set","text":"On frozen inputs, parse trimmed plain IP text; preserve the previous IP on errors, reject CIDRs, and accept bare addresses as IP values."},
      {"id":"ipMaskValue.Set","text":"On frozen inputs, parse dotted IPv4 text into mask bytes; preserve the previous mask on errors, reject CIDRs, and accept bare addresses as masks."}
    ],
    "provenance":"native2-finite-v3"
  },
  {
    "schema":"riido-short-behavior-claim-v1",
    "request":"For frozen integer hooks, retry unchanged input until nil error, including nil output. Concatenate each failed message followed by a line feed; zero hooks must return an error.",
    "candidates":[
      {"id":"OrComposeDecodeHookFunc","text":"Try frozen hooks on unchanged input; stop on nil error, including nil output. If none succeed, return an error concatenating each message followed by a line feed."},
      {"id":"ComposeDecodeHookFunc","text":"Chain frozen hooks, passing each output onward; stop on the first error. With no hooks, return the original input without error."}
    ],
    "provenance":"native2-finite-v3"
  }
]
`

func firstRow() string { return strings.Split(publicRows, "\n")[0] }
func mustReplace(t *testing.T, raw, old, replacement string) string {
	t.Helper()
	if !strings.Contains(raw, old) {
		t.Fatal("test mutation target absent")
	}
	return strings.Replace(raw, old, replacement, 1)
}
func digest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }

func TestExactPublicRowsProjectToExistingInputs(t *testing.T) {
	if len(publicRows) != 2597 || digest(publicRows) != "843e4776823fe44ad14f0c7a8b6bdf29827d92596ef3d8bf1dd31d498e87ca1c" {
		t.Fatal("public row fixture pin")
	}
	if len(publicInputs) != 1483 || digest(publicInputs) != "63e195e3d103a5725d61c298c57159fc77f216bd253879926fc4309c5415abfb" {
		t.Fatal("public input fixture pin")
	}
	var expected []shortclaim.Input
	if err := json.Unmarshal([]byte(publicInputs), &expected); err != nil || len(expected) != 2 {
		t.Fatal("reference fixture")
	}
	rows := strings.Split(strings.TrimSuffix(publicRows, "\n"), "\n")
	wantCounts := [2]int{3, 2}
	wantGroups := [2]int{77, 78}
	wantInputs := [2]int{5, 4}
	wantObservations := [2]int{15, 8}
	wantIDs := [2]string{"next60-pflag-native-ipnet", "next60-mapstructure-native-or"}
	wantFamilies := [2]string{"upstream53-cobra-pflag-connected", "upstream53-mapstructure"}
	wantRevisions := [2]string{"c966cfef47379dcb01e7929504d66d94b540945b", "52aa5c6dc1d27226460807054ca2107b2d54fb2d"}
	for i, row := range rows {
		e, err := LoadDevelopmentRow(strings.NewReader(row))
		if err != nil || !e.Valid() {
			t.Fatalf("row %d failed: %v", i, err)
		}
		expected[i].Provenance = InputProvenance
		v, err := shortclaim.ValidateInput(expected[i])
		if err != nil || e.Input().Prepared() != v.Prepared() {
			t.Fatalf("row %d request/candidate projection differs", i)
		}
		m := e.Metadata()
		if m.StableID != wantIDs[i] || m.SourceFamily != wantFamilies[i] || m.SourceRevision != wantRevisions[i] || m.WholeGroup != wantGroups[i] || m.Role != Role || m.FeaturePolicy != FeaturePolicy || m.TextRevision != "native2-finite-v3" {
			t.Fatalf("row %d metadata mapping", i)
		}
		if m.FiniteScope.InputCount != wantInputs[i] || m.FiniteScope.CandidateObservations != wantObservations[i] || m.FiniteScope.AllInputGuarantee || m.FiniteScope.UnseenSourceGuarantee {
			t.Fatalf("row %d scope mapping", i)
		}
		s := e.Supervision()
		if s.Count != wantCounts[i] {
			t.Fatal("candidate order/count")
		}
		for j := 0; j < shortclaim.MaxCandidates; j++ {
			if s.Labels[j] != (j == 0) {
				t.Fatalf("row %d label %d", i, j)
			}
			weight := 0
			if j < s.Count {
				weight = 1
			}
			if s.Weights[j] != weight {
				t.Fatalf("row %d weight %d", i, j)
			}
		}
	}
}

func TestMetadataAndSupervisionNeverChangeNormalizedText(t *testing.T) {
	base, err := LoadDevelopmentRow(strings.NewReader(firstRow()))
	if err != nil {
		t.Fatal(err)
	}
	mutated := firstRow()
	for _, change := range [][2]string{
		{"next60-pflag-native-ipnet", "different-development-parent"},
		{"upstream53-cobra-pflag-connected", "different-source-family"},
		{"c966cfef47379dcb01e7929504d66d94b540945b", "0123456789abcdef0123456789abcdef01234567"},
		{"native2-finite-v3", "text-revision-4"},
		{"\"whole_group\":77", "\"whole_group\":99"},
		{"\"input_count\":5", "\"input_count\":6"},
		{"\"candidate_observations\":15", "\"candidate_observations\":18"},
		{"Only the frozen five IPv4 inputs and four integer-hook configurations; no all-input or unseen-source guarantee.", "Different finite bookkeeping description."},
		{"ipNetValue.Set", "alternativeMetadataID"},
		{"\"label\":true", "\"label\":false"},
	} {
		mutated = mustReplace(t, mutated, change[0], change[1])
	}
	other, err := LoadDevelopmentRow(strings.NewReader(mutated))
	if err != nil {
		t.Fatal(err)
	}
	a, b := base.Input().Prepared(), other.Input().Prepared()
	if a.NormalizedRequest != b.NormalizedRequest || a.Request != b.Request || a.Count != b.Count || a.Provenance != InputProvenance || b.Provenance != InputProvenance {
		t.Fatal("metadata contaminated request projection")
	}
	for i := 0; i < a.Count; i++ {
		if a.Candidates[i].Text != b.Candidates[i].Text || a.Candidates[i].NormalizedText != b.Candidates[i].NormalizedText {
			t.Fatal("metadata contaminated candidate projection")
		}
	}
	if a.Candidates[0].ID == b.Candidates[0].ID || base.Metadata() == other.Metadata() || base.Supervision() == other.Supervision() {
		t.Fatal("metadata mutation not exercised")
	}
}

func TestAccessorsReturnIndependentFixedArrayCopies(t *testing.T) {
	e, err := LoadDevelopmentRow(strings.NewReader(firstRow()))
	if err != nil {
		t.Fatal(err)
	}
	before := e
	s := e.Supervision()
	s.Labels[0] = false
	s.Weights[0] = 99
	s.Count = 0
	m := e.Metadata()
	m.StableID = "mutated"
	m.FiniteScope.InputCount = 999
	p := e.Input().Prepared()
	p.Candidates[0].NormalizedText = "mutated"
	p.Count = 0
	if e != before || e.Supervision().Labels[0] != true || e.Metadata().StableID != "next60-pflag-native-ipnet" || e.Input().Prepared().Count != 3 {
		t.Fatal("accessor mutated owned data")
	}
	if (Example{}).Valid() {
		t.Fatal("zero example valid")
	}
}

func TestRejectMalformedOrIneligibleRows(t *testing.T) {
	base := firstRow()
	cases := []struct {
		name, old, replacement string
		want                   Error
	}{
		{"schema", Schema, "other-schema", ErrSchema},
		{"role", "development_train", "heldout_validation", ErrRole},
		{"feature_policy", FeaturePolicy, "whole_json", ErrMetadata},
		{"missing_label", "\"label\":true,", "", ErrMissing},
		{"null_label", "\"label\":true", "\"label\":null", ErrLabel},
		{"numeric_label", "\"label\":true", "\"label\":1", ErrLabel},
		{"missing_weight", ",\"sample_weight\":1", "", ErrMissing},
		{"null_weight", "\"sample_weight\":1", "\"sample_weight\":null", ErrWeight},
		{"zero_weight", "\"sample_weight\":1", "\"sample_weight\":0", ErrWeight},
		{"two_weight", "\"sample_weight\":1", "\"sample_weight\":2", ErrWeight},
		{"float_weight", "\"sample_weight\":1", "\"sample_weight\":1.0", ErrWeight},
		{"exponent_weight", "\"sample_weight\":1", "\"sample_weight\":1e0", ErrWeight},
		{"zero_group", "\"whole_group\":77", "\"whole_group\":0", ErrMetadata},
		{"large_group", "\"whole_group\":77", "\"whole_group\":65536", ErrMetadata},
		{"float_group", "\"whole_group\":77", "\"whole_group\":77.0", ErrMetadata},
		{"source_revision", "c966cfef47379dcb01e7929504d66d94b540945b", "../private-source", ErrMetadata},
		{"source_uppercase", "c966cfef47379dcb01e7929504d66d94b540945b", "C966cfef47379dcb01e7929504d66d94b540945b", ErrMetadata},
		{"stable_path", "next60-pflag-native-ipnet", "../private-id", ErrMetadata},
		{"empty_family", "upstream53-cobra-pflag-connected", "", ErrMetadata},
		{"duplicate_id", "ipValue.Set", "ipNetValue.Set", ErrInput},
		{"unknown_field", "\"stable_id\":", "\"unexpected\":false,\"stable_id\":", ErrField},
		{"case_alias", "\"schema\":", "\"Schema\":", ErrField},
		{"case_duplicate", "\"stable_id\":", "\"SCHEMA\":\"irrelevant\",\"stable_id\":", ErrDuplicate},
		{"escaped_duplicate", "\"stable_id\":", "\"sche\\u006da\":\"irrelevant\",\"stable_id\":", ErrDuplicate},
		{"candidate_duplicate", "\"text\":", "\"te\\u0078t\":\"irrelevant\",\"text\":", ErrDuplicate},
		{"candidate_case_alias", "\"label\":true", "\"Label\":true", ErrField},
		{"scope_duplicate", "\"input_count\":5", "\"input_count\":5,\"INPUT_COUNT\":5", ErrDuplicate},
		{"scope_missing", "\"unseen_source_guarantee\":false", "\"unseen_source_guarantee\":null", ErrScope},
		{"all_input", "\"all_input_guarantee\":false", "\"all_input_guarantee\":true", ErrScope},
		{"unseen_source", "\"unseen_source_guarantee\":false", "\"unseen_source_guarantee\":true", ErrScope},
		{"scope_count", "\"candidate_observations\":15", "\"candidate_observations\":14", ErrScope},
		{"scope_input_bound", "\"input_count\":5", "\"input_count\":4097", ErrScope},
		{"scope_unknown", "\"input_count\":5", "\"unknown\":0,\"input_count\":5", ErrField},
		{"surrogate_high", "For frozen IPv4 strings", "For \\ud800 frozen IPv4 strings", ErrUnicode},
		{"surrogate_low", "For frozen IPv4 strings", "For \\udc00 frozen IPv4 strings", ErrUnicode},
		{"unknown_value_surrogate", "\"stable_id\":", "\"unknown\":\"\\ud800\",\"stable_id\":", ErrUnicode},
		{"invalid_utf8", "For frozen IPv4 strings", "For " + string([]byte{255}) + " frozen IPv4 strings", ErrUnicode},
		{"long_request", "For frozen IPv4 strings, trim CIDRs into masked networks. Preserve the previous network on errors and reject bare addresses.", strings.Repeat("a", 513), ErrInput},
		{"many_words", "For frozen IPv4 strings, trim CIDRs into masked networks. Preserve the previous network on errors and reject bare addresses.", strings.Repeat("word ", 33), ErrInput},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw := mustReplace(t, base, c.old, c.replacement)
			e, err := LoadDevelopmentRow(strings.NewReader(raw))
			if err != c.want || e != (Example{}) {
				t.Fatalf("want fixed %v and zero example; got %v", c.want, err)
			}
		})
	}
	for _, raw := range []string{base + "{}", base + "garbage", "[" + base + "]", "null", ""} {
		if e, err := LoadDevelopmentRow(strings.NewReader(raw)); err != ErrJSON || e != (Example{}) {
			t.Fatal("trailing/non-object JSON accepted")
		}
	}
}

func TestCandidateCardinalityAndMetadataBounds(t *testing.T) {
	base := firstRow()
	start := strings.Index(base, "\"candidates\":[") + len("\"candidates\":[")
	end := strings.Index(base[start:], "],\"role\"") + start
	if start < len("\"candidates\":[") || end < start {
		t.Fatal("fixture candidates")
	}
	for _, list := range []string{"", strings.Repeat(`{"metadata_id":"x","text":"valid text","label":false,"sample_weight":1},`, 8) + `{"metadata_id":"ninth","text":"valid text","label":false,"sample_weight":1}`} {
		raw := base[:start] + list + base[end:]
		if e, err := LoadDevelopmentRow(strings.NewReader(raw)); err != ErrCandidates || e != (Example{}) {
			t.Fatal("candidate cardinality")
		}
	}
	for _, change := range [][2]string{
		{"next60-pflag-native-ipnet", strings.Repeat("a", 65)},
		{"upstream53-cobra-pflag-connected", strings.Repeat("a", 129)},
		{"native2-finite-v3", strings.Repeat("a", 129)},
	} {
		if _, err := LoadDevelopmentRow(strings.NewReader(mustReplace(t, base, change[0], change[1]))); err != ErrMetadata {
			t.Fatal("metadata bound")
		}
	}
}

type countedReader struct {
	r        io.Reader
	consumed int
}

func (r *countedReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.consumed += n
	return n, err
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("synthetic-private-reader-diagnostic")
}

func TestWireReadBoundAndRedactedErrors(t *testing.T) {
	row := firstRow()
	max := strings.Repeat(" ", MaxRowBytes-len(row)) + row
	if _, err := LoadDevelopmentRow(strings.NewReader(max)); err != nil {
		t.Fatal("exact wire cap rejected")
	}
	r := &countedReader{r: strings.NewReader(max + strings.Repeat(" ", MaxRowBytes))}
	if e, err := LoadDevelopmentRow(r); err != ErrBounds || e != (Example{}) || r.consumed != MaxRowBytes+1 {
		t.Fatal("unbounded read")
	}
	for _, reader := range []io.Reader{nil, failingReader{}} {
		e, err := LoadDevelopmentRow(reader)
		if e != (Example{}) || err != ErrRead || strings.Contains(err.Error(), "private") {
			t.Fatal("raw reader error retained")
		}
	}
}

func TestUnicodeScalarAndEscapedLiteralRemainValid(t *testing.T) {
	for _, s := range []string{"For \\ud83d\\ude00 frozen IPv4 strings", "For \\uFFFD frozen IPv4 strings", "For \\\\ud800 frozen IPv4 strings"} {
		row := mustReplace(t, firstRow(), "For frozen IPv4 strings", s)
		if _, err := LoadDevelopmentRow(strings.NewReader(row)); err != nil {
			t.Fatalf("valid scalar/literal rejected: %v", err)
		}
	}
	// Whole JSONL input is intentionally rejected; this API handles one row.
	if e, err := LoadDevelopmentRow(strings.NewReader(publicRows)); err != ErrJSON || !reflect.DeepEqual(e, Example{}) {
		t.Fatal("multiple development rows accepted")
	}
}
