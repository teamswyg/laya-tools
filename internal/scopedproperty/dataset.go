package scopedproperty

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const (
	Schema      = "riido-scoped-property-probes-v1"
	Origin      = "original_authored_quoted_property_drafts_56c_apache2_preparation_only"
	ParentCount = 12
	// OriginalDatasetSHA256 binds the whole frozen probes-56b.json, including its
	// canonical indentation and final newline, before any proposal is prepared.
	OriginalDatasetSHA256 = "0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e"
)

type Candidate struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	SourceID      string `json:"source_id"`
	CodeSHA256    string `json:"code_sha256"`
	BundleSHA256  string `json:"source_bundle_sha256"`
	CaptionReview string `json:"caption_review"`
}

type Parent struct {
	ID               string      `json:"id"`
	OriginalParentID string      `json:"original_parent_id"`
	PropertyID       string      `json:"property_id"`
	PropertyVersion  int         `json:"property_version"`
	TruthTableSHA256 string      `json:"truth_table_sha256"`
	CaptionReview    string      `json:"caption_review"`
	Request          string      `json:"request"`
	Candidates       []Candidate `json:"candidates"`
}

type Dataset struct {
	Schema        string   `json:"schema"`
	Origin        string   `json:"origin"`
	CaptionReview string   `json:"caption_review"`
	Parents       []Parent `json:"parents"`
}

type captionDraft struct {
	property string
	request  string
	texts    [4]string
}

// These are pending, bounded prose drafts. Neither their spelling nor the
// preparation factory grants semantic-review READY or any outcome label.
func captionDrafts() [3]captionDraft {
	return [3]captionDraft{
		{
			property: QuotedCommaProperty,
			request:  "ASCII without backslashes; raw<=128 bytes, tokens<=8, decoded bytes/token<=8. Preserve commas inside balanced double quotes, strip quotes, split outside commas, preserve empties. Return successful Count/Tokens; never panic.",
			texts: [4]string{
				"Strip double quotes, split commas outside balanced quotes, preserve decoded content and empty fields. Within bounded ASCII without backslashes return Count/Tokens and success; never panic.",
				"Split every comma, retaining quote bytes and empty fields. Retained quotes count toward field limits; overflow errors. Otherwise return literal fields/Count without interpreting quotes; never panic.",
				"Strip double quotes, split commas outside balanced quotes, preserve decoded content and empty fields. Within bounded ASCII without backslashes return Count/Tokens and success; never panic.",
				"Strip double quotes, split commas outside balanced quotes, preserve decoded content and empty fields. Within bounded ASCII without backslashes return Count/Tokens and success; never panic.",
			},
		},
		{
			property: OutsideEscapeProperty,
			request:  "ASCII without quotes/trailing escapes; raw<=128 bytes, tokens<=8, decoded bytes/token<=8. Backslash consumes next byte as content; split unescaped commas, preserve empty fields. Return successful Count/Tokens; never panic.",
			texts: [4]string{
				"Remove an escape backslash and keep the next byte as content; split unescaped commas, preserve empties. Bounded ASCII without quotes/trailing escapes returns Count/Tokens, success; never panic.",
				"Keep backslashes literally, split every comma and preserve empties. Retained backslashes count toward field limits; return Count/Tokens or bounds error on overflow. Never panic.",
				"Keep outside-quote backslashes literally, split every comma and preserve empties. Retained backslashes count toward field limits; return Count/Tokens or bounds error on overflow. Never panic.",
				"Remove an escape backslash and keep the next byte as content; split unescaped commas, preserve empties. Bounded ASCII without quotes/trailing escapes returns Count/Tokens, success; never panic.",
			},
		},
		{
			property: SyntaxZeroProperty,
			request:  "ASCII raw<=128 bytes, tokens<=8, decoded bytes/token<=8: unclosed quotes or trailing escapes must return exact syntax error, Count=0 and all eight Tokens empty; never panic.",
			texts: [4]string{
				"For bounded ASCII syntax errors return exact syntax error, Count=0 and all eight Tokens empty, without panic.",
				"Treat quotes/backslashes literally, split every comma and retain field bytes. Within retained-field limits return success, Count/Tokens, including unclosed quotes/trailing escapes; never panic.",
				"Unclosed quotes/trailing escapes inside quotes return syntax error with zero output. Outside quotes retain backslashes literally; trailing outside escapes may return success. Never panic.",
				"Report syntax errors while retaining parsed and unfinished field content and partial Count. Unused Tokens remain empty; never panic.",
			},
		},
	}
}

// CreateDataset prepares twelve proposals: four existing unknown parents times
// three properties. PrepareDataset validates original SourcePins/closures but
// never runs candidates, groups them or looks up old acceptable sets.
func CreateDataset() (Dataset, error) {
	return canonicalDataset()
}

