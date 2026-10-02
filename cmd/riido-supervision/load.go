// SPDX-License-Identifier: Apache-2.0
// Public maintainer port of the unchanged private67 metadata loader.
// Prototype loader SHA: f226630753c3478f74a9d925e949f4231bab251d24b14e3da793bfbef944ece8.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

const maxFileBytes = 2 << 20
const maxTotalBytes = 16 << 20
const planSHA = "c9a07943f8047b69e32d93354e812ca5b158b1dc7476264bd5b8505a86242408"
const basePath = "experiments/short-claim/"
const consistent = "consistent_with_scoped_evidence"
const contradicts = "contradicts_scoped_evidence"

type Pin struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
	SHA   string `json:"sha256"`
}
type Expected struct{ Parents, Candidates, KnownOrNoAnswer, Unknown, Groups, Original, Supplement int }
type Config struct {
	Pins              [6]Pin
	Expected          Expected
	MaxFile, MaxTotal int64
}

var frozenConfig = Config{Pins: [6]Pin{
	{basePath + "probes-56.json", 49941, "0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0"},
	{basePath + "probes-56b.json", 43598, "0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e"},
	{basePath + "results-56b.json", 266818, "4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4"},
	{basePath + "caption-coverage-59.json", 360591, "7c1bd449d533d3d46a9211dea0a436ce338fe8f16ee3197d554744c3c24b0a29"},
	{basePath + "content-review-assessments-60.json", 1280021, "1beb5c06d3fb70c0a5e0727de793f1890b4ab4a47f1d392a6666b068f096c526"},
	{basePath + "content-review-manifest-61.json", 467067, "f8f0b2320bb214f102575384cace477067a0f075b98ee09c071e778d9285224b"},
}, Expected: Expected{72, 216, 51, 21, 17, 738, 432}, MaxFile: maxFileBytes, MaxTotal: maxTotalBytes}

type Ledger struct {
	FileReadAttempts         int   `json:"file_read_attempts"`
	FilesVerified            int   `json:"files_verified"`
	BytesRead                int64 `json:"bytes_read"`
	BytesVerified            int64 `json:"bytes_verified"`
	OriginalBindingsVerified int   `json:"original_bindings_verified"`
	ReviewRecordsVerified    int   `json:"review_records_verified"`
	OriginalSlotsJoined      int   `json:"original_slots_joined"`
	SupplementalSlotsJoined  int   `json:"supplemental_slots_joined"`
}
type loader struct {
	root   *os.Root
	cfg    Config
	ledger Ledger
	pins   []Pin
}

