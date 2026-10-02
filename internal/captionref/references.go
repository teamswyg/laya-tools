// SPDX-License-Identifier: Apache-2.0
// Package captionref connects immutable public development text to source
// expressions. It does not execute candidates or approve their descriptions.
package captionref

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

	"github.com/teamswyg/laya-tools/internal/storedaudit"
)

const Schema = "riido-caption-reference-preparation-59-v1"

type Artifact = storedaudit.FilePin
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrPin      Error = "captionref_input_pin_mismatch"
	ErrMetadata Error = "captionref_historical_contract_mismatch"
	ErrAST      Error = "captionref_literal_ast_invalid"
	ErrTarget   Error = "captionref_literal_target_invalid"
	ErrSpan     Error = "captionref_literal_span_invalid"
	ErrCounts   Error = "captionref_reference_count_mismatch"
)

// InputPins is closed and ordered: stored prose/outcomes first, literal source
// bytes second. No candidate package is imported to recover a missing table.
func InputPins() [6]Artifact {
	return [6]Artifact{
		{Path: "experiments/short-claim/probes-56.json", SHA256: storedaudit.LegacySHA256, Bytes: 49941},
		{Path: "experiments/short-claim/probes-56b.json", SHA256: storedaudit.TypedSHA256, Bytes: 43598},
		{Path: "experiments/short-claim/results-56b.json", SHA256: storedaudit.ResultSHA256, Bytes: 266818},
		{Path: "internal/behaviorprobe/data.go", SHA256: "b377343a64c019c205f69865286a21ce5d91dda3051c31aa5c69b75a40300559", Bytes: 28420},
		{Path: "internal/typedbehavior/state.go", SHA256: "35324ed595bcf384c484f33d98c6866164bebb2b636986ee8ea51bb979bdad7c", Bytes: 19402},
		{Path: "internal/typedbehavior/flow.go", SHA256: "ea7f09d31e15e561173932ac0bb61db17746cdce929e0f976d97994d282d898b", Bytes: 18242},
	}
}

type Inputs struct {
	Stored   [3][]byte
	Literals [3][]byte
}

type Counters struct {
	BindAttempts       int `json:"metadata_bind_attempts"`
	BindCompleted      int `json:"metadata_bind_completed"`
	ASTParseAttempts   int `json:"ast_parse_attempts"`
	ASTParsesCompleted int `json:"ast_parses_completed"`
	ParentsGenerated   int `json:"parents_generated"`
	CaptionsGenerated  int `json:"captions_generated"`
	ContractsGenerated int `json:"contracts_generated"`
}

type ExpressionReference struct {
	File              string `json:"file"`
	Symbol            string `json:"top_level_symbol"`
	Kind              string `json:"expression_kind"`
	StartByte         int    `json:"start_byte"`
	EndByte           int    `json:"end_byte"`
	StartLine         int    `json:"start_line"`
	EndLine           int    `json:"end_line"`
	RawSHA256         string `json:"raw_expression_sha256"`
	VectorExpressions int    `json:"literal_vector_expressions"`
}

type TextReference struct {
	Path    string `json:"path"`
	Pointer string `json:"json_pointer"`
	SHA256  string `json:"decoded_utf8_sha256"`
	Bytes   int    `json:"decoded_utf8_bytes"`
}

type CandidateReference struct {
	Index           int                        `json:"candidate_index"`
	ID              string                     `json:"candidate_id"`
	Caption         TextReference              `json:"caption"`
	SourceID        string                     `json:"source_id"`
	CodeSHA256      string                     `json:"historical_code_sha256"`
	BundleSHA256    string                     `json:"historical_bundle_sha256,omitempty"`
	HistoricalTruth storedaudit.CandidateTruth `json:"historical_candidate_outcome"`
	SourceClosure   string                     `json:"source_closure_review"`
	SourceFidelity  string                     `json:"source_fidelity_review"`
	ContractReview  string                     `json:"request_contract_coverage_review"`
}

type ParentReference struct {
	Index         int                  `json:"original_parent_index"`
	ID            string               `json:"parent_id"`
	Cohort        string               `json:"cohort"`
	Prototype     string               `json:"prototype"`
	ContractID    string               `json:"contract_id"`
	GroupID       int                  `json:"historical_group_id"`
	Request       TextReference        `json:"request"`
	State         string               `json:"historical_truth_state"`
	UnknownReason string               `json:"historical_unknown_reason,omitempty"`
	Acceptable    []int                `json:"historical_acceptable_candidate_indices"`
	Candidates    []CandidateReference `json:"candidates"`
	Review        string               `json:"request_contract_review"`
}

