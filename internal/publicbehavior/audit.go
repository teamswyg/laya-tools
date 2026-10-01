package publicbehavior

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

//go:embed audit.go
var compiledAudit []byte

func AuditSourceArtifact() SourceArtifact {
	h := sha256.Sum256(compiledAudit)
	return SourceArtifact{Path: "internal/publicbehavior/audit.go", Bytes: len(compiledAudit), SHA256: hex.EncodeToString(h[:])}
}

// Expectations are cold-path audit metadata, never model features.
type Check struct {
	Field string          `json:"field"`
	Want  json.RawMessage `json:"want"`
}

type Vector struct {
	ID               string   `json:"id"`
	PropertyID       string   `json:"property_id"`
	Family           string   `json:"source_family"`
	Input            Input    `json:"input"`
	LeftHex          string   `json:"left_utf8_hex"`
	RightHex         string   `json:"right_utf8_hex"`
	ExpectationKind  string   `json:"expectation_kind"`
	Checks           []Check  `json:"checks"`
	SourceHypothesis []Check  `json:"source_read_hypothesis"`
	Provenance       []string `json:"provenance"`
	Note             string   `json:"note"`
}

type Property struct {
	ID       string `json:"id"`
	Family   string `json:"source_family"`
	Contract string `json:"finite_contract"`
}

type Probes struct {
	Schema     string     `json:"schema"`
	Authorship string     `json:"authorship"`
	Properties []Property `json:"properties"`
	Vectors    []Vector   `json:"vectors"`
}

type CheckResult struct {
	Field   string          `json:"field"`
	Want    json.RawMessage `json:"want"`
	Got     json.RawMessage `json:"got"`
	Matches bool            `json:"matches"`
}

type VectorResult struct {
	ID              string        `json:"id"`
	PropertyID      string        `json:"property_id"`
	Family          string        `json:"source_family"`
	ExpectationKind string        `json:"expectation_kind"`
	State           string        `json:"state"`
	Observation     Observation   `json:"observation"`
	Checks          []CheckResult `json:"checks"`
	Matches         bool          `json:"matches_finite_expectation"`
}

type AuditResult struct {
	Schema                 string         `json:"schema"`
	State                  string         `json:"state"`
	Rows                   []VectorResult `json:"rows"`
	DispatchAttempts       int            `json:"dispatch_attempts"`
	CompletedObservations  int            `json:"completed_observations"`
	PrimaryAPICalls        int            `json:"primary_api_calls_excluding_getters"`
	MatchedExpectations    int            `json:"matched_expectations"`
	MismatchedExpectations int            `json:"mismatched_expectations"`
	UnknownObservations    int            `json:"unknown_observations"`
	SourceFamilies         int            `json:"source_families"`
	PropertyContracts      int            `json:"property_contracts"`
	IndependentFinalTasks  int            `json:"new_protected_final_tasks"`
	RankingRuns            int            `json:"ranking_runs"`
	Fits                   int            `json:"fits"`
	ModelCalls             int            `json:"model_calls"`
	TrainingReady          bool           `json:"training_ready"`
}