func fail(code string) error { return errors.New("scope67_" + code) }
func sha(b []byte) string    { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func hexSHA(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func safePath(s string) bool {
	return s != "" && s == path.Clean(s) && !strings.HasPrefix(s, "/") && s != ".." && !strings.HasPrefix(s, "../") && !strings.ContainsRune(s, '\\')
}

// Every JSON object rejects duplicate decoded keys, including escaped aliases.
// Objects/arrays are bounded by the byte cap; depth and per-object key count
// additionally bound parser work. No original runtime package is imported.
func jsonValue(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fail("json_depth")
	}
	t, e := d.Token()
	if e != nil {
		return fail("json")
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		var keys []string
		for d.More() {
			k, e := d.Token()
			s, ok := k.(string)
			if e != nil || !ok || len(keys) >= 512 {
				return fail("json_object")
			}
			for _, old := range keys {
				if old == s {
					return fail("json_duplicate")
				}
			}
			keys = append(keys, s)
			if e := jsonValue(d, depth+1); e != nil {
				return e
			}
		}
		t, e = d.Token()
		if e != nil || t != json.Delim('}') {
			return fail("json")
		}
	case '[':
		for d.More() {
			if e := jsonValue(d, depth+1); e != nil {
				return e
			}
		}
		t, e = d.Token()
		if e != nil || t != json.Delim(']') {
			return fail("json")
		}
	default:
		return fail("json")
	}
	return nil
}
func strictJSON(b []byte, out any) error {
	if !utf8.Valid(b) || !json.Valid(b) {
		return fail("json")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if e := jsonValue(d, 0); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return fail("json_trailing")
	}
	if json.Unmarshal(b, out) != nil {
		return fail("json_shape")
	}
	return nil
}
func (l *loader) read(p Pin, out any) error {
	l.ledger.FileReadAttempts++
	if !safePath(p.Path) || p.Bytes < 0 || p.Bytes > l.cfg.MaxFile || !hexSHA(p.SHA) || l.ledger.BytesRead > l.cfg.MaxTotal-p.Bytes {
		return fail("read_bounds")
	}
	st, e := l.root.Stat(p.Path)
	if e != nil || !st.Mode().IsRegular() || st.Size() != p.Bytes {
		return fail("regular_size")
	}
	f, e := l.root.Open(p.Path)
	if e != nil {
		return fail("read")
	}
	defer f.Close()
	st, e = f.Stat()
	if e != nil || !st.Mode().IsRegular() || st.Size() != p.Bytes {
		return fail("regular_size")
	}
	b := make([]byte, p.Bytes)
	n, e := io.ReadFull(f, b)
	l.ledger.BytesRead += int64(n)
	if e != nil || int64(n) != p.Bytes || sha(b) != p.SHA {
		return fail("pin")
	}
	st, e = f.Stat()
	if e != nil || !st.Mode().IsRegular() || st.Size() != p.Bytes {
		return fail("regular_size")
	}
	if e := strictJSON(b, out); e != nil {
		return e
	}
	l.pins = append(l.pins, p)
	l.ledger.FilesVerified++
	l.ledger.BytesVerified += p.Bytes
	return nil
}

type TextRef struct {
	Path            string `json:"path"`
	Pointer         string `json:"json_pointer"`
	SHA             string `json:"decoded_utf8_sha256"`
	Bytes           int    `json:"decoded_utf8_bytes"`
	Artifact        *Pin   `json:"artifact"`
	ContractPointer string `json:"reference59_contract_pointer"`
}
type Candidate struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	SourceID  string `json:"source_id"`
	CodeSHA   string `json:"code_sha256"`
	BundleSHA string `json:"source_bundle_sha256"`
}
type Parent struct {
	ID         string      `json:"id"`
	Prototype  string      `json:"prototype"`
	ContractID string      `json:"contract_id"`
	Request    string      `json:"request"`
	Candidates []Candidate `json:"candidates"`
}
type Probes struct {
	Schema  string   `json:"schema"`
	Parents []Parent `json:"parents"`
}
type CandidateTruth struct {
	ID    string `json:"candidate_id"`
	State string `json:"state"`
}
type Truth struct {
	ID         string           `json:"parent_id"`
	State      string           `json:"state"`
	Acceptable []int            `json:"acceptable_candidate_indices"`
	Candidates []CandidateTruth `json:"candidate_truth"`
}
type Group struct {
	ID            int      `json:"id"`
	Parents       []string `json:"parents"`
	Prototypes    []string `json:"prototypes"`
	Sources       []string `json:"sources"`
	CoreTemplates []string `json:"core_templates"`
}
type Saved struct {
	Schema    string `json:"schema"`
	InputSHA  string `json:"input_sha256"`
	LegacySHA string `json:"legacy_input_sha256"`
	Legacy    struct {
		Outcomes []Truth `json:"outcomes"`
	} `json:"legacy_truth"`
	Typed struct {
		Outcomes []Truth `json:"outcomes"`
	} `json:"typed_truth"`
	Combined struct {
		Groups []Group `json:"groups"`
	} `json:"combined_groups"`
}
type ReferenceCandidate struct {
	Index     int            `json:"candidate_index"`
	ID        string         `json:"candidate_id"`
	Caption   TextRef        `json:"caption"`
	SourceID  string         `json:"source_id"`
	CodeSHA   string         `json:"historical_code_sha256"`
	BundleSHA string         `json:"historical_bundle_sha256"`
	Outcome   CandidateTruth `json:"historical_candidate_outcome"`
}
type ReferenceParent struct {
	Index      int                  `json:"original_parent_index"`
	ID         string               `json:"parent_id"`
	Cohort     string               `json:"cohort"`
	Prototype  string               `json:"prototype"`
	ContractID string               `json:"contract_id"`
	Group      int                  `json:"historical_group_id"`
	Request    TextRef              `json:"request"`
	State      string               `json:"historical_truth_state"`
	Acceptable []int                `json:"historical_acceptable_candidate_indices"`
	Candidates []ReferenceCandidate `json:"candidates"`
}
type Contract struct {
	Historical struct {
		Prototype string `json:"prototype"`
		Semantics string `json:"input_semantics"`
	} `json:"historical_contract"`
}
type Reference struct {
	Schema     string `json:"schema"`
	References struct {
		Parents   []ReferenceParent `json:"parents"`
		Contracts []Contract        `json:"contracts"`
		Groups    []Group           `json:"historical_groups"`
	} `json:"references"`
}
type CandidateKey struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
}
type Index struct {
	Claim      string        `json:"claim_id"`
	Kind       string        `json:"existing59_entry_kind"`
	SHA        string        `json:"record_sha256"`
	State      string        `json:"state"`
	Axis       string        `json:"axis"`
	Additional bool          `json:"counts_as_additional_original59_entry"`
	ParentID   string        `json:"parent_id"`
	Prototype  string        `json:"prototype"`
	Candidate  *CandidateKey `json:"candidate"`
}
type Overlay struct {
	Schema          string            `json:"schema"`
	Prototype       string            `json:"prototype"`
	Records         []json.RawMessage `json:"review_records"`
	Mapped          []Index           `json:"selected_existing59_entries"`
	Supplement      []Index           `json:"supplementary_observation_and_negative_axis_records"`
	ShardSupplement []Index           `json:"supplementary_axis_records"`
}
type Shard struct {
	Path      string `json:"path"`
	Bytes     int64  `json:"bytes"`
	SHA       string `json:"sha256"`
	Prototype string `json:"prototype"`
}
type Manifest struct {
	Schema     string  `json:"schema"`
	Shards     []Shard `json:"family_shards"`
	Mapped     []Index `json:"selected_existing59_entries"`
	Supplement []Index `json:"supplementary_observation_and_negative_axis_records"`
}
type Record struct {
	Schema          string        `json:"schema"`
	Claim           string        `json:"claim_id"`
	Kind            string        `json:"record_kind"`
	Axis            string        `json:"axis"`
	State           string        `json:"state"`
	SHA             string        `json:"review_record_sha256"`
	ParentIndex     *int          `json:"original_parent_index"`
	ParentID        *string       `json:"parent_id"`
	Candidate       *CandidateKey `json:"candidate_index_and_id_when_applicable"`
	Prototype       string        `json:"prototype"`
	CaptionFidelity bool          `json:"caption_fidelity_assessment"`
	Approval        bool          `json:"aggregate_candidate_approval"`
	SourceID        string        `json:"source_id_provenance_only"`
	Text            struct {
		Request  *TextRef `json:"request"`
		Caption  *TextRef `json:"caption"`
		Contract *TextRef `json:"contract_semantics"`
	} `json:"request_and_caption_pointer_sha_bytes"`
	Range   []int  `json:"exact_text_byte_range"`
	Quote   string `json:"verbatim_claim"`
	Binding struct {
		Root string `json:"selected_root"`
	} `json:"official60_binding_artifact_sha256_and_source_refs"`
}