type HistoricalContract struct {
	Prototype string `json:"prototype"`
	Core      string `json:"core_template"`
	Semantics string `json:"input_semantics"`
	Vectors   int    `json:"literal_vectors"`
	TableSHA  string `json:"truth_table_sha256"`
}

type ContractReference struct {
	Historical HistoricalContract  `json:"historical_contract"`
	Expression ExpressionReference `json:"literal_source_reference"`
	Reified    bool                `json:"literal_payload_reified"`
	Review     string              `json:"observation_fields_review"`
}

type OriginalSummary struct {
	Parents     int    `json:"parents"`
	Captions    int    `json:"candidate_captions"`
	Answerable  int    `json:"answerable"`
	NoAnswer    int    `json:"no_answer"`
	Unknown     int    `json:"unknown"`
	Groups      int    `json:"connected_groups"`
	KnownGroups int    `json:"known_containing_groups"`
	Histogram   [5]int `json:"candidate_count_histogram"`
}

type Report struct {
	Schema         string                  `json:"schema"`
	Scope          string                  `json:"scope"`
	Preparation    bool                    `json:"preparation_only"`
	RawInputs      [3]Artifact             `json:"raw_inputs"`
	LiteralSources [3]Artifact             `json:"literal_source_files"`
	Original       OriginalSummary         `json:"original"`
	Parents        []ParentReference       `json:"parents"`
	Contracts      []ContractReference     `json:"contracts"`
	Groups         [17]storedaudit.Group   `json:"historical_groups"`
	Edges          []storedaudit.GroupEdge `json:"historical_group_edges"`
	GroupPolicy    string                  `json:"historical_group_policy"`
	Binding        string                  `json:"metadata_binding_validation"`
	ContentReview  string                  `json:"content_review"`
	NewLabels      int                     `json:"new_labels"`
	NewParents     int                     `json:"new_independent_parents"`
	Roles          int                     `json:"roles_assigned"`
	Fits           int                     `json:"fits"`
	ModelCalls     int                     `json:"model_calls"`
	SourceCalls    int                     `json:"source_api_calls"`
	Rankings       int                     `json:"ranking_calls"`
	TrainingReady  bool                    `json:"training_ready"`
	StopReasons    []string                `json:"original_stop_reasons_preserved"`
	Limitations    []string                `json:"limitations"`
}