func canonicalDataset() (Dataset, error) {
	original, e := typedbehavior.PrepareDataset()
	if e != nil {
		return Dataset{}, errors.New("scoped_original_source_preparation_failed")
	}
	if e := verifyOriginalDataset(original); e != nil {
		return Dataset{}, e
	}
	ids, sources, defs, captions := OriginalParentIDs(), SourceIDs(), Definitions(), captionDrafts()
	out := Dataset{Schema: Schema, Origin: Origin, CaptionReview: PendingReview, Parents: make([]Parent, 0, ParentCount)}
	for i, def := range defs {
		if captions[i].property != def.ID {
			return Dataset{}, errors.New("scoped_draft_registry_invalid")
		}
		for _, originalID := range ids {
			n := slices.IndexFunc(original.Parents, func(p typedbehavior.Parent) bool { return p.ID == originalID })
			if n < 0 {
				return Dataset{}, errors.New("scoped_original_parent_missing")
			}
			old := original.Parents[n]
			if old.Prototype != "quoted-delimiters" || len(old.Candidates) != 3 || old.ContractID != "unknown-incomplete-caption" && old.ContractID != "unknown-ambiguous" {
				return Dataset{}, errors.New("scoped_original_parent_binding_invalid")
			}
			p := Parent{ID: fmt.Sprintf("scope56c-p%d-%s", i+1, originalID), OriginalParentID: old.ID,
				PropertyID: def.ID, PropertyVersion: def.Version, TruthTableSHA256: def.TruthTableSHA256,
				CaptionReview: PendingReview, Request: captions[i].request, Candidates: make([]Candidate, len(old.Candidates))}
			for j, c := range old.Candidates {
				index := slices.Index(sources[:], c.SourceID)
				if index < 0 {
					return Dataset{}, errors.New("scoped_original_source_not_registered")
				}
				p.Candidates[j] = Candidate{ID: c.ID, Text: captions[i].texts[index], SourceID: c.SourceID,
					CodeSHA256: c.CodeSHA256, BundleSHA256: c.BundleSHA256, CaptionReview: PendingReview}
			}
			if _, e := shortclaim.Validate(featureInput(p)); e != nil {
				return Dataset{}, errors.New("scoped_caption_text_or_candidate_bounds_invalid")
			}
			out.Parents = append(out.Parents, p)
		}
	}
	return out, nil
}

// Bind the entire reconstructed original to the frozen public bytes, rather
// than accepting just the selected parents' current metadata or source pins.
func verifyOriginalDataset(original typedbehavior.Dataset) error {
	raw, e := json.MarshalIndent(original, "", "  ")
	if e != nil {
		return errors.New("scoped_original_dataset_encoding_failed")
	}
	if sha(append(raw, '\n')) != OriginalDatasetSHA256 {
		return errors.New("scoped_original_dataset_hash_mismatch")
	}
	return nil
}

// ValidateDataset accepts only this closed pending proposal version. A later
// caption-review status or different caption requires a separate frozen version.
// Canonical generation binds the current actual original source/closure pins.
func ValidateDataset(d Dataset) error {
	if d.Schema != Schema || d.Origin != Origin || d.CaptionReview != PendingReview || len(d.Parents) != ParentCount {
		return errors.New("scoped_dataset_identity_or_review_invalid")
	}
	want, e := canonicalDataset()
	if e != nil {
		return e
	}
	for i, p := range d.Parents {
		w := want.Parents[i]
		if p.ID != w.ID || p.OriginalParentID != w.OriginalParentID || p.PropertyID != w.PropertyID || p.PropertyVersion != w.PropertyVersion || p.TruthTableSHA256 != w.TruthTableSHA256 || p.CaptionReview != PendingReview || p.Request != w.Request || len(p.Candidates) != len(w.Candidates) {
			return errors.New("scoped_parent_or_property_binding_invalid")
		}
		for j, c := range p.Candidates {
			if c != w.Candidates[j] {
				return errors.New("scoped_candidate_source_order_or_caption_invalid")
			}
		}
		if _, e := shortclaim.Validate(featureInput(p)); e != nil {
			return errors.New("scoped_caption_text_or_candidate_bounds_invalid")
		}
	}
	return nil
}

func featureInput(p Parent) shortclaim.Input {
	out := shortclaim.Input{Schema: shortclaim.Schema, Request: p.Request,
		Provenance: "scoped-property-development-56c", Candidates: make([]shortclaim.Candidate, len(p.Candidates))}
	for i, c := range p.Candidates {
		out.Candidates[i] = shortclaim.Candidate{ID: fmt.Sprintf("candidate-%d", i), Text: c.Text}
	}
	return out
}

// FeatureInputs carries only request/candidate text from source records. The
// runtime envelope's IDs and provenance are regenerated constants, not features.
// Calling it does not validate, evaluate, rank or change the source records.
func FeatureInputs(d Dataset) []shortclaim.Input {
	out := make([]shortclaim.Input, len(d.Parents))
	for i, p := range d.Parents {
		out[i] = featureInput(p)
	}
	return out
}