type Witness struct {
	Claim     string `json:"claim_id"`
	Kind      string `json:"entry_kind"`
	State     string `json:"state"`
	RecordSHA string `json:"record_sha256"`
	Artifact  Pin    `json:"artifact"`
	Pointer   string `json:"json_pointer"`
}
type CandidateOutput struct {
	Index       int      `json:"candidate_index"`
	ID          string   `json:"candidate_id"`
	SourceID    string   `json:"source_id_provenance_only"`
	Caption     TextRef  `json:"caption"`
	Label       *int     `json:"original_nullable_label"`
	Mask        bool     `json:"proposed_loss_mask"`
	Reasons     []string `json:"mask_reasons"`
	Closure     Witness  `json:"source_closure_review"`
	Fidelity    Witness  `json:"source_fidelity_review"`
	Coverage    Witness  `json:"request_coverage_review"`
	Observation Witness  `json:"observation_fields_review"`
	Negative    Witness  `json:"explicit_negative_boundaries_review"`
}
type ParentOutput struct {
	Index          int               `json:"original_parent_index"`
	ID             string            `json:"parent_id"`
	Cohort         string            `json:"cohort"`
	Prototype      string            `json:"prototype"`
	Group          int               `json:"whole_group_id"`
	State          string            `json:"original_truth_state"`
	Acceptable     []int             `json:"original_acceptable_candidate_indices"`
	Request        TextRef           `json:"request"`
	RequestReview  Witness           `json:"request_contract_review"`
	ContractReview Witness           `json:"contract_observation_review"`
	Candidates     []CandidateOutput `json:"candidates"`
}
type GroupOutput struct {
	ID                int      `json:"whole_group_id"`
	Parents           []string `json:"original_parent_ids"`
	Candidates        int      `json:"original_candidates"`
	KnownCandidates   int      `json:"known_candidates"`
	UnknownCandidates int      `json:"unknown_candidates"`
	EligiblePositive  int      `json:"eligible_positive_candidates"`
	EligibleNegative  int      `json:"eligible_negative_candidates"`
	MaskedKnown       int      `json:"masked_known_candidates"`
}
type Output struct {
	Schema                  string         `json:"schema"`
	PlanSHA                 string         `json:"plan_sha256"`
	ProposedSupervisionOnly bool           `json:"proposed_supervision_only"`
	TrainingReady           bool           `json:"training_ready"`
	DiversityCleared        bool           `json:"authoring_diversity_cleared"`
	OriginalEntries         int            `json:"original_entries_joined"`
	SupplementaryEntries    int            `json:"supplementary_entries_joined"`
	KnownOrNoAnswerParents  int            `json:"known_or_no_answer_parents"`
	UnknownParents          int            `json:"unknown_parents"`
	OriginalCandidates      int            `json:"original_candidates"`
	Inputs                  []Pin          `json:"input_artifacts"`
	Parents                 []ParentOutput `json:"parents"`
	Groups                  []GroupOutput  `json:"whole_groups"`
}

