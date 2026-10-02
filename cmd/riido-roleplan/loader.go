// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/roleplan"
)

const inputSHA = "ae2794224ab8fdb03aa879b2522777beb5065005286b6394db3ba4a8abcdd1ce"
const roleSourceSHA = "df8d53d0600e80c331c4912d90ff199aef753704f34c87e085fa738fe920ba0b"
const roleTestSHA = "c858a8138bde70629a51effaeb8259878419afffe7d2f5e0880f1b5bb79e9b92"
const auditSourceSHA = "f1fb50eb693baa06cd1e9df7ad57275ee557a6bcc17439befbaf9617a6082224"
const maxInputBytes = 8 << 20
const maxOutputBytes = 1 << 20
const planSchema = "riido-whole-group-role-execution-plan-66-v2"
const inputSchema = "riido-complete-stored-membership-proposal-65-v1"

type pin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}
type reference struct {
	Path        string `json:"path"`
	ArtifactSHA string `json:"artifact_sha256"`
	Pointer     string `json:"json_pointer"`
	ValueSHA    string `json:"canonical_go_json_value_sha256"`
}
type parent struct {
	ID         string    `json:"parent_id"`
	Index      int       `json:"original_parent_index"`
	Truth      string    `json:"saved_truth_state"`
	Original   reference `json:"original_parent_reference"`
	Outcome    reference `json:"saved_outcome_reference"`
	Candidates int       `json:"candidate_positions"`
}
type memberBinding struct {
	ID         string      `json:"id"`
	Kind       string      `json:"kind"`
	References []reference `json:"references"`
}
type relationBinding struct {
	From       string      `json:"from"`
	Kind       string      `json:"kind"`
	To         string      `json:"to"`
	References []reference `json:"references"`
}
type group struct {
	ID               string                  `json:"component_id"`
	OriginalID       int64                   `json:"original_group_id"`
	Known            bool                    `json:"has_known_members"`
	Original         reference               `json:"original_whole_group_reference"`
	Parents          []parent                `json:"parents"`
	Members          []string                `json:"canonical_members"`
	Relations        []roleplan.Relationship `json:"relationships"`
	MemberBindings   []memberBinding         `json:"member_bindings"`
	RelationBindings []relationBinding       `json:"relationship_bindings"`
	Digest           string                  `json:"proposal_membership_sha256"`
	EncodedBytes     int                     `json:"proposal_encoded_bytes"`
}
type snapshot struct {
	Schema         string          `json:"schema"`
	State          string          `json:"state"`
	Inputs         []pin           `json:"input_artifacts"`
	Groups         []group         `json:"groups"`
	Encoding       json.RawMessage `json:"encoding_proposal"`
	Checks         json.RawMessage `json:"checks"`
	Preservation   json.RawMessage `json:"preservation"`
	NonGroupPolicy json.RawMessage `json:"non_group_provenance_policy"`
	RelationRecipe json.RawMessage `json:"relation_recipe"`
	Gaps           []string        `json:"unknowns_and_gaps"`
	SeedAssigned   bool            `json:"seed_assigned"`
	OrderCalls     int             `json:"OrderDigest_calls"`
	AllocateCalls  int             `json:"AllocateCounts_calls"`
	AssignCalls    int             `json:"Assign_calls"`
	Roles          int             `json:"roles_assigned"`
	Fits           int             `json:"fits"`
	NewLabels      int             `json:"new_labels"`
	CandidateCalls int             `json:"candidate_API_calls"`
	SourceCalls    int             `json:"source_AST_Rebind_SourcePins_Bind_Generate_calls"`
	Models         int             `json:"models"`
	Paid           int             `json:"paid_API_processes"`
	Final          int             `json:"protected_final_reads"`
	TrainingReady  bool            `json:"training_ready"`
}
type requirement struct {
	Key          string   `json:"key"`
	Role         string   `json:"role"`
	ComponentIDs []string `json:"component_ids"`
	Minimum      int      `json:"minimum_known_components"`
}
type executionPlan struct {
	Schema              string        `json:"schema"`
	State               string        `json:"state"`
	ExecutionAuthorized bool          `json:"execution_authorized"`
	Input               pin           `json:"membership_input"`
	InputSources        []pin         `json:"input_sources"`
	BindingEvidence     []pin         `json:"stored_binding_evidence"`
	SupportSources      []pin         `json:"support_sources"`
	Implementation      []pin         `json:"implementation_sources"`
	BinarySHA           string        `json:"binary_sha256"`
	GoVersion           string        `json:"go_version"`
	CGO                 string        `json:"cgo_enabled"`
	Trimpath            bool          `json:"trimpath"`
	CPUThreads          int           `json:"cpu_threads"`
	HeapSoftLimit       int64         `json:"go_heap_soft_limit_bytes"`
	Seed                string        `json:"seed_ascii"`
	SeedHex             string        `json:"seed_hex"`
	MembershipDomain    string        `json:"membership_domain"`
	OrderDomain         string        `json:"order_domain"`
	Encoding            string        `json:"membership_encoding"`
	Algorithm           string        `json:"allocation_algorithm"`
	KnownFloors         [3]int        `json:"existing_known_group_floors"`
	Coverage            []requirement `json:"stored_truth_coverage"`
	PlannedExecutions   int           `json:"planned_official_executions"`
	Retries             int           `json:"retries"`
	Fits                int           `json:"fits"`
	Models              int           `json:"models"`
	Paid                int           `json:"paid_calls"`
	Final               int           `json:"protected_final_reads"`
	TrainingReady       bool          `json:"training_ready"`
	RoleFeatures        bool          `json:"roles_or_graph_as_scorer_features"`
	Masks               bool          `json:"loss_masks_or_new_labels"`
	UnknownPolicy       string        `json:"unknown_policy"`
	CoveragePolicy      string        `json:"coverage_policy"`
	StopPolicy          string        `json:"failure_policy"`
	TrustLimits         []string      `json:"trust_limits"`
}
type loaded struct {
	Snapshot    snapshot
	Components  []roleplan.Component
	Memberships []roleplan.Membership
	Coverage    roleplan.CoveragePlan
}

