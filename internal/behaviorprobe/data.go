// Package behaviorprobe prepares authored finite development microcontracts.
// Code truth and natural-language fidelity are separate checks; no data here is
// a protected-final sample or automatically eligible for fitting.
package behaviorprobe

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
)

const (
	DatasetSchema = "riido-behavior-probes-v1"
	DatasetOrigin = "original_authored_go_microcontracts_56_apache2_development_only"
	MaxDataBytes  = 64 << 20
	MaxParents    = 120
)

// Embedding our own authored source pins executable function declarations and
// literal tables without reading a mutable runtime source path. This source is
// used only by offline preparation/audit, never as a text-scoring feature.
//
//go:embed data.go
var authoredSource string

type Dataset struct {
	Schema  string   `json:"schema"`
	Origin  string   `json:"origin"`
	Parents []Parent `json:"parents"`
}

type Parent struct {
	ID         string            `json:"id"`
	Prototype  string            `json:"prototype"`
	ContractID string            `json:"contract_id"`
	Request    string            `json:"request"`
	Candidates []SourceCandidate `json:"candidates"`
}

// Source identity and hashes are provenance, never predictive features.
type SourceCandidate struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	SourceID   string `json:"source_id"`
	CodeSHA256 string `json:"code_sha256"`
}

// InputText is the only model/baseline projection. Index order is preserved;
// target, source, prototype, candidate ID and role fields cannot enter it.
type InputText struct {
	Request    string
	Candidates []string
}

func FeatureInputs(d Dataset) []InputText {
	out := make([]InputText, len(d.Parents))
	for i, p := range d.Parents {
		out[i].Request = p.Request
		out[i].Candidates = make([]string, len(p.Candidates))
		for j, c := range p.Candidates {
			out[i].Candidates[j] = c.Text
		}
	}
	return out
}

func Load(path string) (Dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return Dataset{}, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, MaxDataBytes+1))
	if err != nil {
		return Dataset{}, err
	}
	if len(b) > MaxDataBytes {
		return Dataset{}, fmt.Errorf("dataset byte limit")
	}
	if !utf8.Valid(b) {
		return Dataset{}, fmt.Errorf("invalid dataset UTF-8")
	}
	var d Dataset
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return Dataset{}, err
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return Dataset{}, fmt.Errorf("extra dataset JSON")
	}
	if err := validateDataset(d); err != nil {
		return Dataset{}, err
	}
	return d, nil
}

func boundedText(s string) bool {
	if s == "" || !utf8.ValidString(s) || len(s) > 512 {
		return false
	}
	n := lexicalhint.NormalizeText(s)
	return n != "" && len(n) <= 512 && len(strings.Fields(n)) <= 32
}