func sameInts(a, b []int) bool {
	return len(a) == len(b) && reflect.DeepEqual(append([]int{}, a...), append([]int{}, b...))
}
func sameRef(r TextRef, p Pin, pointer, text string) bool {
	if r.Path != "" && r.Artifact != nil {
		return false
	}
	file := r.Path
	if r.Artifact != nil {
		if *r.Artifact != p {
			return false
		}
		file = r.Artifact.Path
	}
	return file == p.Path && r.Pointer == pointer && r.Bytes == len(text) && r.SHA == sha([]byte(text))
}
func validState(s string) bool {
	for _, x := range [5]string{consistent, contradicts, "omits_required_scope", "unsupported_or_uncertain", "pending"} {
		if s == x {
			return true
		}
	}
	return false
}
func contractIndex(cs []Contract, prototype string) int {
	for i, c := range cs {
		if c.Historical.Prototype == prototype {
			return i
		}
	}
	return -1
}
func parentIndex(ps []ReferenceParent, id string) int {
	for i, p := range ps {
		if p.ID == id {
			return i
		}
	}
	return -1
}
func truthIndex(ts []Truth, id string) int {
	for i, t := range ts {
		if t.ID == id {
			return i
		}
	}
	return -1
}
func contains(xs []int, x int) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

type slot struct {
	parent, candidate, contract int
	kind                        string
	w                           Witness
	seen                        bool
}

func makeSlots(ref Reference) []slot {
	var ss []slot
	for i, p := range ref.References.Parents {
		ss = append(ss, slot{parent: i, candidate: -1, contract: -1, kind: "request_contract"})
		for c := range p.Candidates {
			for _, kind := range [5]string{"candidate_closure", "candidate_fidelity", "candidate_coverage", "observation_fields", "explicit_negative_boundaries"} {
				ss = append(ss, slot{parent: i, candidate: c, contract: -1, kind: kind})
			}
		}
	}
	for i := range ref.References.Contracts {
		ss = append(ss, slot{parent: -1, candidate: -1, contract: i, kind: "contract_observation"})
	}
	return ss
}
func getSlot(ss []slot, pi, ci, ki int, kind string) int {
	for i, s := range ss {
		if s.parent == pi && s.candidate == ci && s.contract == ki && s.kind == kind {
			return i
		}
	}
	return -1
}

// Existing digest recipe requires sorted JSON object encoding. A per-record
// map is used only for that canonical hash; joins/state/counts use slices.
func recordDigest(raw json.RawMessage) (string, error) {
	var object map[string]any
	if json.Unmarshal(raw, &object) != nil {
		return "", fail("record_digest")
	}
	if _, ok := object["review_record_sha256"]; !ok {
		return "", fail("record_digest")
	}
	delete(object, "review_record_sha256")
	b, e := json.MarshalIndent(object, "", "  ")
	if e != nil {
		return "", fail("record_digest")
	}
	return sha(append(b, '\n')), nil
}