// Canonical makes duplicate/unknown fields and trailing data fail closed when
// compared with a bounded input byte stream by the CLI.
func Canonical(v any) ([]byte, error) {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func ValidateProbes(p Probes) error {
	if p.Schema != "riido-public-behavior-probes-57-v1" || p.Authorship == "" || len(p.Properties) != 6 || len(p.Vectors) == 0 || len(p.Vectors) > 64 {
		return errors.New("publicaudit_probe_shape")
	}
	for i, prop := range p.Properties {
		if prop.ID == "" || prop.Contract == "" || !knownFamily(prop.Family) {
			return errors.New("publicaudit_property_shape")
		}
		for _, prior := range p.Properties[:i] {
			if prior.ID == prop.ID {
				return errors.New("publicaudit_duplicate_property")
			}
		}
	}
	var familyCoverage [2]bool
	for _, prop := range p.Properties {
		if prop.Family == "Masterminds/semver" {
			familyCoverage[0] = true
		} else {
			familyCoverage[1] = true
		}
	}
	if !familyCoverage[0] || !familyCoverage[1] {
		return errors.New("publicaudit_source_family_coverage")
	}
	for i, v := range p.Vectors {
		if v.ID == "" || len(v.ID) > 96 || !knownFamily(v.Family) || len(v.Checks) == 0 || len(v.Checks) > 48 || len(v.Provenance) == 0 || len(v.Note) > 4096 {
			return errors.New("publicaudit_vector_shape")
		}
		for _, prior := range p.Vectors[:i] {
			if prior.ID == v.ID {
				return errors.New("publicaudit_duplicate_vector")
			}
		}
		found := false
		for _, prop := range p.Properties {
			if prop.ID == v.PropertyID && prop.Family == v.Family {
				found = true
			}
		}
		if !found || hex.EncodeToString([]byte(v.Input.Left)) != v.LeftHex || hex.EncodeToString([]byte(v.Input.Right)) != v.RightHex || !utf8.ValidString(v.Input.Left) || !utf8.ValidString(v.Input.Right) || len(v.Input.Left) > 512 || len(v.Input.Right) > 512 {
			return errors.New("publicaudit_input_binding")
		}
		if !slices.Contains([]string{"normative", "descriptive_api", "strict_policy"}, v.ExpectationKind) {
			return errors.New("publicaudit_expectation_kind")
		}
		if (v.Family == "Masterminds/semver" && !slices.Contains([]string{"strict_parse", "compare"}, v.Input.Operation)) || (v.Family == "bmatcuk/doublestar" && !slices.Contains([]string{"glob_match", "glob_validate"}, v.Input.Operation)) {
			return errors.New("publicaudit_operation_binding")
		}
		if (v.Input.Operation == "strict_parse" || v.Input.Operation == "glob_validate") && v.Input.Right != "" {
			return errors.New("publicaudit_unused_input")
		}
		for _, u := range v.Provenance {
			if !strings.HasPrefix(u, "https://") || len(u) > 512 {
				return errors.New("publicaudit_public_provenance")
			}
		}
		for _, checks := range [][]Check{v.Checks, v.SourceHypothesis} {
			for j, c := range checks {
				for _, prior := range checks[:j] {
					if prior.Field == c.Field {
						return errors.New("publicaudit_duplicate_check")
					}
				}
				if !validCheck(v.Input.Operation, c) {
					return errors.New("publicaudit_check_shape")
				}
			}
		}
		required := []string{"supported", "panicked", "error_kind"}
		switch v.Input.Operation {
		case "strict_parse":
			required = append(required, "nil_version", "major", "minor", "patch", "original", "string", "prerelease", "metadata", "error_func", "error_num", "error_cause")
		case "compare":
			required = append(required, "comparison_called", "comparison", "versions.0.nil_version", "versions.1.nil_version", "versions.0.error_kind", "versions.1.error_kind")
		case "glob_match":
			required = append(required, "matched")
		case "glob_validate":
			required = append(required, "valid_pattern")
		}
		for _, field := range required {
			found := false
			for _, c := range v.Checks {
				if c.Field == field {
					found = true
				}
			}
			if !found {
				return errors.New("publicaudit_incomplete_expectation")
			}
		}
	}
	for _, prop := range p.Properties {
		found := false
		for _, v := range p.Vectors {
			if v.PropertyID == prop.ID {
				found = true
			}
		}
		if !found {
			return errors.New("publicaudit_empty_property")
		}
	}
	return nil
}

func knownFamily(s string) bool { return s == "Masterminds/semver" || s == "bmatcuk/doublestar" }

func validCheck(op string, c Check) bool {
	field := c.Field
	versionField := false
	if strings.HasPrefix(field, "versions.") {
		parts := strings.Split(field, ".")
		if op != "compare" || len(parts) != 3 || (parts[1] != "0" && parts[1] != "1") {
			return false
		}
		field = parts[2]
		versionField = true
	}
	if versionField && !slices.Contains([]string{"called", "returned", "nil_version", "error_kind", "error_func", "error_num", "error_cause", "error_message", "major", "minor", "patch", "original", "string", "prerelease", "metadata"}, field) {
		return false
	}
	if !versionField {
		common := slices.Contains([]string{"supported", "panicked", "error_kind", "error_func", "error_num", "error_cause", "error_message"}, field)
		allowed := common
		switch op {
		case "strict_parse":
			allowed = allowed || slices.Contains([]string{"nil_version", "major", "minor", "patch", "original", "string", "prerelease", "metadata", "parse_error_side"}, field)
		case "compare":
			allowed = allowed || slices.Contains([]string{"comparison", "comparison_called", "parse_error_side"}, field)
		case "glob_match":
			allowed = allowed || field == "matched"
		case "glob_validate":
			allowed = allowed || field == "valid_pattern"
		}
		if !allowed {
			return false
		}
	}
	if len(c.Want) == 0 || len(c.Want) > 1024 {
		return false
	}
	var value any
	// A single scalar, in its exact canonical spelling. Numeric decoding here
	// uses uint64/int directly, never a float, so large integers stay exact.
	switch field {
	case "called", "returned", "nil_version", "panicked", "matched", "valid_pattern", "supported", "comparison_called":
		var x bool
		if json.Unmarshal(c.Want, &x) != nil {
			return false
		}
		value = x
	case "major", "minor", "patch":
		var x uint64
		if json.Unmarshal(c.Want, &x) != nil {
			return false
		}
		value = x
	case "comparison":
		var x int
		if json.Unmarshal(c.Want, &x) != nil || x < -1 || x > 1 {
			return false
		}
		value = x
	case "error_kind", "error_func", "error_num", "error_cause", "error_message", "original", "string", "prerelease", "metadata", "parse_error_side":
		var x string
		if json.Unmarshal(c.Want, &x) != nil {
			return false
		}
		value = x
	default:
		return false
	}
	raw, err := json.Marshal(value)
	return err == nil && bytes.Equal(raw, c.Want)
}

func observationField(raw []byte, path string) (json.RawMessage, error) {
	parts := strings.Split(path, ".")
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return nil, errors.New("publicaudit_observation_shape")
	}
	if len(parts) == 1 {
		v, ok := fields[path]
		if !ok {
			return nil, errors.New("publicaudit_unknown_field")
		}
		return v, nil
	}
	if len(parts) != 3 || parts[0] != "versions" {
		return nil, errors.New("publicaudit_unknown_field")
	}
	var versions [2]json.RawMessage
	if json.Unmarshal(fields["versions"], &versions) != nil {
		return nil, errors.New("publicaudit_version_shape")
	}
	i, err := strconv.Atoi(parts[1])
	if err != nil || i < 0 || i > 1 {
		return nil, errors.New("publicaudit_version_index")
	}
	fields = nil
	if json.Unmarshal(versions[i], &fields) != nil {
		return nil, errors.New("publicaudit_version_shape")
	}
	v, ok := fields[parts[2]]
	if !ok {
		return nil, errors.New("publicaudit_unknown_field")
	}
	return v, nil
}