func validID(s string) bool {
	if len(s) == 0 || len(s) > 96 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func validateDataset(d Dataset) error {
	if d.Schema != DatasetSchema || d.Origin != DatasetOrigin {
		return fmt.Errorf("unrecognized dataset schema or origin")
	}
	if len(d.Parents) < 1 || len(d.Parents) > MaxParents {
		return fmt.Errorf("parent count limit")
	}
	sources, err := SourcePins()
	if err != nil {
		return err
	}
	for i, p := range d.Parents {
		if !validID(p.ID) || !boundedText(p.Request) || len(p.Candidates) < 1 || len(p.Candidates) > 8 {
			return fmt.Errorf("invalid parent %d", i)
		}
		for j := 0; j < i; j++ {
			if d.Parents[j].ID == p.ID {
				return fmt.Errorf("duplicate parent ID")
			}
		}
		if contractIndex(p.Prototype) < 0 || p.ContractID != p.Prototype+"-v1" && p.ContractID != "unknown-ambiguous" {
			return fmt.Errorf("unsupported contract identity")
		}
		for j, c := range p.Candidates {
			if !validID(c.ID) || !boundedText(c.Text) {
				return fmt.Errorf("invalid candidate %d/%d", i, j)
			}
			for k := 0; k < j; k++ {
				if p.Candidates[k].ID == c.ID {
					return fmt.Errorf("duplicate candidate ID")
				}
			}
			s := slices.IndexFunc(sources, func(s SourcePin) bool { return s.ID == c.SourceID })
			if s < 0 || c.CodeSHA256 != sources[s].CodeSHA256 {
				return fmt.Errorf("unknown or changed source pin")
			}
		}
	}
	return nil
}

type SourcePin struct {
	ID                   string `json:"id"`
	Prototype            string `json:"prototype"`
	CoreTemplate         string `json:"core_template"`
	CodeSHA256           string `json:"code_sha256"`
	NormalizedCodeSHA256 string `json:"normalized_code_sha256"`
}

// Code hashes cover formatted complete declarations. Normalization erases local
// names and declaration names, preserving operators, literals and builtin names.
// Thus renamed negative candidates also connect their parent components.
func SourcePins() ([]SourcePin, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "authored.go", authoredSource, 0)
	if err != nil {
		return nil, err
	}
	out := make([]SourcePin, len(sourceCatalog))
	for i, s := range sourceCatalog {
		var fn *ast.FuncDecl
		for _, d := range f.Decls {
			if x, ok := d.(*ast.FuncDecl); ok && x.Name.Name == s.function {
				fn = x
				break
			}
		}
		if fn == nil {
			return nil, fmt.Errorf("missing authored function")
		}
		var b bytes.Buffer
		if err := format.Node(&b, fset, fn); err != nil {
			return nil, err
		}
		code := b.Bytes()
		out[i] = SourcePin{s.id, s.prototype, s.core, digest(code), normalizedCodeSHA(code)}
	}
	return out, nil
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func normalizedCodeSHA(code []byte) string {
	var scan scanner.Scanner
	fset := token.NewFileSet()
	scan.Init(fset.AddFile("code.go", -1, len(code)), code, nil, 0)
	var out strings.Builder
	for {
		_, tok, lit := scan.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.IDENT && lit != "append" && lit != "len" && lit != "make" && lit != "int" && lit != "bool" && lit != "true" && lit != "false" && lit != "nil" {
			lit = "_"
		}
		fmt.Fprintf(&out, "%d:%s;", tok, lit)
	}
	return digest([]byte(out.String()))
}

type CandidateTruth struct {
	CandidateID    string `json:"candidate_id"`
	State          string `json:"state"`
	VectorsChecked int    `json:"vectors_checked"`
	FailedVectors  int    `json:"failed_vectors"`
}

type Outcome struct {
	ParentID   string           `json:"parent_id"`
	State      string           `json:"state"`
	Acceptable []int            `json:"acceptable_candidate_indices"`
	Candidates []CandidateTruth `json:"candidate_truth"`
	Reason     string           `json:"reason,omitempty"`
}

type ControlTruth struct {
	SourceID         string `json:"source_id"`
	ExpectedControl  string `json:"expected_control"`
	VectorsChecked   int    `json:"vectors_checked"`
	FailedVectors    int    `json:"failed_vectors"`
	TruthTableSHA256 string `json:"truth_table_sha256"`
}

type ContractEvidence struct {
	Prototype        string `json:"prototype"`
	CoreTemplate     string `json:"core_template"`
	InputSemantics   string `json:"input_semantics"`
	Vectors          int    `json:"literal_vectors"`
	TruthTableSHA256 string `json:"truth_table_sha256"`
}

type GroupEdge struct {
	Left    string   `json:"left"`
	Right   string   `json:"right"`
	Reasons []string `json:"reasons"`
}

type Group struct {
	ID            int      `json:"id"`
	Parents       []string `json:"parents"`
	Prototypes    []string `json:"prototypes"`
	CoreTemplates []string `json:"core_templates"`
	Sources       []string `json:"sources"`
}

type Report struct {
	Schema                        string             `json:"schema"`
	DatasetSHA256                 string             `json:"typed_dataset_sha256"`
	SourceArtifactSHA256          string             `json:"source_artifact_sha256"`
	Parents                       int                `json:"parents"`
	Candidates                    int                `json:"candidates"`
	KnownParents                  int                `json:"known_parents"`
	UnknownParents                int                `json:"unknown_parents"`
	NoAnswerParents               int                `json:"no_answer_parents"`
	PrototypeCount                int                `json:"prototype_count"`
	ConnectedGroups               int                `json:"connected_groups"`
	MinimumGroups                 int                `json:"minimum_operational_groups"`
	MeetsGroupMinimum             bool               `json:"meets_group_minimum"`
	NaturalLanguageReviewRequired bool               `json:"natural_language_review_required"`
	TrainingReady                 bool               `json:"training_ready"`
	Partitioned                   bool               `json:"partitioned"`
	FinalEligible                 bool               `json:"final_eligible"`
	Outcomes                      []Outcome          `json:"outcomes"`
	Sources                       []SourcePin        `json:"sources"`
	Controls                      []ControlTruth     `json:"controls"`
	Contracts                     []ContractEvidence `json:"contracts"`
	Edges                         []GroupEdge        `json:"group_edges"`
	Groups                        []Group            `json:"groups"`
	Limitations                   []string           `json:"limitations"`
}

func Evaluate(d Dataset) (Report, error) {
	if err := validateDataset(d); err != nil {
		return Report{}, err
	}
	sources, err := SourcePins()
	if err != nil {
		return Report{}, err
	}
	b, err := json.Marshal(d)
	if err != nil {
		return Report{}, err
	}
	r := Report{Schema: "riido-behavior-probe-audit-v1", DatasetSHA256: digest(b), SourceArtifactSHA256: digest([]byte(authoredSource)), Parents: len(d.Parents), MinimumGroups: 15, NaturalLanguageReviewRequired: true, Sources: sources, Limitations: []string{
		"Finite authored vectors establish observed candidate behavior only, not general semantic correctness.",
		"All prototypes, paraphrases and controls share one author; connected components do not prove statistical independence.",
		"Unknown requirements have no boolean labels; known no-answer sets are distinct from unknown.",
		"No roles, partition, fitting, model calls or protected-final eligibility are created by this audit.",
	}}
	for _, c := range contracts {
		table, _ := json.Marshal(c.vectors)
		r.Contracts = append(r.Contracts, ContractEvidence{c.id, c.core, inputSemantics(c.id), len(c.vectors), digest(table)})
	}
	for _, s := range sourceCatalog {
		c := contracts[contractIndex(s.prototype)]
		checked, failed, unknown := check(s.run, c.vectors)
		if unknown || s.correct && failed != 0 || !s.correct && failed == 0 {
			return Report{}, fmt.Errorf("authored correct/wrong control does not match its independent table")
		}
		table, _ := json.Marshal(c.vectors)
		role := "wrong"
		if s.correct {
			role = "correct"
		}
		r.Controls = append(r.Controls, ControlTruth{s.id, role, checked, failed, digest(table)})
	}
	prototypes := []string{}
	for _, p := range d.Parents {
		prototypes = appendUnique(prototypes, p.Prototype)
		r.Candidates += len(p.Candidates)
		o := Outcome{ParentID: p.ID, State: "known", Acceptable: []int{}, Candidates: []CandidateTruth{}}
		if p.ContractID == "unknown-ambiguous" {
			o.State, o.Reason = "unknown", "request_policy_ambiguous_no_literal_oracle"
			for _, c := range p.Candidates {
				o.Candidates = append(o.Candidates, CandidateTruth{CandidateID: c.ID, State: "unknown"})
			}
			r.UnknownParents++
		} else {
			for i, c := range p.Candidates {
				s := sourceCatalog[slices.IndexFunc(sourceCatalog, func(s sourceDef) bool { return s.id == c.SourceID })]
				checked, failed, unknown := check(s.run, contracts[contractIndex(p.Prototype)].vectors)
				state := "rejected"
				if unknown {
					state, o.State, o.Reason = "unknown", "unknown", "candidate_check_failed"
				} else if failed == 0 {
					state = "acceptable"
					o.Acceptable = append(o.Acceptable, i)
				}
				o.Candidates = append(o.Candidates, CandidateTruth{c.ID, state, checked, failed})
			}
			if o.State == "unknown" {
				o.Acceptable = []int{}
				r.UnknownParents++
			} else if len(o.Acceptable) == 0 {
				o.State = "no_answer"
				r.NoAnswerParents++
				r.KnownParents++
			} else {
				r.KnownParents++
			}
		}
		r.Outcomes = append(r.Outcomes, o)
	}
	r.PrototypeCount = len(prototypes)
	r.Edges, r.Groups = groupAudit(d, sources)
	r.ConnectedGroups = len(r.Groups)
	r.MeetsGroupMinimum = r.ConnectedGroups >= r.MinimumGroups
	return r, nil
}

type vector struct {
	Input []int `json:"input"`
	Want  []int `json:"want"`
}
type contract struct {
	id, core string
	vectors  []vector
}

// The literal expected vectors below are independently authored, never computed
// by a candidate function. They deliberately include polarity, boundary and
// ordering counterexamples, including known no-answer control sets.
var contracts = []contract{
	{"retain-active", "boolean-retention", []vector{{[]int{0, 0}, []int{0}}, {[]int{0, 1}, []int{0}}, {[]int{1, 0}, []int{1}}, {[]int{1, 1}, []int{0}}}},
	{"allow-owner", "boolean-authorization", []vector{{[]int{0, 0, 0}, []int{0}}, {[]int{0, 0, 1}, []int{0}}, {[]int{0, 1, 0}, []int{0}}, {[]int{0, 1, 1}, []int{1}}, {[]int{1, 0, 0}, []int{1}}, {[]int{1, 0, 1}, []int{1}}, {[]int{1, 1, 0}, []int{1}}, {[]int{1, 1, 1}, []int{1}}}},
	{"closed-window", "closed-boundary", []vector{{[]int{-1}, []int{0}}, {[]int{2}, []int{0}}, {[]int{3}, []int{1}}, {[]int{4}, []int{1}}, {[]int{7}, []int{1}}, {[]int{8}, []int{0}}}},
	{"quota-total", "sum-before-comparison", []vector{{[]int{0, 0, 0}, []int{1}}, {[]int{2, 3, 5}, []int{1}}, {[]int{2, 4, 5}, []int{0}}, {[]int{6, 0, 5}, []int{0}}, {[]int{0, 6, 5}, []int{0}}, {[]int{1, 1, 5}, []int{1}}}},
	{"stable-odd", "stable-filter", []vector{{[]int{}, []int{}}, {[]int{2, 1, 3, 4, 1}, []int{1, 3, 1}}, {[]int{-3, -2, -1, 0, 5}, []int{-3, -1, 5}}, {[]int{9, 7, 5}, []int{9, 7, 5}}, {[]int{0, 2, 4}, []int{}}}},
	{"remove-first", "single-removal", []vector{{[]int{}, []int{}}, {[]int{1, 2, 3, 2}, []int{1, 3, 2}}, {[]int{2, 2, 2}, []int{2, 2}}, {[]int{3, 1}, []int{3, 1}}, {[]int{2}, []int{}}}},
	{"compact-runs", "adjacent-compaction", []vector{{[]int{}, []int{}}, {[]int{1, 1, 2, 2, 1}, []int{1, 2, 1}}, {[]int{3, 3, 3}, []int{3}}, {[]int{1, 2, 1}, []int{1, 2, 1}}, {[]int{-1, -1, 0, 0}, []int{-1, 0}}}},
	{"rotate-left", "cyclic-position", []vector{{[]int{}, []int{}}, {[]int{1}, []int{1}}, {[]int{1, 2}, []int{2, 1}}, {[]int{1, 2, 3}, []int{2, 3, 1}}, {[]int{8, 5, 3, 1}, []int{5, 3, 1, 8}}}},
	{"nondecreasing", "adjacent-order", []vector{{[]int{}, []int{1}}, {[]int{2}, []int{1}}, {[]int{1, 1, 2}, []int{1}}, {[]int{1, 3, 2, 4}, []int{0}}, {[]int{3, 2, 1}, []int{0}}, {[]int{-3, -2, 0}, []int{1}}}},
	{"clamp-window", "closed-boundary", []vector{{[]int{2, 3, 7}, []int{3}}, {[]int{3, 3, 7}, []int{3}}, {[]int{5, 3, 7}, []int{5}}, {[]int{7, 3, 7}, []int{7}}, {[]int{8, 3, 7}, []int{7}}, {[]int{-8, -5, -1}, []int{-5}}}},
	{"prefix-balance", "prefix-state", []vector{{[]int{}, []int{1}}, {[]int{1, -1}, []int{1}}, {[]int{-1, 1}, []int{0}}, {[]int{1, 1, -1, -1}, []int{1}}, {[]int{1, -1, -1, 1}, []int{0}}, {[]int{1, 1, -1}, []int{0}}, {[]int{-1}, []int{0}}}},
	{"transform-order", "ordered-arithmetic", []vector{{[]int{0}, []int{-6}}, {[]int{3}, []int{0}}, {[]int{5}, []int{4}}, {[]int{-2}, []int{-10}}, {[]int{9}, []int{12}}}},
}

func contractIndex(id string) int {
	return slices.IndexFunc(contracts, func(c contract) bool { return c.id == id })
}

func inputSemantics(id string) string {
	switch id {
	case "retain-active":
		return "[active,expired] are boolean 0/1; output1 means retain,0 remove; exhaustive four states"
	case "allow-owner":
		return "[administrator,owner,verified] are boolean 0/1; output1 accept,0 reject; exhaustive eight states"
	case "closed-window":
		return "[integer]; output1 accept iff within inclusive3..7; six literal integer examples only"
	case "quota-total":
		return "[used,requested,limit] are nonnegative small integers; output1 accept; six literal triples only, no overflow claim"
	case "stable-odd":
		return "integer sequence; output retained integer sequence; five literal sequences only"
	case "remove-first":
		return "integer sequence with target integer2; output remaining sequence; five literal sequences only"
	case "compact-runs":
		return "integer sequence; output one integer per adjacent equal run; five literal sequences only"
	case "rotate-left":
		return "integer sequence; output cyclic left rotation; five literal sequences only"
	case "nondecreasing":
		return "integer sequence; output1 accept,0 reject; six literal sequences only"
	case "clamp-window":
		return "[value,lower,upper] with lower<=upper; output one clamped integer; six literal triples only"
	case "prefix-balance":
		return "parenthesis steps +1=open,-1=close; output1 accept,0 reject; seven literal sequences only"
	case "transform-order":
		return "[integer]; output one transformed integer; five small literal integers only, no overflow claim"
	}
	return "unregistered"
}

type sourceDef struct {
	id, prototype, core, function string
	run                           func([]int) []int
	correct                       bool
}

var sourceCatalog = []sourceDef{
	{"retain-correct", "retain-active", "boolean-retention", "retainCorrect", retainCorrect, true},
	{"retain-reversed", "retain-active", "boolean-retention", "retainReversed", retainReversed, false},
	{"retain-ignore-expiry", "retain-active", "boolean-retention", "retainIgnoreExpiry", retainIgnoreExpiry, false},
	{"owner-correct", "allow-owner", "boolean-authorization", "ownerCorrect", ownerCorrect, true},
	{"owner-unverified", "allow-owner", "boolean-authorization", "ownerUnverified", ownerUnverified, false},
	{"owner-admin-verified", "allow-owner", "boolean-authorization", "ownerAdminVerified", ownerAdminVerified, false},
	{"window-correct", "closed-window", "closed-boundary", "windowCorrect", windowCorrect, true},
	{"window-exclusive", "closed-window", "closed-boundary", "windowExclusive", windowExclusive, false},
	{"window-inverted", "closed-window", "closed-boundary", "windowInverted", windowInverted, false},
	{"quota-correct", "quota-total", "sum-before-comparison", "quotaCorrect", quotaCorrect, true},
	{"quota-exclusive", "quota-total", "sum-before-comparison", "quotaExclusive", quotaExclusive, false},
	{"quota-separate", "quota-total", "sum-before-comparison", "quotaSeparate", quotaSeparate, false},
	{"odd-correct", "stable-odd", "stable-filter", "oddCorrect", oddCorrect, true},
	{"odd-reversed", "stable-odd", "stable-filter", "oddReversed", oddReversed, false},
	{"odd-positive", "stable-odd", "stable-filter", "oddPositive", oddPositive, false},
	{"first-correct", "remove-first", "single-removal", "firstCorrect", firstCorrect, true},
	{"first-all", "remove-first", "single-removal", "firstAll", firstAll, false},
	{"first-last", "remove-first", "single-removal", "firstLast", firstLast, false},
	{"compact-correct", "compact-runs", "adjacent-compaction", "compactCorrect", compactCorrect, true},
	{"compact-global", "compact-runs", "adjacent-compaction", "compactGlobal", compactGlobal, false},
	{"compact-copy", "compact-runs", "adjacent-compaction", "compactCopy", compactCopy, false},
	{"rotate-correct", "rotate-left", "cyclic-position", "rotateCorrect", rotateCorrect, true},
	{"rotate-right", "rotate-left", "cyclic-position", "rotateRight", rotateRight, false},
	{"rotate-reverse", "rotate-left", "cyclic-position", "rotateReverse", rotateReverse, false},
	{"sorted-correct", "nondecreasing", "adjacent-order", "sortedCorrect", sortedCorrect, true},
	{"sorted-strict", "nondecreasing", "adjacent-order", "sortedStrict", sortedStrict, false},
	{"sorted-endpoints", "nondecreasing", "adjacent-order", "sortedEndpoints", sortedEndpoints, false},
	{"clamp-correct", "clamp-window", "closed-boundary", "clampCorrect", clampCorrect, true},
	{"clamp-zero", "clamp-window", "closed-boundary", "clampZero", clampZero, false},
	{"clamp-wrap", "clamp-window", "closed-boundary", "clampWrap", clampWrap, false},
	{"balance-correct", "prefix-balance", "prefix-state", "balanceCorrect", balanceCorrect, true},
	{"balance-final-only", "prefix-balance", "prefix-state", "balanceFinalOnly", balanceFinalOnly, false},
	{"balance-first-only", "prefix-balance", "prefix-state", "balanceFirstOnly", balanceFirstOnly, false},
	{"transform-correct", "transform-order", "ordered-arithmetic", "transformCorrect", transformCorrect, true},
	{"transform-reversed", "transform-order", "ordered-arithmetic", "transformReversed", transformReversed, false},
	{"transform-add", "transform-order", "ordered-arithmetic", "transformAdd", transformAdd, false},
}

func check(fn func([]int) []int, vectors []vector) (checked, failed int, unknown bool) {
	defer func() {
		if recover() != nil {
			unknown = true
		}
	}()
	for _, v := range vectors {
		input := slices.Clone(v.Input)
		got := fn(input)
		checked++
		if !slices.Equal(input, v.Input) {
			return checked, failed, true
		}
		if !slices.Equal(got, v.Want) {
			failed++
		}
	}
	return
}

func appendUnique(out []string, s string) []string {
	if !slices.Contains(out, s) {
		return append(out, s)
	}
	return out
}

func groupAudit(d Dataset, pins []SourcePin) ([]GroupEdge, []Group) {
	owners := make([]int, len(d.Parents))
	for i := range owners {
		owners[i] = i
	}
	find := func(i int) int {
		for owners[i] != i {
			i = owners[i]
		}
		return i
	}
	var edges []GroupEdge
	for i, a := range d.Parents {
		for j := i + 1; j < len(d.Parents); j++ {
			b := d.Parents[j]
			reasons := []string{}
			if a.Prototype == b.Prototype {
				reasons = append(reasons, "shared_prototype_paraphrase_or_counterexample")
			}
			if contracts[contractIndex(a.Prototype)].core == contracts[contractIndex(b.Prototype)].core {
				reasons = append(reasons, "shared_core_template")
			}
			if lexicalhint.NormalizeText(a.Request) == lexicalhint.NormalizeText(b.Request) {
				reasons = append(reasons, "normalized_request_duplicate")
			}
			for _, x := range a.Candidates {
				for _, y := range b.Candidates {
					if x.SourceID == y.SourceID {
						reasons = appendUnique(reasons, "shared_candidate_source_including_negative")
					}
					px := pins[slices.IndexFunc(pins, func(p SourcePin) bool { return p.ID == x.SourceID })]
					py := pins[slices.IndexFunc(pins, func(p SourcePin) bool { return p.ID == y.SourceID })]
					if px.NormalizedCodeSHA256 == py.NormalizedCodeSHA256 {
						reasons = appendUnique(reasons, "normalized_candidate_code_duplicate")
					}
					if px.CoreTemplate == py.CoreTemplate {
						reasons = appendUnique(reasons, "shared_candidate_core_template")
					}
				}
			}
			if len(reasons) > 0 {
				edges = append(edges, GroupEdge{a.ID, b.ID, reasons})
				owners[find(j)] = find(i)
			}
		}
	}
	var groups []Group
	roots := []int{}
	for i, p := range d.Parents {
		root := find(i)
		k := slices.Index(roots, root)
		if k < 0 {
			k = len(groups)
			roots = append(roots, root)
			groups = append(groups, Group{ID: k})
		}
		g := &groups[k]
		g.Parents = append(g.Parents, p.ID)
		g.Prototypes = appendUnique(g.Prototypes, p.Prototype)
		g.CoreTemplates = appendUnique(g.CoreTemplates, contracts[contractIndex(p.Prototype)].core)
		for _, c := range p.Candidates {
			g.Sources = appendUnique(g.Sources, c.SourceID)
		}
	}
	return edges, groups
}

// Authored candidates are pure, small Go functions. Their observations are
// compared with the independent literal tables above, not used as the oracle.
func retainCorrect(v []int) []int {
	if v[0] == 1 && v[1] == 0 {
		return []int{1}
	}
	return []int{0}
}
func retainReversed(v []int) []int {
	if v[0] == 0 || v[1] == 1 {
		return []int{1}
	}
	return []int{0}
}
func retainIgnoreExpiry(v []int) []int {
	if v[0] == 1 {
		return []int{1}
	}
	return []int{0}
}
func ownerCorrect(v []int) []int {
	if v[0] == 1 || v[1] == 1 && v[2] == 1 {
		return []int{1}
	}
	return []int{0}
}
func ownerUnverified(v []int) []int {
	if v[0] == 1 || v[1] == 1 {
		return []int{1}
	}
	return []int{0}
}
func ownerAdminVerified(v []int) []int {
	if (v[0] == 1 || v[1] == 1) && v[2] == 1 {
		return []int{1}
	}
	return []int{0}
}
func windowCorrect(v []int) []int {
	if v[0] >= 3 && v[0] <= 7 {
		return []int{1}
	}
	return []int{0}
}
func windowExclusive(v []int) []int {
	if v[0] > 3 && v[0] < 7 {
		return []int{1}
	}
	return []int{0}
}
func windowInverted(v []int) []int {
	if v[0] < 3 || v[0] > 7 {
		return []int{1}
	}
	return []int{0}
}
func quotaCorrect(v []int) []int {
	if v[0]+v[1] <= v[2] {
		return []int{1}
	}
	return []int{0}
}
func quotaExclusive(v []int) []int {
	if v[0]+v[1] < v[2] {
		return []int{1}
	}
	return []int{0}
}
func quotaSeparate(v []int) []int {
	if v[0] <= v[2] && v[1] <= v[2] {
		return []int{1}
	}
	return []int{0}
}
func oddCorrect(v []int) []int {
	out := []int{}
	for _, x := range v {
		if x%2 != 0 {
			out = append(out, x)
		}
	}
	return out
}
func oddReversed(v []int) []int {
	out := []int{}
	for i := len(v) - 1; i >= 0; i-- {
		if v[i]%2 != 0 {
			out = append(out, v[i])
		}
	}
	return out
}
func oddPositive(v []int) []int {
	out := []int{}
	for _, x := range v {
		if x > 0 && x%2 != 0 {
			out = append(out, x)
		}
	}
	return out
}
func firstCorrect(v []int) []int {
	out := []int{}
	removed := false
	for _, x := range v {
		if x == 2 && !removed {
			removed = true
			continue
		}
		out = append(out, x)
	}
	return out
}
func firstAll(v []int) []int {
	out := []int{}
	for _, x := range v {
		if x != 2 {
			out = append(out, x)
		}
	}
	return out
}
func firstLast(v []int) []int {
	last := -1
	for i, x := range v {
		if x == 2 {
			last = i
		}
	}
	out := []int{}
	for i, x := range v {
		if i != last {
			out = append(out, x)
		}
	}
	return out
}
func compactCorrect(v []int) []int {
	out := []int{}
	for _, x := range v {
		if len(out) == 0 || out[len(out)-1] != x {
			out = append(out, x)
		}
	}
	return out
}
func compactGlobal(v []int) []int {
	out := []int{}
	for _, x := range v {
		seen := false
		for _, y := range out {
			if x == y {
				seen = true
			}
		}
		if !seen {
			out = append(out, x)
		}
	}
	return out
}
func compactCopy(v []int) []int {
	out := []int{}
	for _, x := range v {
		out = append(out, x)
	}
	return out
}
func rotateCorrect(v []int) []int {
	if len(v) == 0 {
		return []int{}
	}
	out := append([]int{}, v[1:]...)
	return append(out, v[0])
}
func rotateRight(v []int) []int {
	if len(v) == 0 {
		return []int{}
	}
	out := []int{v[len(v)-1]}
	return append(out, v[:len(v)-1]...)
}
func rotateReverse(v []int) []int {
	out := []int{}
	for i := len(v) - 1; i >= 0; i-- {
		out = append(out, v[i])
	}
	return out
}
func sortedCorrect(v []int) []int {
	for i := 1; i < len(v); i++ {
		if v[i] < v[i-1] {
			return []int{0}
		}
	}
	return []int{1}
}
func sortedStrict(v []int) []int {
	for i := 1; i < len(v); i++ {
		if v[i] <= v[i-1] {
			return []int{0}
		}
	}
	return []int{1}
}
func sortedEndpoints(v []int) []int {
	if len(v) > 1 && v[0] > v[len(v)-1] {
		return []int{0}
	}
	return []int{1}
}
func clampCorrect(v []int) []int {
	if v[0] < v[1] {
		return []int{v[1]}
	}
	if v[0] > v[2] {
		return []int{v[2]}
	}
	return []int{v[0]}
}
func clampZero(v []int) []int {
	if v[0] < v[1] || v[0] > v[2] {
		return []int{0}
	}
	return []int{v[0]}
}
func clampWrap(v []int) []int {
	if v[0] < v[1] {
		return []int{v[2]}
	}
	if v[0] > v[2] {
		return []int{v[1]}
	}
	return []int{v[0]}
}
func balanceCorrect(v []int) []int {
	balance := 0
	for _, x := range v {
		balance += x
		if balance < 0 {
			return []int{0}
		}
	}
	if balance == 0 {
		return []int{1}
	}
	return []int{0}
}
func balanceFinalOnly(v []int) []int {
	balance := 0
	for _, x := range v {
		balance += x
	}
	if balance == 0 {
		return []int{1}
	}
	return []int{0}
}
func balanceFirstOnly(v []int) []int {
	if len(v) > 0 && v[0] < 0 {
		return []int{0}
	}
	balance := 0
	for _, x := range v {
		balance += x
	}
	if balance == 0 {
		return []int{1}
	}
	return []int{0}
}
func transformCorrect(v []int) []int  { return []int{(v[0] - 3) * 2} }
func transformReversed(v []int) []int { return []int{v[0]*2 - 3} }
func transformAdd(v []int) []int      { return []int{(v[0] + 3) * 2} }