func bindOriginal(cfg Config, probes [2]Probes, saved Saved, ref Reference) error {
	if probes[0].Schema != "riido-behavior-probes-v1" || probes[1].Schema != "riido-typed-behavior-probes-v2" || saved.Schema != "riido-typed-truth-development-result-v1" || ref.Schema != "riido-caption-reference-preparation-record-59-v1" || saved.InputSHA != cfg.Pins[1].SHA || saved.LegacySHA != cfg.Pins[0].SHA {
		return fail("original_schema_pin")
	}
	ps := ref.References.Parents
	cs := ref.References.Contracts
	groups := saved.Combined.Groups
	if len(ps) != cfg.Expected.Parents || len(groups) != cfg.Expected.Groups || len(probes[0].Parents)+len(probes[1].Parents) != len(ps) || len(saved.Legacy.Outcomes)+len(saved.Typed.Outcomes) != len(ps) || len(cs) > 64 || !reflect.DeepEqual(groups, ref.References.Groups) {
		return fail("original_cardinality")
	}
	for i, c := range cs {
		if c.Historical.Prototype == "" || c.Historical.Semantics == "" {
			return fail("contract")
		}
		for j := 0; j < i; j++ {
			if cs[j].Historical.Prototype == c.Historical.Prototype {
				return fail("contract_duplicate")
			}
		}
	}
	known, unknown, candidates := 0, 0, 0
	for i, p := range ps {
		if p.Index != i || p.ID == "" || p.Group < 0 || parentIndex(ps, p.ID) != i || contractIndex(cs, p.Prototype) < 0 || len(p.Candidates) < 1 || len(p.Candidates) > 8 {
			return fail("parent_binding")
		}
		cohort, local := 0, i
		if i >= len(probes[0].Parents) {
			cohort, local = 1, i-len(probes[0].Parents)
		}
		raw := probes[cohort].Parents[local]
		wantCohort := "legacy"
		truth := saved.Legacy.Outcomes
		if cohort == 1 {
			wantCohort = "typed"
			truth = saved.Typed.Outcomes
		}
		ti := truthIndex(truth, p.ID)
		if ti < 0 {
			return fail("truth_missing")
		}
		t := truth[ti]
		if p.Cohort != wantCohort || raw.ID != p.ID || raw.Prototype != p.Prototype || raw.ContractID != p.ContractID || len(raw.Candidates) != len(p.Candidates) || !sameRef(p.Request, cfg.Pins[cohort], fmt.Sprintf("/parents/%d/request", local), raw.Request) || p.State != t.State || !sameInts(p.Acceptable, t.Acceptable) || len(t.Candidates) != len(p.Candidates) {
			return fail("original_parent_mismatch")
		}
		if p.State == "unknown" {
			unknown++
			if len(p.Acceptable) != 0 {
				return fail("unknown_acceptable")
			}
		} else if p.State == "known" || p.State == "no_answer" {
			known++
			if (p.State == "known") != (len(p.Acceptable) > 0) {
				return fail("truth_acceptable")
			}
		} else {
			return fail("truth_state")
		}
		var seen [8]bool
		for _, a := range p.Acceptable {
			if a < 0 || a >= len(p.Candidates) || seen[a] {
				return fail("acceptable")
			}
			seen[a] = true
		}
		for c, rc := range p.Candidates {
			r := raw.Candidates[c]
			if rc.Index != c || rc.ID != r.ID || rc.SourceID != r.SourceID || rc.CodeSHA != r.CodeSHA || rc.BundleSHA != r.BundleSHA || rc.Outcome != t.Candidates[c] || rc.Outcome.ID != rc.ID || !sameRef(rc.Caption, cfg.Pins[cohort], fmt.Sprintf("/parents/%d/candidates/%d/text", local, c), r.Text) {
				return fail("candidate_binding")
			}
			for j := 0; j < c; j++ {
				if p.Candidates[j].ID == rc.ID {
					return fail("candidate_duplicate")
				}
			}
			if p.State != "unknown" {
				want := "rejected"
				if seen[c] {
					want = "acceptable"
				}
				if t.Candidates[c].State != want {
					return fail("candidate_truth")
				}
			}
			candidates++
		}
	}
	if known != cfg.Expected.KnownOrNoAnswer || unknown != cfg.Expected.Unknown || candidates != cfg.Expected.Candidates {
		return fail("truth_counts")
	}
	var assigned []bool
	assigned = make([]bool, len(ps))
	for gi, g := range groups {
		if g.ID < 0 || len(g.Parents) == 0 {
			return fail("group")
		}
		for j := 0; j < gi; j++ {
			if groups[j].ID == g.ID {
				return fail("group_duplicate")
			}
		}
		for _, id := range g.Parents {
			pi := parentIndex(ps, id)
			if pi < 0 || assigned[pi] || ps[pi].Group != g.ID {
				return fail("group_membership")
			}
			assigned[pi] = true
		}
	}
	for _, a := range assigned {
		if !a {
			return fail("group_missing")
		}
	}
	return nil
}

func recordKey(r Record, ref Reference) (pi, ci, ki int, kind, text string, err error) {
	pi, ci, ki = -1, -1, -1
	if r.Schema != "riido-caption-content-review-record-60-v1" || r.Claim == "" || !validState(r.State) || !hexSHA(r.SHA) || r.Approval {
		return pi, ci, ki, "", "", fail("record_contract")
	}
	if r.Kind == "contract_observation" {
		if r.ParentID != nil || r.ParentIndex != nil || r.Candidate != nil || r.Text.Request != nil || r.Text.Caption != nil || r.Text.Contract == nil || r.Axis != "observation_fields" || r.CaptionFidelity {
			return pi, ci, ki, "", "", fail("contract_record")
		}
		for i, c := range ref.References.Contracts {
			if r.Text.Contract.ContractPointer == fmt.Sprintf("/references/contracts/%d", i) {
				ki = i
				text = c.Historical.Semantics
				break
			}
		}
		if ki < 0 || r.Prototype != "" && r.Prototype != ref.References.Contracts[ki].Historical.Prototype || r.Claim != "contract-observation/"+ref.References.Contracts[ki].Historical.Prototype || r.Text.Contract.SHA != sha([]byte(text)) || r.Text.Contract.Bytes != len(text) {
			return pi, ci, ki, "", "", fail("contract_reference")
		}
		return pi, ci, ki, "contract_observation", text, nil
	}
	if r.ParentID == nil || r.ParentIndex == nil || *r.ParentIndex < 0 || *r.ParentIndex >= len(ref.References.Parents) {
		return pi, ci, ki, "", "", fail("record_parent")
	}
	pi = *r.ParentIndex
	p := ref.References.Parents[pi]
	if *r.ParentID != p.ID || r.Prototype != "" && r.Prototype != p.Prototype || r.Text.Request == nil || r.Text.Contract != nil {
		return pi, ci, ki, "", "", fail("record_parent")
	}
	if r.Kind == "request_contract" {
		if r.Candidate != nil || r.Text.Caption != nil || r.Axis != "request_contract_coverage" || r.CaptionFidelity {
			return pi, ci, ki, "", "", fail("request_record")
		}
		return pi, ci, ki, "request_contract", "", nil
	}
	if r.Candidate == nil || r.Candidate.Index < 0 || r.Candidate.Index >= len(p.Candidates) || r.Text.Caption == nil {
		return pi, ci, ki, "", "", fail("record_candidate")
	}
	ci = r.Candidate.Index
	c := p.Candidates[ci]
	if r.Candidate.ID != c.ID || r.SourceID != c.SourceID || r.Binding.Root != c.SourceID {
		return pi, ci, ki, "", "", fail("record_source_binding")
	}
	if r.Kind == "source_closure_only" && r.Axis == "source_fidelity" && !r.CaptionFidelity {
		return pi, ci, ki, "candidate_closure", "", nil
	}
	if r.Kind == "candidate_axis" {
		switch r.Axis {
		case "source_fidelity":
			if r.CaptionFidelity {
				return pi, ci, ki, "candidate_fidelity", "", nil
			}
		case "request_contract_coverage":
			if !r.CaptionFidelity {
				return pi, ci, ki, "candidate_coverage", "", nil
			}
		case "observation_fields", "explicit_negative_boundaries":
			if !r.CaptionFidelity {
				return pi, ci, ki, r.Axis, "", nil
			}
		}
	}
	return pi, ci, ki, "", "", fail("record_kind_axis")
}

