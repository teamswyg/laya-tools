// Package finiteproperty audits a separately authored finite-input contract.
// It never promotes the general pending captions from preparation56c.
package finiteproperty

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/teamswyg/laya-tools/internal/scopedproperty"
	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const Schema = "riido-finite-input-property-56e-v2"

type Candidate struct {
	ID           string `json:"id"`
	Text         string `json:"text"`
	SourceID     string `json:"source_id"`
	CodeSHA256   string `json:"code_sha256"`
	BundleSHA256 string `json:"source_bundle_sha256"`
}

type Parent struct {
	ID               string       `json:"id"`
	OriginalParentID string       `json:"original_parent_id"`
	PropertyID       string       `json:"property_id"`
	Request          string       `json:"request"`
	Candidates       [3]Candidate `json:"candidates"`
}

type Dataset struct {
	Schema                string                               `json:"schema"`
	Origin                string                               `json:"origin"`
	ReviewID              string                               `json:"caption_review_id"`
	OriginalDatasetSHA256 string                               `json:"original_dataset_sha256"`
	Parents               [12]Parent                           `json:"parents"`
	Oracle                [3]scopedproperty.PropertyDefinition `json:"independent_literal_oracle"`
	ConnectedSources      [4]string                            `json:"connected_sources"`
	NewIndependentParents int                                  `json:"new_independent_parents"`
}

type prose struct {
	Request string
	Text    [4]string
}

// These AI-assisted requests are directly authored task text, not a projection of
// evaluator literals. Hex spelling retains byte distinctions after punctuation
// normalization. It does not prove a semantic encoder understands those bytes.
// Candidate prose is read from the source algorithm before observing outcomes.
func captions() [3]prose {
	return [3]prose{
		{
			Request: "Only ASCII inputs encoded as hex 0x22612c62222c63 and 0x616222632c64226566. Remove quotes, preserve quoted commas, split outside commas. Require success, exact fields/Count, unused token slots empty, no panic.",
			Text: [4]string{
				"For the requested inputs, remove quotes, preserve quoted commas, split outside commas. Return successful fields and Count, with unused token slots empty and no panic.",
				"For the requested inputs, retain quotes as literal bytes and split every comma. Return successful fields and Count, unused token slots empty, no panic.",
				"For the requested inputs, remove quotes, preserve quoted commas, split outside commas. Return successful fields and Count, with unused token slots empty and no panic.",
				"For the requested inputs, remove quotes, preserve quoted commas, split outside commas. Return successful fields and Count, with unused token slots empty and no panic.",
			},
		},
		{
			Request: "Only ASCII inputs encoded as hex 0x615c2c622c63 and 0x615c5c2c62. Backslash consumes the next byte as content; unescaped commas split. Require success, exact fields/Count, unused token slots empty, no panic.",
			Text: [4]string{
				"For the requested inputs, backslash consumes the next byte as content; unescaped commas split. Return successful fields and Count, unused token slots empty, no panic.",
				"For the requested inputs, retain backslashes and split every comma. Return successful fields and Count, unused token slots empty, no panic.",
				"For the requested inputs, retain backslashes and split every comma. Return successful fields and Count, unused token slots empty, no panic.",
				"For the requested inputs, backslash consumes the next byte as content; unescaped commas split. Return successful fields and Count, unused token slots empty, no panic.",
			},
		},
		{
			Request: "Only ASCII inputs encoded as hex 0x22612c62, 0x615c, 0x612c625c, 0x22615c, 0x612c2262. Require exact syntax error, Count zero, all eight token slots empty, no panic.",
			Text: [4]string{
				"For the requested inputs, report syntax error, Count zero, all eight token slots empty, no panic.",
				"For the requested inputs, retain quotes and backslashes, split every comma, and return success with fields and Count. Unused slots stay empty; no panic.",
				"For the requested inputs, unclosed quotes or trailing escapes inside quotes yield syntax error and zero output; outside backslashes remain literal with success. No panic.",
				"For the requested inputs, report syntax error while retaining parsed and unfinished fields and partial Count. Unused token slots remain empty; no panic.",
			},
		},
	}
}