func errCode(code string) error { return errors.New(code) }
func hash(b []byte) string      { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validSHA(s string) bool {
	if len(s) != 64 || strings.ToLower(s) != s {
		return false
	}
	b, e := hex.DecodeString(s)
	return e == nil && !bytes.Equal(b, make([]byte, 32))
}
func safePath(s string) bool {
	return s != "" && utf8.ValidString(s) && !strings.Contains(s, "\\") && !filepath.IsAbs(s) && filepath.ToSlash(filepath.Clean(s)) == s && s != "." && s != ".." && !strings.HasPrefix(s, "../")
}
func readBounded(path string, limit int) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, errCode("input_open_failed")
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil || !s.Mode().IsRegular() || s.Size() < 1 || s.Size() > int64(limit) {
		return nil, errCode("input_bounds")
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if e != nil || len(b) > limit {
		return nil, errCode("input_read_failed")
	}
	return b, nil
}
func pinnedBytes(path string, p pin, limit int) ([]byte, error) {
	if !validSHA(p.SHA256) || p.Bytes < 1 || p.Bytes > limit {
		return nil, errCode("pin_invalid")
	}
	b, e := readBounded(path, limit)
	if e != nil {
		return nil, e
	}
	if len(b) != p.Bytes || hash(b) != p.SHA256 {
		return nil, errCode("input_pin_mismatch")
	}
	return b, nil
}

// Duplicate detection is performed before decoding, including nested metadata.
// It is bounded maintainer parsing, not original source interpretation.
func jsonValue(d *json.Decoder, depth int) error {
	if depth > 128 {
		return errCode("json_depth")
	}
	t, e := d.Token()
	if e != nil {
		return errCode("json_invalid")
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		var keys []string
		for d.More() {
			t, e = d.Token()
			if e != nil {
				return errCode("json_invalid")
			}
			k, ok := t.(string)
			if !ok || len(k) > 512 {
				return errCode("json_key_invalid")
			}
			for _, old := range keys {
				if old == k {
					return errCode("json_key_duplicate")
				}
			}
			keys = append(keys, k)
			if len(keys) > 128 {
				return errCode("json_object_bounds")
			}
			if e = jsonValue(d, depth+1); e != nil {
				return e
			}
		}
		t, e = d.Token()
		if e != nil || t != json.Delim('}') {
			return errCode("json_invalid")
		}
	case '[':
		n := 0
		for d.More() {
			n++
			if n > 65536 {
				return errCode("json_array_bounds")
			}
			if e = jsonValue(d, depth+1); e != nil {
				return e
			}
		}
		t, e = d.Token()
		if e != nil || t != json.Delim(']') {
			return errCode("json_invalid")
		}
	default:
		return errCode("json_invalid")
	}
	return nil
}
func decodeStrict(b []byte, out any) error {
	if len(b) == 0 || len(b) > maxInputBytes || !utf8.Valid(b) || bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		return errCode("json_invalid")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if e := jsonValue(d, 0); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return errCode("json_multiple_values")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(out); e != nil {
		return errCode("json_shape_invalid")
	}
	return nil
}
func canonical(v any) ([]byte, error) {
	raw, e := json.Marshal(v)
	if e != nil {
		return nil, errCode("json_encode_failed")
	}
	// Plan/result canonicalization: Go UseNumber decoding followed by sorted
	// object-key MarshalIndent, two spaces and one final LF. No input65 bytes
	// are rewritten by this recipe.
	var tree any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if e = d.Decode(&tree); e != nil {
		return nil, errCode("json_encode_failed")
	}
	b, e := json.MarshalIndent(tree, "", "  ")
	if e != nil {
		return nil, errCode("json_encode_failed")
	}
	return append(b, '\n'), nil
}
func pinnedListEqual(a, b []pin) bool { return reflect.DeepEqual(a, b) }
func expectedKnownIDs() []string {
	ids := []int{0, 4, 8, 12, 16, 20, 24, 28, 32, 40, 44, 48, 52, 56, 60, 68}
	out := make([]string, len(ids))
	for i, n := range ids {
		out[i] = "original-whole-group:" + strconv.Itoa(n)
	}
	return out
}
func expectedCoverage() []requirement {
	var out []requirement
	for _, key := range []string{"stored_truth_answerable", "stored_truth_no_answer"} {
		for i, role := range []string{"development_train", "development_validation", "development_calibration"} {
			out = append(out, requirement{key, role, expectedKnownIDs(), [3]int{9, 3, 3}[i]})
		}
	}
	return out
}
func compileCoverage(qs []requirement) (roleplan.CoveragePlan, error) {
	p := roleplan.CoveragePlan{Declared: true}
	for _, q := range qs {
		var r roleplan.Role
		switch q.Role {
		case "development_train":
			r = roleplan.DevelopmentTrain
		case "development_validation":
			r = roleplan.DevelopmentValidation
		case "development_calibration":
			r = roleplan.DevelopmentCalibration
		default:
			return p, errCode("coverage_role_invalid")
		}
		p.Requirements = append(p.Requirements, roleplan.CoverageRequirement{Key: q.Key, Role: r, ComponentIDs: append([]string(nil), q.ComponentIDs...), MinimumKnownComponents: q.Minimum})
	}
	return p, nil
}
func validatePlan(p executionPlan) error {
	if p.Schema != planSchema || p.State != "private_pre_execution_draft" && p.State != "frozen_for_whole_group_assignment" || p.ExecutionAuthorized != (p.State == "frozen_for_whole_group_assignment") {
		return errCode("plan_state_invalid")
	}
	if p.Input.SHA256 != inputSHA || p.Input.Bytes != 1986200 || p.Input.Path != "membership-proposal-65.json" {
		return errCode("plan_input_invalid")
	}
	if p.GoVersion != "go1.27.1" || p.CGO != "0" || !p.Trimpath || p.CPUThreads != 1 || p.HeapSoftLimit != 256<<20 {
		return errCode("plan_runtime_invalid")
	}
	if p.Seed != "1729" || p.SeedHex != "31373239" || p.MembershipDomain != roleplan.MembershipDomain || p.OrderDomain != roleplan.OrderDomain || p.Encoding != "LP64BE_UTF8_members_sorted_relationships_From_Kind_To_v1" || p.Algorithm != "whole-component-sha256-order-weighted-3-1-1-v1" || p.KnownFloors != [3]int{9, 3, 3} {
		return errCode("plan_recipe_invalid")
	}
	if !reflect.DeepEqual(p.Coverage, expectedCoverage()) {
		return errCode("plan_coverage_invalid")
	}
	if p.PlannedExecutions != 1 || p.Retries != 0 || p.Fits != 0 || p.Models != 0 || p.Paid != 0 || p.Final != 0 || p.TrainingReady || p.RoleFeatures || p.Masks {
		return errCode("plan_claim_invalid")
	}
	if p.UnknownPolicy != "retain_all_21_unknown;whole_group_64_train_provenance;never_labels_or_labeled_floor" || p.CoveragePolicy != "stored_truth_only;answerable_and_no_answer_sets_identical;no_future_loss_eligibility_or_semantic_readiness" || p.StopPolicy != "one_attempt;preserve_refusal;no_seed_search_split_merge_masks_or_retry" {
		return errCode("plan_policy_invalid")
	}
	if !reflect.DeepEqual(p.TrustLimits, trustLimits()) {
		return errCode("plan_limits_invalid")
	}
	if !pinnedListEqual(p.InputSources, requiredInputs()) || !pinnedListEqual(p.SupportSources, requiredSupport()) || !pinnedListEqual(p.BindingEvidence, requiredEvidence()) {
		return errCode("plan_source_pins_invalid")
	}
	if len(p.Implementation) != 6 {
		return errCode("plan_implementation_invalid")
	}
	names := [6]string{"cmd/riido-roleplan/main.go", "cmd/riido-roleplan/loader.go", "internal/roleplan/role.go", "internal/roleplan/audit_source.go", "go.mod", "go.sum"}
	for i, f := range p.Implementation {
		if f.Path != names[i] || !validSHA(f.SHA256) || f.Bytes < 1 || f.Bytes > maxInputBytes {
			return errCode("plan_implementation_invalid")
		}
	}
	if p.Implementation[2].SHA256 != roleSourceSHA {
		return errCode("plan_encoder_source_invalid")
	}
	if p.Implementation[3].SHA256 != auditSourceSHA {
		return errCode("plan_provenance_helper_invalid")
	}
	if p.ExecutionAuthorized && !validSHA(p.BinarySHA) || !p.ExecutionAuthorized && p.BinarySHA != "" && !validSHA(p.BinarySHA) {
		return errCode("plan_binary_invalid")
	}
	return nil
}
func requiredInputs() []pin {
	return []pin{
		{"experiments/short-claim/results-56b.json", "4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4", 266818},
		{"experiments/short-claim/probes-56.json", "0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0", 49941},
		{"experiments/short-claim/probes-56b.json", "0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e", 43598},
		{"experiments/short-claim/source-inventory-60.json", "d99d635929555d4123bbaa84aac99d0fb29db345f8b813ed70074dd3e09cdbd4", 174651},
		{"experiments/short-claim/content-review-assessments-60.json", "1beb5c06d3fb70c0a5e0727de793f1890b4ab4a47f1d392a6666b068f096c526", 1280021},
		{"experiments/short-claim/content-review-manifest-61.json", "f8f0b2320bb214f102575384cace477067a0f075b98ee09c071e778d9285224b", 467067},
		{"experiments/short-claim/content-review-mechanics-61.json", "1f8ee9bcf1f62cc18a575c86e413fd50384634bc213dd0e0dd215fc1890ab5d8", 9669},
		{"experiments/short-claim/role-recipe-59.json", "a4a6b1094f3b279ef16748dbd70a801a72f259a66e04cf379b75e7913a9a3619", 3265},
	}
}
func requiredSupport() []pin {
	return []pin{
		{"internal/roleplan/role.go", roleSourceSHA, 16123},
		{"internal/roleplan/role_test.go", roleTestSHA, 24825},
		{"experiments/short-claim/role-preparation/encoding-proposal-62.json", "aaf05ac7f335a1f57ccdd1f3b1dfc1c4c3d533a49a9d018ffdc6232236ef741f", 5315},
	}
}

func requiredEvidence() []pin {
	return []pin{
		{"binding-receipt-65.json", "c6be0a869af3946e9c4c52b602be49f08a74fe7bce9abe810f9f93b7e747ab9b", 15776},
		{"INDEPENDENT-AUDIT-65.json", "fb1314fa6bc35d297fa12f1c2c8e3e4528a0bef33198cded05eb4da71cf42a0d", 13400},
	}
}

func trustLimits() []string {
	return []string{
		"stored graph completeness only; missing external edges not proved absent",
		"source/text uncertainty remains; no loss eligibility or semantic readiness",
		"previously observed development data; not blinded or protected final",
		"English finite Go claims from a synthetic caption pipeline; source diversity uncleared",
	}
}

func checkClosedSourceRoot(root string) error {
	for _, dir := range []string{"cmd/riido-roleplan", "internal/roleplan"} {
		entries, e := os.ReadDir(filepath.Join(root, dir))
		if e != nil {
			return errCode("implementation_directory_failed")
		}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".go") {
				continue
			}
			if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
				return errCode("implementation_source_shape_invalid")
			}
			allowed := false
			if dir == "cmd/riido-roleplan" {
				switch name {
				case "main.go", "loader.go", "main_test.go":
					allowed = true
				}
			} else {
				// Exact known test exceptions are not extra runtime declarations.
				// Non-test sources remain only role.go and audit_source.go.
				switch name {
				case "role.go", "audit_source.go", "role_test.go", "audit_source_test.go", "archives_test.go", "frozen_membership65_test.go":
					allowed = true
				}
			}
			if !allowed {
				return errCode("implementation_extra_go_source")
			}
		}
	}
	return nil
}