func indexJoin(o Overlay) ([]Index, error) {
	ss := o.Supplement
	if len(o.ShardSupplement) > 0 {
		if len(ss) > 0 {
			return nil, fail("supplement_shape")
		}
		ss = o.ShardSupplement
	}
	all := append(append([]Index(nil), o.Mapped...), ss...)
	if len(all) != len(o.Records) {
		return nil, fail("record_index_cardinality")
	}
	for i, x := range all {
		if x.Claim == "" || !hexSHA(x.SHA) || !validState(x.State) {
			return nil, fail("index")
		}
		for j := 0; j < i; j++ {
			if all[j].Claim == x.Claim || all[j].SHA == x.SHA {
				return nil, fail("index_duplicate")
			}
		}
		if i < len(o.Mapped) {
			switch x.Kind {
			case "request_contract", "candidate_closure", "candidate_fidelity", "candidate_coverage", "contract_observation":
			default:
				return nil, fail("mapped_kind")
			}
		} else if x.Additional || x.Kind != "" || x.Axis != "observation_fields" && x.Axis != "explicit_negative_boundaries" {
			return nil, fail("supplement_kind")
		}
	}
	return all, nil
}

func (l *loader) joinOverlay(pin Pin, o Overlay, probes [2]Probes, ref Reference, slots []slot) error {
	if len(o.Records) > len(slots) || len(o.Mapped) > len(slots) || len(o.Supplement) > len(slots) || len(o.ShardSupplement) > len(slots) {
		return fail("overlay_bounds")
	}
	indexes, e := indexJoin(o)
	if e != nil {
		return e
	}
	used := make([]bool, len(indexes))
	for ri, raw := range o.Records {
		var r Record
		if e := strictJSON(raw, &r); e != nil {
			return e
		}
		digest, e := recordDigest(raw)
		if e != nil || digest != r.SHA {
			return fail("record_digest")
		}
		ix := -1
		for i, x := range indexes {
			if x.Claim == r.Claim {
				ix = i
				break
			}
		}
		if ix < 0 || used[ix] || indexes[ix].SHA != r.SHA || indexes[ix].State != r.State {
			return fail("record_index")
		}
		used[ix] = true
		pi, ci, ki, kind, text, e := recordKey(r, ref)
		if e != nil {
			return e
		}
		if indexes[ix].Kind != "" && indexes[ix].Kind != kind || indexes[ix].Kind == "" && indexes[ix].Axis != kind {
			return fail("record_index_kind")
		}
		x := indexes[ix]
		if x.Axis != "" && x.Axis != r.Axis || x.ParentID != "" && (r.ParentID == nil || x.ParentID != *r.ParentID) || x.Candidate != nil && !reflect.DeepEqual(x.Candidate, r.Candidate) {
			return fail("index_binding")
		}
		prototype := ""
		if pi >= 0 {
			prototype = ref.References.Parents[pi].Prototype
		} else {
			prototype = ref.References.Contracts[ki].Historical.Prototype
		}
		if x.Prototype != "" && x.Prototype != prototype {
			return fail("index_binding")
		}
		if o.Prototype != "" {
			prototype := ""
			if pi >= 0 {
				prototype = ref.References.Parents[pi].Prototype
			} else {
				prototype = ref.References.Contracts[ki].Historical.Prototype
			}
			if prototype != o.Prototype {
				return fail("shard_prototype")
			}
		}
		if pi >= 0 {
			cohort, local := 0, pi
			if pi >= len(probes[0].Parents) {
				cohort, local = 1, pi-len(probes[0].Parents)
			}
			p := probes[cohort].Parents[local]
			if !sameRef(*r.Text.Request, l.cfg.Pins[cohort], fmt.Sprintf("/parents/%d/request", local), p.Request) {
				return fail("record_request_ref")
			}
			text = p.Request
			if ci >= 0 {
				text = p.Candidates[ci].Text
				if !sameRef(*r.Text.Caption, l.cfg.Pins[cohort], fmt.Sprintf("/parents/%d/candidates/%d/text", local, ci), text) {
					return fail("record_caption_ref")
				}
			}
		}
		if len(r.Range) != 2 || r.Range[0] < 0 || r.Range[0] > r.Range[1] || r.Range[1] > len(text) || !utf8.ValidString(text[r.Range[0]:r.Range[1]]) || r.Quote != text[r.Range[0]:r.Range[1]] {
			return fail("record_quote")
		}
		si := getSlot(slots, pi, ci, ki, kind)
		if si < 0 || slots[si].seen {
			return fail("record_slot_duplicate")
		}
		slots[si].seen = true
		slots[si].w = Witness{r.Claim, kind, r.State, r.SHA, pin, fmt.Sprintf("/review_records/%d", ri)}
		l.ledger.ReviewRecordsVerified++
		if kind == "observation_fields" || kind == "explicit_negative_boundaries" {
			l.ledger.SupplementalSlotsJoined++
		} else {
			l.ledger.OriginalSlotsJoined++
		}
	}
	return nil
}