// Prepare binds the old candidate identity/order/closures without running any
// candidate, label audit or ranking. The old dataset remains pending unchanged.
func Prepare() (Dataset, error) {
	old, err := scopedproperty.CreateDataset()
	if err != nil {
		return Dataset{}, err
	}
	out := Dataset{Schema: Schema, Origin: "original_authored_finite_input_56e_apache2",
		ReviewID: "finite-source-reading-review-56e-v2", OriginalDatasetSHA256: scopedproperty.OriginalDatasetSHA256,
		Oracle: scopedproperty.Definitions(), ConnectedSources: scopedproperty.SourceIDs()}
	text := captions()
	for i, p := range old.Parents {
		property := i / 4
		row := Parent{ID: fmt.Sprintf("finite56e-p%d-%s", property+1, p.OriginalParentID), OriginalParentID: p.OriginalParentID,
			PropertyID: fmt.Sprintf("finite-input-56e-v2-%d", property+1), Request: text[property].Request}
		for j, c := range p.Candidates {
			source := -1
			for k, id := range out.ConnectedSources {
				if id == c.SourceID {
					source = k
				}
			}
			if source < 0 {
				return Dataset{}, errors.New("finite_source_binding_invalid")
			}
			row.Candidates[j] = Candidate{c.ID, text[property].Text[source], c.SourceID, c.CodeSHA256, c.BundleSHA256}
		}
		if _, err := shortclaim.Validate(FeatureInput(row)); err != nil {
			return Dataset{}, errors.New("finite_caption_bounds_invalid")
		}
		out.Parents[i] = row
	}
	if err := validateScopeText(out); err != nil {
		return Dataset{}, err
	}
	return out, nil
}

// The publicly stated finite scope must equal the independent oracle's input
// scope. This comparison validates the request; it never fills in missing text.
func validateScopeText(d Dataset) error {
	for property, def := range d.Oracle {
		var declared [5]string
		n := 0
		for _, word := range strings.Fields(d.Parents[property*4].Request) {
			word = strings.TrimSuffix(strings.TrimSuffix(word, ","), ".")
			if !strings.HasPrefix(word, "0x") {
				continue
			}
			if n == len(declared) {
				return errors.New("finite_scope_text_invalid")
			}
			input, err := hex.DecodeString(word[2:])
			if err != nil {
				return errors.New("finite_scope_text_invalid")
			}
			declared[n] = string(input)
			n++
		}
		if n != def.LiteralCount {
			return errors.New("finite_scope_text_invalid")
		}
		for i := 0; i < n; i++ {
			if declared[i] != def.Literals[i].Input {
				return errors.New("finite_scope_text_invalid")
			}
		}
	}
	return nil
}

// FeatureInput is a prospective public text envelope, also used for preflight
// bounds validation. The observation loop never scores it. Metadata, expected
// values and source identity never enter its text features.
func FeatureInput(p Parent) shortclaim.Input {
	out := shortclaim.Input{Schema: shortclaim.Schema, Request: p.Request, Provenance: "finite-input-development-56e", Candidates: make([]shortclaim.Candidate, 3)}
	for i, c := range p.Candidates {
		out.Candidates[i] = shortclaim.Candidate{ID: fmt.Sprintf("candidate-%d", i), Text: c.Text}
	}
	return out
}

func Validate(d Dataset) error {
	want, err := Prepare()
	if err != nil {
		return err
	}
	if d != want {
		return errors.New("finite_dataset_changed")
	}
	return nil
}

func SHA256(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

func JSON(v any) ([]byte, error) {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

//go:embed dataset.go audit.go
var compiledSource embed.FS

func SourceArtifacts() [2]typedbehavior.SourceArtifact {
	var out [2]typedbehavior.SourceArtifact
	for i, name := range [2]string{"dataset.go", "audit.go"} {
		raw, _ := compiledSource.ReadFile(name)
		out[i] = typedbehavior.SourceArtifact{Name: name, SHA256: SHA256(raw)}
	}
	return out
}