func digest(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func textReference(path, pointer, text string) TextReference {
	return TextReference{Path: path, Pointer: pointer, SHA256: digest([]byte(text)), Bytes: len(text)}
}

// Generate validates the fixed bytes, then binds stored outcomes once. Every
// counter increments before its operation; errors preserve the work attempted.
// Nothing in this function executes literal expressions or candidate source.
func Generate(in Inputs, counts *Counters) (Report, error) {
	if counts == nil || *counts != (Counters{}) {
		return Report{}, ErrCounts
	}
	pins := InputPins()
	for i := range in.Stored {
		if len(in.Stored[i]) != pins[i].Bytes || digest(in.Stored[i]) != pins[i].SHA256 {
			return Report{}, ErrPin
		}
		if len(in.Literals[i]) != pins[i+3].Bytes || digest(in.Literals[i]) != pins[i+3].SHA256 {
			return Report{}, ErrPin
		}
	}
	counts.BindAttempts++
	d, err := storedaudit.Bind(in.Stored[0], in.Stored[1], in.Stored[2])
	if err != nil {
		return Report{}, err
	}
	counts.BindCompleted++
	var historical struct {
		Legacy struct {
			Contracts []HistoricalContract `json:"contracts"`
		} `json:"legacy_truth"`
		Typed struct {
			Contracts []HistoricalContract `json:"contracts"`
		} `json:"typed_truth"`
	}
	if json.Unmarshal(in.Stored[2], &historical) != nil || len(historical.Legacy.Contracts) != 12 || len(historical.Typed.Contracts) != 6 {
		return Report{}, ErrMetadata
	}
	contracts := append(historical.Legacy.Contracts, historical.Typed.Contracts...)
	r := Report{
		Schema: Schema, Preparation: true, Scope: "References over previously observed public development records; not a fresh truth audit, source-closure approval or completed semantic review.",
		RawInputs: [3]Artifact{pins[0], pins[1], pins[2]}, LiteralSources: [3]Artifact{pins[3], pins[4], pins[5]},
		Original: OriginalSummary{len(d.Rows), d.Candidates, d.Answerable, d.NoAnswer, d.Unknown, len(d.Groups), d.LabeledGroups, d.CandidateCountHistogram},
		Parents:  make([]ParentReference, 0, 72), Contracts: make([]ContractReference, 0, 18),
		Groups: d.Groups, Edges: d.Metadata.Edges, GroupPolicy: d.Metadata.GroupPolicy,
		Binding: "exact frozen raw binding passed", ContentReview: "pending", StopReasons: d.Metadata.StopReasons,
		Limitations: []string{
			"Raw expression SHA is distinct from the historical serialized literal-table SHA and formatted candidate code/bundle digests.",
			"AST references do not reify literal Input/Want, execute candidates, recover Got values or expand finite input support.",
			"Incorrect implementations can have faithful captions; source fidelity and request-contract coverage require separate independent review.",
			"Candidate helper/type/error-sentinel closure remains pending; historical metadata pins alone do not approve it.",
			"Original72 parents/216 candidate positions/unknown21/groups remain unchanged; no labels, roles, fits or rankings are created.",
			"Previously observed English development data is not unseen, blind, Korean-input performance or protected final.",
		},
	}
	for source := range in.Literals {
		counts.ASTParseAttempts++
		refs, err := literalReferences(pins[source+3].Path, in.Literals[source], source)
		if err != nil {
			return r, err
		}
		counts.ASTParsesCompleted++
		for _, ref := range refs {
			found := -1
			for i, c := range contracts {
				if c.Prototype == ref.prototype {
					if found >= 0 {
						return r, ErrMetadata
					}
					found = i
				}
			}
			if found < 0 {
				return r, ErrMetadata
			}
			if contracts[found].Vectors != ref.span.VectorExpressions {
				return r, ErrMetadata
			}
			for _, old := range r.Contracts {
				if old.Historical.Prototype == ref.prototype {
					return r, ErrTarget
				}
			}
			r.Contracts = append(r.Contracts, ContractReference{Historical: contracts[found], Expression: ref.span, Review: "pending"})
			counts.ContractsGenerated++
		}
	}
	for i, row := range d.Rows {
		local, path := i, pins[0].Path
		if i >= 48 {
			local, path = i-48, pins[1].Path
		}
		pointer := fmt.Sprintf("/parents/%d", local)
		p := ParentReference{Index: i, ID: row.Audit.ParentID, Cohort: row.Audit.Cohort, Prototype: row.Audit.Prototype, ContractID: row.Audit.ContractID, GroupID: row.Audit.GroupID,
			Request: textReference(path, pointer+"/request", row.Text.Request), State: row.Truth.State, UnknownReason: row.Truth.Reason,
			Acceptable: append([]int{}, row.Truth.AcceptableIndices[:row.Truth.AcceptableCount]...), Candidates: make([]CandidateReference, 0, row.Text.CandidateCount), Review: "pending"}
		for j := 0; j < row.Text.CandidateCount; j++ {
			a := row.Audit.Candidates[j]
			p.Candidates = append(p.Candidates, CandidateReference{Index: j, ID: a.ID, Caption: textReference(path, fmt.Sprintf("%s/candidates/%d/text", pointer, j), row.Text.Candidates[j]),
				SourceID: a.SourceID, CodeSHA256: a.CodeSHA256, BundleSHA256: a.BundleSHA256, HistoricalTruth: row.Truth.Candidates[j],
				SourceClosure: "pending", SourceFidelity: "pending", ContractReview: "pending"})
			counts.CaptionsGenerated++
		}
		r.Parents = append(r.Parents, p)
		counts.ParentsGenerated++
	}
	if counts.ParentsGenerated != 72 || counts.CaptionsGenerated != 216 || counts.ContractsGenerated != 18 || counts.ASTParsesCompleted != 3 {
		return r, ErrCounts
	}
	return r, nil
}

type namedReference struct {
	prototype string
	span      ExpressionReference
}

func expressionSpan(fs *token.FileSet, node ast.Node, file, symbol, kind string, raw []byte) (ExpressionReference, error) {
	start, end := fs.PositionFor(node.Pos(), false), fs.PositionFor(node.End()-1, false)
	endByte := end.Offset + 1
	if start.Offset < 0 || start.Offset >= endByte || endByte > len(raw) || start.Line < 1 || end.Line < start.Line {
		return ExpressionReference{}, ErrSpan
	}
	return ExpressionReference{File: file, Symbol: symbol, Kind: kind, StartByte: start.Offset, EndByte: endByte, StartLine: start.Line, EndLine: end.Line, RawSHA256: digest(raw[start.Offset:endByte])}, nil
}

// literalReferences reads a closed set of top-level variable initializers.
// Local shadows, //line remapping, aliases and callable expressions never become
// candidate truth. Returned hashes cover original source bytes [start,end).
func literalReferences(path string, raw []byte, source int) ([]namedReference, error) {
	if source < 0 || source >= 3 {
		return nil, ErrTarget
	}
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, path, raw, 0)
	if err != nil {
		return nil, ErrAST
	}
	targets := [3][]struct {
		symbol, prototype, element string
		inferredArray              bool
	}{
		{{"contracts", "", "contract", false}},
		{{"atomicVectors", "atomic-commit", "atomicVector", false}, {"identityVectors", "error-identity", "identityVector", false}, {"snapshotVectors", "owned-snapshot", "snapshotVector", false}},
		{{"lifecycleVectors", "cancellation-lifecycle", "lifecycleVector", true}, {"quotedVectors", "quoted-delimiters", "quotedVector", true}, {"graphVectors", "ancestor-cycle", "graphVector", true}},
	}
	var refs []namedReference
	for _, target := range targets[source] {
		var literal *ast.CompositeLit
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				v, ok := spec.(*ast.ValueSpec)
				if !ok {
					return nil, ErrTarget
				}
				for _, name := range v.Names {
					if name.Name != target.symbol {
						continue
					}
					if literal != nil || len(v.Names) != 1 || len(v.Values) != 1 {
						return nil, ErrTarget
					}
					literal, ok = v.Values[0].(*ast.CompositeLit)
					if !ok {
						return nil, ErrTarget
					}
					array, ok := literal.Type.(*ast.ArrayType)
					if !ok {
						return nil, ErrTarget
					}
					element, ok := array.Elt.(*ast.Ident)
					if !ok || element.Name != target.element {
						return nil, ErrTarget
					}
					if target.inferredArray {
						if _, ok := array.Len.(*ast.Ellipsis); !ok {
							return nil, ErrTarget
						}
					} else if array.Len != nil {
						return nil, ErrTarget
					}
				}
			}
		}
		if literal == nil {
			return nil, ErrTarget
		}
		if source == 0 {
			if len(literal.Elts) != 12 {
				return nil, ErrTarget
			}
			for _, element := range literal.Elts {
				el, ok := element.(*ast.CompositeLit)
				if !ok || el.Type != nil || len(el.Elts) != 3 {
					return nil, ErrTarget
				}
				label, ok := el.Elts[0].(*ast.BasicLit)
				if !ok || label.Kind != token.STRING {
					return nil, ErrTarget
				}
				prototype, err := strconv.Unquote(label.Value)
				if err != nil || prototype == "" {
					return nil, ErrTarget
				}
				for _, old := range refs {
					if old.prototype == prototype {
						return nil, ErrTarget
					}
				}
				span, err := expressionSpan(fs, el, path, target.symbol, "legacy_contract_element", raw)
				if err != nil {
					return nil, err
				}
				vectors, ok := el.Elts[2].(*ast.CompositeLit)
				if !ok {
					return nil, ErrTarget
				}
				vectorType, ok := vectors.Type.(*ast.ArrayType)
				if !ok || vectorType.Len != nil {
					return nil, ErrTarget
				}
				vectorElement, ok := vectorType.Elt.(*ast.Ident)
				if !ok || vectorElement.Name != "vector" {
					return nil, ErrTarget
				}
				span.VectorExpressions = len(vectors.Elts)
				refs = append(refs, namedReference{prototype, span})
			}
		} else {
			span, err := expressionSpan(fs, literal, path, target.symbol, "typed_vector_initializer", raw)
			if err != nil {
				return nil, err
			}
			span.VectorExpressions = len(literal.Elts)
			refs = append(refs, namedReference{target.prototype, span})
		}
	}
	return refs, nil
}