func compareManifest(m Manifest, overlays []Overlay) error {
	var mapped, supp []Index
	for _, o := range overlays {
		mapped = append(mapped, o.Mapped...)
		supp = append(supp, o.ShardSupplement...)
	}
	for _, pair := range [][2][]Index{{m.Mapped, mapped}, {m.Supplement, supp}} {
		a, b := append([]Index(nil), pair[0]...), append([]Index(nil), pair[1]...)
		sort.Slice(a, func(i, j int) bool { return a[i].Claim < a[j].Claim })
		sort.Slice(b, func(i, j int) bool { return b[i].Claim < b[j].Claim })
		if !reflect.DeepEqual(a, b) {
			return fail("manifest_index")
		}
	}
	return nil
}

func mask(label *int, request, contract, closure, fidelity, coverage string) (bool, []string) {
	reasons := make([]string, 0, 6)
	if label == nil {
		reasons = append(reasons, "original_unknown_no_loss_label")
	}
	for _, x := range [][2]string{{"request_contract", request}, {"contract_observation", contract}, {"source_closure", closure}, {"source_fidelity", fidelity}} {
		if x[1] != consistent {
			reasons = append(reasons, x[0]+":"+x[1])
		}
	}
	if label != nil {
		want := contradicts
		if *label == 1 {
			want = consistent
		}
		if coverage != want {
			reasons = append(reasons, "candidate_coverage_label_mismatch:"+coverage)
		}
	}
	return len(reasons) == 0, reasons
}

func output(l *loader, ref Reference, slots []slot) Output {
	out := Output{Schema: "riido-proposed-supervision-scope-mask-67-v1", PlanSHA: planSHA, ProposedSupervisionOnly: true, OriginalEntries: l.ledger.OriginalSlotsJoined, SupplementaryEntries: l.ledger.SupplementalSlotsJoined, Inputs: append([]Pin(nil), l.pins...)}
	witness := func(pi, ci, ki int, kind string) Witness { return slots[getSlot(slots, pi, ci, ki, kind)].w }
	for pi, p := range ref.References.Parents {
		v := ParentOutput{Index: p.Index, ID: p.ID, Cohort: p.Cohort, Prototype: p.Prototype, Group: p.Group, State: p.State, Acceptable: append([]int{}, p.Acceptable...), Request: p.Request, RequestReview: witness(pi, -1, -1, "request_contract"), ContractReview: witness(-1, -1, contractIndex(ref.References.Contracts, p.Prototype), "contract_observation")}
		if p.State == "unknown" {
			out.UnknownParents++
		} else {
			out.KnownOrNoAnswerParents++
		}
		for ci, c := range p.Candidates {
			x := CandidateOutput{Index: ci, ID: c.ID, SourceID: c.SourceID, Caption: c.Caption, Closure: witness(pi, ci, -1, "candidate_closure"), Fidelity: witness(pi, ci, -1, "candidate_fidelity"), Coverage: witness(pi, ci, -1, "candidate_coverage"), Observation: witness(pi, ci, -1, "observation_fields"), Negative: witness(pi, ci, -1, "explicit_negative_boundaries")}
			if p.State != "unknown" {
				y := 0
				if contains(p.Acceptable, ci) {
					y = 1
				}
				x.Label = &y
			}
			x.Mask, x.Reasons = mask(x.Label, v.RequestReview.State, v.ContractReview.State, x.Closure.State, x.Fidelity.State, x.Coverage.State)
			v.Candidates = append(v.Candidates, x)
			out.OriginalCandidates++
		}
		out.Parents = append(out.Parents, v)
	}
	for _, g := range ref.References.Groups {
		v := GroupOutput{ID: g.ID, Parents: append([]string(nil), g.Parents...)}
		for _, id := range g.Parents {
			p := out.Parents[parentIndex(ref.References.Parents, id)]
			for _, c := range p.Candidates {
				v.Candidates++
				if c.Label == nil {
					v.UnknownCandidates++
				} else {
					v.KnownCandidates++
					if c.Mask {
						if *c.Label == 1 {
							v.EligiblePositive++
						} else {
							v.EligibleNegative++
						}
					} else {
						v.MaskedKnown++
					}
				}
			}
		}
		out.Groups = append(out.Groups, v)
	}
	return out
}

