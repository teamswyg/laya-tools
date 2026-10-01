package typedbehavior

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const (
	Schema       = "riido-typed-behavior-probes-v2"
	Origin       = "original_authored_typed_go_microcontracts_56b_apache2_development_only"
	MaxDataBytes = 64 << 20
	MaxParents   = 120
)

type Candidate struct {
	ID           string `json:"id"`
	Text         string `json:"text"`
	SourceID     string `json:"source_id"`
	CodeSHA256   string `json:"code_sha256"`
	BundleSHA256 string `json:"source_bundle_sha256"`
}

type Parent struct {
	ID         string      `json:"id"`
	Prototype  string      `json:"prototype"`
	ContractID string      `json:"contract_id"`
	Request    string      `json:"request"`
	Candidates []Candidate `json:"candidates"`
}

type Dataset struct {
	Schema  string   `json:"schema"`
	Origin  string   `json:"origin"`
	Parents []Parent `json:"parents"`
}

// InputText is the entire scoring projection. Even derived function types,
// literal facts, candidate IDs, source identities and roles are absent.
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

func LoadBytes(raw []byte) (Dataset, error) {
	if len(raw) > MaxDataBytes || !utf8.Valid(raw) {
		return Dataset{}, errors.New("typed_dataset_byte_or_utf8_limit")
	}
	var d Dataset
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&d) != nil || decoder.Decode(new(any)) != io.EOF {
		return Dataset{}, errors.New("typed_dataset_json_invalid")
	}
	pins, e := SourcePins()
	if e != nil {
		return Dataset{}, e
	}
	if e = validateDataset(d, pins); e != nil {
		return Dataset{}, e
	}
	return d, nil
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

func validateDataset(d Dataset, pins []SourcePin) error {
	if d.Schema != Schema || d.Origin != Origin || len(d.Parents) < 1 || len(d.Parents) > MaxParents {
		return errors.New("typed_dataset_schema_or_count_invalid")
	}
	contracts := Contracts()
	for i, p := range d.Parents {
		if !validID(p.ID) || !validID(p.Prototype) {
			return errors.New("typed_parent_identity_invalid")
		}
		for j := 0; j < i; j++ {
			if d.Parents[j].ID == p.ID {
				return errors.New("typed_parent_duplicate")
			}
		}
		if !slices.ContainsFunc(contracts, func(c ContractSpec) bool { return c.Prototype == p.Prototype }) {
			return errors.New("typed_contract_not_registered")
		}
		if p.ContractID != p.Prototype+"-v2" && p.ContractID != "unknown-ambiguous" && p.ContractID != "unknown-incomplete-caption" {
			return errors.New("typed_contract_identity_invalid")
		}
		fixture := slices.IndexFunc(fixturePrototypes, func(f prototypeFixture) bool { return f.prototype == p.Prototype })
		if fixture < 0 {
			return errors.New("typed_caption_registry_missing")
		}
		f := fixturePrototypes[fixture]
		if p.ContractID == p.Prototype+"-v2" && (!f.complete || !slices.Contains(f.requests[:], p.Request)) {
			return errors.New("typed_unreviewed_caption_cannot_be_known")
		}
		input := shortclaim.Input{Schema: shortclaim.Schema, Request: p.Request, Provenance: "typed-development-56b"}
		for j, c := range p.Candidates {
			if !validID(c.ID) {
				return errors.New("typed_candidate_identity_invalid")
			}
			for k := 0; k < j; k++ {
				if p.Candidates[k].ID == c.ID {
					return errors.New("typed_candidate_duplicate")
				}
			}
			pin := slices.IndexFunc(pins, func(s SourcePin) bool { return s.ID == c.SourceID })
			if pin < 0 || pins[pin].Prototype != p.Prototype || pins[pin].CodeSHA256 != c.CodeSHA256 || pins[pin].BundleSHA256 != c.BundleSHA256 {
				return errors.New("typed_candidate_source_or_closure_changed")
			}
			caption := slices.IndexFunc(f.sources[:], func(s caption) bool { return s.id == c.SourceID })
			if caption < 0 || f.sources[caption].text != c.Text {
				return errors.New("typed_candidate_caption_not_reviewed")
			}
			// IDs are deliberately regenerated; metadata is not a scoring feature.
			id := string(rune('a' + j))
			input.Candidates = append(input.Candidates, shortclaim.Candidate{ID: id, Text: c.Text})
		}
		if _, e := shortclaim.Validate(input); e != nil {
			return errors.New("typed_text_or_candidate_bounds_invalid")
		}
	}
	return nil
}
