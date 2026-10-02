// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sourceinventory"
)

// No number, truth label, outcome, role or generated SourcePins API is needed
// by this projection. Full metadata bytes are pinned before this function.
type storedRoot struct {
	ID              string            `json:"id"`
	Prototype       string            `json:"prototype"`
	Core            string            `json:"core_template"`
	Function        string            `json:"function"`
	Code            string            `json:"code_sha256"`
	Normalized      string            `json:"normalized_code_sha256"`
	Bundle          string            `json:"source_bundle_sha256"`
	Components      []storedComponent `json:"semantic_components"`
	StandardImports []string          `json:"standard_imports"`
}

type storedComponent struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	SHA256     string `json:"sha256"`
	Normalized string `json:"normalized_behavior_sha256"`
}

type metadataProjection struct {
	Schema string `json:"schema"`
	Go     string `json:"runtime_go"`
	Legacy struct {
		Schema  string       `json:"schema"`
		Sources []storedRoot `json:"sources"`
	} `json:"legacy_truth"`
	Typed struct {
		Schema  string       `json:"schema"`
		Sources []storedRoot `json:"sources"`
	} `json:"typed_truth"`
}

// Scan JSON structure without converting integers to float64 or uint64.
// Reject duplicates, trailing values and excessive depth before projection.
func uniqueJSON(raw []byte) error {
	if len(raw) == 0 || len(raw) > maxBytes || !utf8.Valid(raw) {
		return fail("inventory60_metadata_json_invalid")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 64 {
			return fail("inventory60_metadata_json_invalid")
		}
		t, err := d.Token()
		if err != nil {
			return fail("inventory60_metadata_json_invalid")
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
				if e != nil || !ok || slices.Contains(keys, s) || len(keys) >= 4096 {
					return fail("inventory60_metadata_json_invalid")
				}
				keys = append(keys, s)
				if e := walk(depth + 1); e != nil {
					return e
				}
			}
			end, e := d.Token()
			if e != nil || end != json.Delim('}') {
				return fail("inventory60_metadata_json_invalid")
			}
		case '[':
			for d.More() {
				if e := walk(depth + 1); e != nil {
					return e
				}
			}
			end, e := d.Token()
			if e != nil || end != json.Delim(']') {
				return fail("inventory60_metadata_json_invalid")
			}
		default:
			return fail("inventory60_metadata_json_invalid")
		}
		return nil
	}
	if e := walk(0); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return fail("inventory60_metadata_json_invalid")
	}
	return nil
}

func descriptorString(s string) bool {
	return s != "" && len(s) <= 1024 && utf8.ValidString(s) && !strings.ContainsAny(s, "\x00\r\n")
}

func historicalRoots(raw []byte) ([]sourceinventory.HistoricalRoot, error) {
	if err := uniqueJSON(raw); err != nil {
		return nil, err
	}
	var m metadataProjection
	// Unknown historical fields are deliberately skipped. All required
	// descriptor fields are subsequently checked, with no numeric conversion.
	if json.Unmarshal(raw, &m) != nil || m.Schema != "riido-typed-truth-development-result-v1" || m.Go != "go1.27.1" || m.Legacy.Schema != "riido-behavior-probe-audit-v1" || m.Typed.Schema != "riido-typed-behavior-truth-v2" || len(m.Legacy.Sources) != 36 || len(m.Typed.Sources) != 24 {
		return nil, fail("inventory60_metadata_shape_invalid")
	}
	out := make([]sourceinventory.HistoricalRoot, 0, 60)
	var relations int
	for cohort, rows := range [2][]storedRoot{m.Legacy.Sources, m.Typed.Sources} {
		var ids []string
		for _, r := range rows {
			if !descriptorString(r.ID) || !descriptorString(r.Prototype) || !descriptorString(r.Core) || !hash(r.Code, 64) || !hash(r.Normalized, 64) || slices.Contains(ids, r.ID) {
				return nil, fail("inventory60_metadata_descriptor_invalid")
			}
			ids = append(ids, r.ID)
			name := [2]string{"legacy", "typed"}[cohort]
			root := sourceinventory.HistoricalRoot{Cohort: name, ID: r.ID, Prototype: r.Prototype, Core: r.Core, Function: r.Function, CodeSHA256: r.Code, NormalizedCodeSHA256: r.Normalized, BundleSHA256: r.Bundle, StandardImports: slices.Clone(r.StandardImports)}
			if cohort == 0 {
				if r.Function != "" || r.Bundle != "" || len(r.Components) != 0 || len(r.StandardImports) != 0 {
					return nil, fail("inventory60_metadata_descriptor_invalid")
				}
			} else {
				if !descriptorString(r.Function) || !hash(r.Bundle, 64) || len(r.Components) == 0 || len(r.Components) > 256 {
					return nil, fail("inventory60_metadata_descriptor_invalid")
				}
				for i, p := range r.StandardImports {
					if !descriptorString(p) || i > 0 && r.StandardImports[i-1] >= p {
						return nil, fail("inventory60_metadata_descriptor_invalid")
					}
				}
				for i, c := range r.Components {
					behavior := c.Kind == "function" || c.Kind == "method"
					if !descriptorString(c.ID) || !slices.Contains([]string{"function", "method", "type", "value"}, c.Kind) || !hash(c.SHA256, 64) || behavior && !hash(c.Normalized, 64) || !behavior && c.Normalized != "" || i > 0 && r.Components[i-1].ID >= c.ID {
						return nil, fail("inventory60_metadata_descriptor_invalid")
					}
					root.Components = append(root.Components, sourceinventory.Component{ID: c.ID, Kind: c.Kind, SHA256: c.SHA256, NormalizedBehaviorSHA256: c.Normalized})
				}
				relations += len(root.Components)
			}
			out = append(out, root)
		}
	}
	if relations != 204 {
		return nil, fail("inventory60_metadata_relation_count_invalid")
	}
	return out, nil
}