func load(root string, cfg Config) (Output, Ledger, error) {
	var zero Output
	var initial Ledger
	if cfg.MaxFile <= 0 || cfg.MaxFile > maxFileBytes || cfg.MaxTotal <= 0 || cfg.MaxTotal > maxTotalBytes || cfg.Expected.Parents < 1 || cfg.Expected.Parents > 256 || cfg.Expected.Candidates < 1 || cfg.Expected.Candidates > 2048 || cfg.Expected.Groups < 1 || cfg.Expected.Groups > 256 || cfg.Expected.Original < 1 || cfg.Expected.Original > 12000 || cfg.Expected.Supplement < 0 || cfg.Expected.Supplement > 4096 {
		return zero, initial, fail("config")
	}
	r, e := os.OpenRoot(root)
	if e != nil {
		return zero, initial, fail("root")
	}
	defer r.Close()
	l := loader{root: r, cfg: cfg}
	stop := func(e error) (Output, Ledger, error) { return zero, l.ledger, e }
	var probes [2]Probes
	var saved Saved
	var ref Reference
	var sixty Overlay
	var manifest Manifest
	if e = l.read(cfg.Pins[0], &probes[0]); e != nil {
		return stop(e)
	}
	if e = l.read(cfg.Pins[1], &probes[1]); e != nil {
		return stop(e)
	}
	if e = l.read(cfg.Pins[2], &saved); e != nil {
		return stop(e)
	}
	if e = l.read(cfg.Pins[3], &ref); e != nil {
		return stop(e)
	}
	if e = l.read(cfg.Pins[4], &sixty); e != nil {
		return stop(e)
	}
	if e = l.read(cfg.Pins[5], &manifest); e != nil {
		return stop(e)
	}
	if len(manifest.Shards) < 1 || len(manifest.Shards) > 64 || manifest.Schema != "riido-caption-content-scoped-review-manifest-61-v1" {
		return stop(fail("manifest"))
	}
	if sixty.Schema != "riido-caption-content-scoped-review-60-v1" {
		return stop(fail("overlay_schema"))
	}
	if e = bindOriginal(cfg, probes, saved, ref); e != nil {
		return stop(e)
	}
	l.ledger.OriginalBindingsVerified = len(ref.References.Parents)
	slots := makeSlots(ref)
	if len(slots) != cfg.Expected.Original+cfg.Expected.Supplement {
		return stop(fail("expected_slots"))
	}
	if e = l.joinOverlay(cfg.Pins[4], sixty, probes, ref, slots); e != nil {
		return stop(e)
	}
	var overlays []Overlay
	for i, s := range manifest.Shards {
		if s.Path != path.Base(s.Path) || !safePath(s.Path) || s.Prototype == "" || s.Path != "content-review-family-61-"+s.Prototype+".json" {
			return stop(fail("shard_path"))
		}
		for j := 0; j < i; j++ {
			if manifest.Shards[j].Path == s.Path || manifest.Shards[j].Prototype == s.Prototype {
				return stop(fail("shard_duplicate"))
			}
		}
		p := Pin{basePath + s.Path, s.Bytes, s.SHA}
		var o Overlay
		if e = l.read(p, &o); e != nil {
			return stop(e)
		}
		if o.Schema != "riido-caption-content-family-review-61-v1" || o.Prototype != s.Prototype {
			return stop(fail("shard_prototype"))
		}
		if e = l.joinOverlay(p, o, probes, ref, slots); e != nil {
			return stop(e)
		}
		overlays = append(overlays, o)
	}
	if e = compareManifest(manifest, overlays); e != nil {
		return stop(e)
	}
	for _, s := range slots {
		if !s.seen {
			return stop(fail("missing_review_slot"))
		}
	}
	if l.ledger.OriginalSlotsJoined != cfg.Expected.Original || l.ledger.SupplementalSlotsJoined != cfg.Expected.Supplement {
		return stop(fail("slot_counts"))
	}
	return output(&l, ref, slots), l.ledger, nil
}