// Audit keeps implementation hypotheses out of acceptance. All vectors run,
// including normative and policy disagreements. No model features are made.
func Audit(p Probes) (AuditResult, error) {
	if err := ValidateProbes(p); err != nil {
		return AuditResult{}, err
	}
	out := AuditResult{Schema: "riido-public-behavior-observations-57-v1", State: "incomplete", SourceFamilies: 2, PropertyContracts: len(p.Properties)}
	for _, v := range p.Vectors {
		out.DispatchAttempts++
		got := Observe(v.Input)
		out.PrimaryAPICalls += got.UpstreamAPICalls
		r := VectorResult{ID: v.ID, PropertyID: v.PropertyID, Family: v.Family, ExpectationKind: v.ExpectationKind, State: "complete", Observation: got, Matches: true}
		unknown := !got.Supported || got.RejectedInputBounds || got.Panicked || got.ErrorKind == "unclassified_error" || got.ErrorCause == "unclassified_error"
		for _, ver := range got.Versions {
			if ver.Called && (!ver.Returned || ver.ErrorKind == "unclassified_error" || ver.ErrorCause == "unclassified_error") {
				unknown = true
			}
		}
		if unknown {
			r.State = "unknown"
			r.Matches = false
			out.UnknownObservations++
			out.Rows = append(out.Rows, r)
			continue
		}
		raw, err := json.Marshal(got)
		if err != nil {
			return out, errors.New("publicaudit_observation_encoding")
		}
		for _, check := range v.Checks {
			field, err := observationField(raw, check.Field)
			if err != nil {
				out.State = "evidence_binding_failed"
				return out, fmt.Errorf("publicaudit_evidence_field:%s", check.Field)
			}
			matches := bytes.Equal(field, check.Want)
			r.Checks = append(r.Checks, CheckResult{check.Field, check.Want, field, matches})
			r.Matches = r.Matches && matches
		}
		out.CompletedObservations++
		if r.Matches {
			out.MatchedExpectations++
		} else {
			out.MismatchedExpectations++
		}
		out.Rows = append(out.Rows, r)
	}
	out.State = "complete"
	if out.UnknownObservations != 0 {
		out.State = "complete_with_unknown"
	}
	return out, nil
}