// No membership encoding, graph union, candidate execution, role order or role
// assignment happens here. Completeness relies on the exact independently
// checked stored65 artifact, not on an opaque digest or this structural check.
func validateSnapshot(s snapshot) (loaded, error) {
	var out loaded
	out.Snapshot = s
	if s.Schema != inputSchema || s.State != "private_metadata_proposal_no_seed_or_roles" || s.SeedAssigned || s.OrderCalls != 0 || s.AllocateCalls != 0 || s.AssignCalls != 0 || s.Roles != 0 || s.Fits != 0 || s.NewLabels != 0 || s.CandidateCalls != 0 || s.SourceCalls != 0 || s.Models != 0 || s.Paid != 0 || s.Final != 0 || s.TrainingReady {
		return out, errCode("snapshot_state_invalid")
	}
	if len(s.Groups) != 17 || !pinnedListEqual(s.Inputs, requiredInputs()) {
		return out, errCode("snapshot_count_or_pins_invalid")
	}
	expected := []int64{0, 4, 8, 12, 16, 20, 24, 28, 32, 40, 44, 48, 52, 56, 60, 64, 68}
	var seen [72]bool
	var known, noanswer, unknown, candidates int
	var ids []string
	for i, g := range s.Groups {
		if g.OriginalID != expected[i] || g.ID != "original-whole-group:"+strconv.FormatInt(g.OriginalID, 10) || g.Known != (g.OriginalID != 64) || !validSHA(g.Digest) || len(g.Members) == 0 || g.EncodedBytes < 1 || g.EncodedBytes > roleplan.MaxEncodedMembershipBytes {
			return out, errCode("snapshot_group_invalid")
		}
		digest, e := hex.DecodeString(g.Digest)
		if e != nil {
			return out, errCode("snapshot_digest_invalid")
		}
		var d [32]byte
		copy(d[:], digest)
		for j, m := range g.Members {
			if len(m) == 0 || len(m) > 512 || !utf8.ValidString(m) || j > 0 && g.Members[j-1] >= m {
				return out, errCode("snapshot_member_invalid")
			}
		}
		if len(g.MemberBindings) != len(g.Members) || len(g.RelationBindings) != len(g.Relations) {
			return out, errCode("snapshot_binding_count_invalid")
		}
		hasKnown, hasNoAnswer := false, false
		for _, p := range g.Parents {
			if p.Index < 0 || p.Index >= 72 || seen[p.Index] || len(p.ID) == 0 || len(p.ID) > 512 || !utf8.ValidString(p.ID) || p.Candidates < 1 || p.Candidates > 8 {
				return out, errCode("snapshot_parent_invalid")
			}
			seen[p.Index] = true
			ids = append(ids, p.ID)
			candidates += p.Candidates
			switch p.Truth {
			case "known":
				known++
				hasKnown = true
			case "no_answer":
				noanswer++
				hasNoAnswer = true
			case "unknown":
				unknown++
			default:
				return out, errCode("snapshot_truth_invalid")
			}
		}
		if g.Known != hasKnown || g.Known != hasNoAnswer {
			return out, errCode("stored_truth_coverage_mismatch")
		}
		out.Components = append(out.Components, roleplan.Component{ID: g.ID, OriginalGroupID: g.OriginalID, MembershipSHA256: d, HasKnownMembers: g.Known})
		out.Memberships = append(out.Memberships, roleplan.Membership{Members: append([]string(nil), g.Members...), Relationships: append([]roleplan.Relationship(nil), g.Relations...)})
	}
	sort.Strings(ids)
	for i := 1; i < len(ids); i++ {
		if ids[i-1] == ids[i] {
			return out, errCode("snapshot_parent_duplicate")
		}
	}
	if len(ids) != 72 || known != 34 || noanswer != 17 || unknown != 21 || candidates != 216 {
		return out, errCode("snapshot_truth_count_invalid")
	}
	for _, v := range seen {
		if !v {
			return out, errCode("snapshot_parent_missing")
		}
	}
	return out, nil
}
func loadBound(input, repoRoot, sourceRoot string, p executionPlan) (loaded, error) {
	var out loaded
	b, e := pinnedBytes(input, p.Input, maxInputBytes)
	if e != nil {
		return out, e
	}
	if e = checkClosedSourceRoot(sourceRoot); e != nil {
		return out, e
	}
	for _, f := range p.BindingEvidence {
		if !safePath(f.Path) {
			return out, errCode("evidence_path_invalid")
		}
		if _, e = pinnedBytes(filepath.Join(filepath.Dir(input), f.Path), f, maxInputBytes); e != nil {
			return out, e
		}
	}
	for _, f := range p.InputSources {
		if !safePath(f.Path) {
			return out, errCode("source_path_invalid")
		}
		if _, e = pinnedBytes(filepath.Join(repoRoot, filepath.FromSlash(f.Path)), f, maxInputBytes); e != nil {
			return out, e
		}
	}
	for _, f := range p.SupportSources {
		if !safePath(f.Path) || !validSHA(f.SHA256) {
			return out, errCode("source_pin_invalid")
		}
		if _, e = pinnedBytes(filepath.Join(repoRoot, filepath.FromSlash(f.Path)), f, maxInputBytes); e != nil {
			return out, e
		}
	}
	compiledRole := roleplan.CompiledSourceSHA256()
	if hex.EncodeToString(compiledRole[:]) != roleSourceSHA {
		return out, errCode("compiled_role_source_mismatch")
	}
	for _, f := range p.Implementation {
		b, e := pinnedBytes(filepath.Join(sourceRoot, filepath.FromSlash(f.Path)), f, maxInputBytes)
		if e != nil {
			return out, e
		}
		if strings.HasPrefix(f.Path, "cmd/riido-roleplan/") {
			compiled, e := compiledFiles.ReadFile(filepath.Base(f.Path))
			if e != nil || !bytes.Equal(b, compiled) {
				return out, errCode("compiled_source_mismatch")
			}
		}
	}
	var s snapshot
	if e = decodeStrict(b, &s); e != nil {
		return out, e
	}
	out, e = validateSnapshot(s)
	if e != nil {
		return out, e
	}
	out.Coverage, e = compileCoverage(p.Coverage)
	return out, e
}
