// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only, text-free audit of explicitly pinned concept metadata.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const maxInput = 200 * 1024
const maxTotal = 2 * 1024 * 1024
const maxTranches = 10
const groupsPerTranche = 40
const maxGroups = maxTranches * groupsPerTranche
const seed = "claims-short-core-full-v1:concepts:1729"
const schema = "three-claims-short-core-full-provisional-concept-tranche-v1"
const correctionSerialization = "UTF-8 JSON, ensure_ascii=false, separators=(comma,colon), trailing LF; remove this correction object and restore old values to reconstruct prior metadata bytes."
const correctionNoPriorSerialization = "UTF-8 JSON, ensure_ascii=false, separators=(comma,colon), trailing LF. Restore each old field value then omit retained_provisional_metadata_correction to reconstruct predecessor exactly."
const correctionWithPriorSerialization = "UTF-8 JSON, ensure_ascii=false, separators=(comma,colon), trailing LF. Restore every old field value and prior_correction_object to reconstruct predecessor."

var tokenPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(;[a-z][a-z0-9_]*)*$`)

// Relation identifiers preserve source spelling, including CLI/UI abbreviations.
// This lexical rule still excludes prose, whitespace, punctuation, and messages.
var relationPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(;[A-Za-z][A-Za-z0-9_]*)*$`)
var pinPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Pair struct {
	KO string `json:"ko"`
	EN string `json:"en"`
}
type Slot struct {
	Slot string `json:"slot"`
	KO   string `json:"ko"`
	EN   string `json:"en"`
}
type Family struct {
	ID        string   `json:"family_id"`
	Ordinal   int      `json:"frame_ordinal"`
	KO        string   `json:"concept_ko"`
	EN        string   `json:"concept_en"`
	Phenomena []string `json:"linguistic_phenomena"`
	Extra     []string `json:"necessary_extra_slots"`
	Caution   string   `json:"brevity_caution"`
}
type Link struct {
	From     string `json:"from_family_id"`
	To       string `json:"to_family_id"`
	Relation string `json:"relation"`
}
type EssentialScopeLedger struct {
	Optional []string `json:"optional_detail_not_presupposed"`
	Source   string   `json:"core_scope_source"`
}
type ScaffoldProvenance struct {
	Ref          string `json:"common_provenance_ref"`
	Construction string `json:"frame_construction"`
}
type Group struct {
	Ordinal  int                   `json:"registration_ordinal"`
	ID       string                `json:"group_id"`
	Domain   string                `json:"domain"`
	Core     Pair                  `json:"core_event"`
	Slots    []Slot                `json:"necessary_slots"`
	Brevity  string                `json:"conceptual_brevity_hypothesis"`
	Families []Family              `json:"families"`
	Links    []Link                `json:"within_group_linkage"`
	Ledger   *EssentialScopeLedger `json:"essential_scope_ledger,omitempty"`
	Scaffold *ScaffoldProvenance   `json:"generation_scaffold_provenance,omitempty"`
}
type Relation struct {
	IDs         []string `json:"group_ids"`
	Ordinals    []int    `json:"registration_ordinals"`
	Type        string   `json:"relation_hypothesis"`
	Rationale   string   `json:"rationale"`
	Distinction string   `json:"necessary_distinction"`
	Status      string   `json:"review_status"`
	Dimensions  []string `json:"relation_dimensions,omitempty"`
}
type Scope struct {
	Full struct {
		Groups   int      `json:"groups"`
		Families int      `json:"bilingual_families"`
		Rows     int      `json:"rows"`
		Locales  []string `json:"locales"`
	} `json:"full_study_required"`
	Tranche struct {
		Range    []int `json:"registration_ordinals"`
		Groups   int   `json:"groups"`
		Families int   `json:"bilingual_families"`
		Rows     int   `json:"intended_rows"`
		Number   int   `json:"production_tranche_number"`
	} `json:"this_tranche"`
	Production struct {
		Tranches int `json:"tranches"`
		Groups   int `json:"groups_per_tranche"`
		Families int `json:"families_per_tranche"`
		Rows     int `json:"rows_per_tranche"`
	} `json:"full_production"`
	NotReduced bool              `json:"not_a_reduced_study"`
	Known      *KnownProvisional `json:"known_provisional_after_this_tranche,omitempty"`
}
type KnownProvisional struct {
	Groups   int `json:"groups"`
	Families int `json:"bilingual_families"`
	Rows     int `json:"intended_rows"`
}
type Boundary struct {
	Utterances  bool   `json:"contains_sample_utterances"`
	Targets     bool   `json:"contains_target_assignments"`
	Counts      bool   `json:"contains_expected_semantic_counts"`
	Triangle    bool   `json:"contains_fixed_semantic_triangle"`
	Bins        bool   `json:"contains_assigned_length_bins"`
	Splits      bool   `json:"contains_assigned_splits"`
	Reference   bool   `json:"contains_training_or_gold_or_reference_material"`
	Framing     string `json:"framing_meaning"`
	Translation string `json:"concept_translation_status"`
}
type Identity struct {
	Seed     string `json:"seed"`
	Group    string `json:"group"`
	Family   string `json:"family"`
	Prefixes string `json:"prefixes"`
	Ordinals string `json:"registration_ordinals"`
	Warning  string `json:"independence_warning"`
}
type Design struct {
	Minimal           string `json:"minimal_event"`
	Within            string `json:"within_group"`
	Across            string `json:"across_group"`
	Brief             string `json:"brief_feasibility"`
	Provenance        string `json:"common_group_provenance"`
	Allocation        string `json:"linkage_allocation"`
	Recruitment       string `json:"recruitment_constraint_from_first40,omitempty"`
	NoMandatoryReport string `json:"no_mandatory_report,omitempty"`
	EssentialScope    string `json:"shared_essential_scope_rule,omitempty"`
}
type Barriers struct {
	Editorial     string `json:"input_editorial"`
	Qualification string `json:"editor_qualification"`
	Sequence      string `json:"reference_sequence"`
	Holds         string `json:"holds"`
	Gates         string `json:"unchanged_gates"`
	Admission     string `json:"admission"`
}
type SourceBoundary struct {
	Allowed  []string            `json:"allowed_content_read"`
	Excluded string              `json:"excluded"`
	Lesson   string              `json:"aggregate_lesson"`
	Pilot    string              `json:"no_pilot_admission"`
	Pins     []DeclaredSourcePin `json:"allowed_source_pins,omitempty"`
}
type DeclaredSourcePin struct {
	Path  string `json:"path"`
	SHA   string `json:"sha256"`
	Bytes int    `json:"bytes"`
}
type GenerationProvenance struct {
	Role        string `json:"role"`
	Origin      string `json:"event_origin"`
	Composition string `json:"composition"`
	Ancestry    string `json:"external_ancestry_status"`
	Limit       string `json:"inspection_limit"`
	Policy      string `json:"scaffold_policy"`
}
type Delta struct {
	Pointer string          `json:"json_pointer"`
	Old     json.RawMessage `json:"old"`
	New     json.RawMessage `json:"new"`
}
type Correction struct {
	Batch         int                `json:"batch_number"`
	Created       string             `json:"created_at_utc"`
	Trigger       string             `json:"trigger"`
	Scope         string             `json:"scope"`
	PriorSHA      string             `json:"prior_sha256"`
	PriorBytes    int                `json:"prior_bytes"`
	Serialization string             `json:"serialization"`
	Deltas        []Delta            `json:"field_deltas"`
	Rows          int                `json:"input_rows_created_or_revised"`
	Prior         *InitialCorrection `json:"prior_correction_object,omitempty"`
	Reason        string             `json:"semantic_reason,omitempty"`
}
type InitialCorrection struct {
	Batch  int    `json:"batch_number"`
	Status string `json:"status"`
	Rows   int    `json:"input_rows_created_or_revised"`
}
type ExternalMetadata struct {
	Range []int  `json:"registration_ordinal_range"`
	Path  string `json:"path"`
	SHA   string `json:"sha256"`
}
type Universe struct {
	Own      []int              `json:"registered_own_ordinal_range"`
	Known    []int              `json:"currently_known_provisional_ordinal_range"`
	External []ExternalMetadata `json:"external_metadata_references"`
	Status   string             `json:"status"`
}
type Metadata struct {
	Schema     string                `json:"schema"`
	Created    string                `json:"created_at_utc"`
	Status     string                `json:"status"`
	Scope      Scope                 `json:"scope"`
	Boundary   Boundary              `json:"metadata_boundary"`
	Identity   Identity              `json:"identity_scheme"`
	Design     Design                `json:"concept_design"`
	Required   []string              `json:"required_before_wording"`
	Barriers   Barriers              `json:"future_production_barriers"`
	Groups     []Group               `json:"groups"`
	Relations  []Relation            `json:"across_group_relation_hypotheses"`
	Sources    SourceBoundary        `json:"source_boundary"`
	Correction *Correction           `json:"retained_provisional_metadata_correction,omitempty"`
	Universe   *Universe             `json:"relation_universe,omitempty"`
	Generation *GenerationProvenance `json:"generation_provenance,omitempty"`
}

// The raw tree rejects duplicate JSON keys and preserves property order solely
// for reconstructing the explicitly retained metadata predecessor.
type member struct {
	key    string
	keyRaw []byte
	value  *node
}
type node struct {
	raw    []byte
	object []member
	array  []*node
	kind   byte
}
type parser struct {
	data []byte
	at   int
}

func (p *parser) space() {
	for p.at < len(p.data) && strings.ContainsRune(" \t\n\r", rune(p.data[p.at])) {
		p.at++
	}
}
func (p *parser) stringEnd() (int, error) {
	s := p.at
	p.at++
	for p.at < len(p.data) {
		c := p.data[p.at]
		p.at++
		if c == '\\' {
			p.at++
			continue
		}
		if c == '"' {
			return s, nil
		}
	}
	return 0, errors.New("invalid_json")
}
func (p *parser) value(depth int) (*node, error) {
	if depth > 64 {
		return nil, errors.New("json_depth_limit")
	}
	p.space()
	start := p.at
	if start >= len(p.data) {
		return nil, errors.New("invalid_json")
	}
	n := &node{kind: p.data[p.at]}
	switch n.kind {
	case '{':
		p.at++
		p.space()
		seen := map[string]bool{}
		if p.at < len(p.data) && p.data[p.at] == '}' {
			p.at++
		} else {
			for {
				p.space()
				if p.at >= len(p.data) || p.data[p.at] != '"' {
					return nil, errors.New("invalid_json")
				}
				ks, err := p.stringEnd()
				if err != nil {
					return nil, err
				}
				kr := p.data[ks:p.at]
				var key string
				if json.Unmarshal(kr, &key) != nil {
					return nil, errors.New("invalid_json")
				}
				if seen[key] {
					return nil, errors.New("duplicate_json_key")
				}
				seen[key] = true
				p.space()
				if p.at >= len(p.data) || p.data[p.at] != ':' {
					return nil, errors.New("invalid_json")
				}
				p.at++
				child, err := p.value(depth + 1)
				if err != nil {
					return nil, err
				}
				n.object = append(n.object, member{key, kr, child})
				p.space()
				if p.at >= len(p.data) {
					return nil, errors.New("invalid_json")
				}
				c := p.data[p.at]
				p.at++
				if c == '}' {
					break
				}
				if c != ',' {
					return nil, errors.New("invalid_json")
				}
			}
		}
	case '[':
		p.at++
		p.space()
		if p.at < len(p.data) && p.data[p.at] == ']' {
			p.at++
		} else {
			for {
				child, err := p.value(depth + 1)
				if err != nil {
					return nil, err
				}
				n.array = append(n.array, child)
				p.space()
				if p.at >= len(p.data) {
					return nil, errors.New("invalid_json")
				}
				c := p.data[p.at]
				p.at++
				if c == ']' {
					break
				}
				if c != ',' {
					return nil, errors.New("invalid_json")
				}
			}
		}
	case '"':
		if _, err := p.stringEnd(); err != nil {
			return nil, err
		}
	default:
		for p.at < len(p.data) && !strings.ContainsRune(",]} \t\r\n", rune(p.data[p.at])) {
			p.at++
		}
	}
	n.raw = p.data[start:p.at]
	return n, nil
}
func tree(data []byte) (*node, error) {
	if !json.Valid(data) {
		return nil, errors.New("invalid_json")
	}
	p := parser{data: data}
	n, err := p.value(0)
	if err != nil {
		return nil, err
	}
	p.space()
	if p.at != len(data) {
		return nil, errors.New("invalid_json")
	}
	return n, nil
}
func (n *node) get(key string) *node {
	for _, m := range n.object {
		if m.key == key {
			return m.value
		}
	}
	return nil
}

var rawType = reflect.TypeOf(json.RawMessage{})

func shape(n *node, t reflect.Type) error {
	if t == rawType {
		return nil
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		if n.kind != '{' {
			return errors.New("schema_type")
		}
		known := map[string]reflect.StructField{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			known[strings.Split(f.Tag.Get("json"), ",")[0]] = f
		}
		for _, m := range n.object {
			f, ok := known[m.key]
			if !ok {
				return errors.New("schema_unknown_field")
			}
			// This sole typed null means no initial correction object.
			if t == reflect.TypeOf(Correction{}) && f.Name == "Prior" && string(m.value.raw) == "null" {
				delete(known, m.key)
				continue
			}
			if err := shape(m.value, f.Type); err != nil {
				return err
			}
			delete(known, m.key)
		}
		for _, f := range known {
			if !strings.Contains(f.Tag.Get("json"), ",omitempty") {
				return errors.New("schema_missing_field")
			}
		}
	case reflect.Slice:
		if n.kind != '[' {
			return errors.New("schema_type")
		}
		for _, v := range n.array {
			if err := shape(v, t.Elem()); err != nil {
				return err
			}
		}
	case reflect.String:
		if n.kind != '"' {
			return errors.New("schema_type")
		}
	case reflect.Bool:
		if string(n.raw) != "true" && string(n.raw) != "false" {
			return errors.New("schema_type")
		}
	case reflect.Int:
		var i int
		if json.Unmarshal(n.raw, &i) != nil {
			return errors.New("schema_type")
		}
	default:
		return errors.New("schema_type")
	}
	return nil
}
func compact(n *node, replacements map[*node][]byte, omit *node) []byte {
	if b, ok := replacements[n]; ok {
		return b
	}
	var b bytes.Buffer
	switch n.kind {
	case '{':
		b.WriteByte('{')
		first := true
		for _, m := range n.object {
			if m.value == omit {
				continue
			}
			if !first {
				b.WriteByte(',')
			}
			first = false
			b.Write(m.keyRaw)
			b.WriteByte(':')
			b.Write(compact(m.value, replacements, omit))
		}
		b.WriteByte('}')
	case '[':
		b.WriteByte('[')
		for i, v := range n.array {
			if i > 0 {
				b.WriteByte(',')
			}
			b.Write(compact(v, replacements, omit))
		}
		b.WriteByte(']')
	default:
		b.Write(n.raw)
	}
	return b.Bytes()
}
func pointer(root *node, path string) (*node, error) {
	n := root
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if n.kind == '[' {
			i, e := strconv.Atoi(part)
			if e != nil || i < 0 || i >= len(n.array) || part != strconv.Itoa(i) {
				return nil, errors.New("correction_pointer")
			}
			n = n.array[i]
		} else {
			n = n.get(part)
			if n == nil {
				return nil, errors.New("correction_pointer")
			}
		}
	}
	return n, nil
}
func correction(m Metadata, n *node) error {
	c := m.Correction
	if c == nil {
		return nil
	}
	if c.Batch != 1 || c.Rows != 0 || !pinPattern.MatchString(c.PriorSHA) || c.PriorBytes <= 0 || c.PriorBytes > maxInput || len(c.Deltas) == 0 || len(c.Deltas) > 32 {
		return errors.New("correction_provenance")
	}
	if c.Prior == nil {
		if c.Serialization != correctionSerialization && c.Serialization != correctionNoPriorSerialization {
			return errors.New("correction_provenance")
		}
	} else if c.Serialization != correctionWithPriorSerialization || c.Prior.Batch != 0 || c.Prior.Rows != 0 || !nonempty(c.Prior.Status) || !nonempty(c.Reason) {
		return errors.New("correction_provenance")
	}
	if strings.TrimSpace(c.Created) == "" || strings.TrimSpace(c.Trigger) == "" || strings.TrimSpace(c.Scope) == "" {
		return errors.New("correction_provenance")
	}
	replacements := map[*node][]byte{}
	seen := map[string]bool{}
	endpointMasks := map[int]int{}
	predecessorRelations := map[int]Relation{}
	for _, d := range c.Deltas {
		p := strings.Split(d.Pointer, "/")
		relationField := len(p) == 4 && p[0] == "" && p[1] == "across_group_relation_hypotheses"
		if !relationField && (len(p) < 5 || p[0] != "" || p[1] != "groups") {
			return errors.New("correction_pointer")
		}
		relationIndex := -1
		t := reflect.TypeOf("")
		switch {
		case relationField:
			var err error
			relationIndex, err = strconv.Atoi(p[2])
			if err != nil || relationIndex < 0 || relationIndex >= len(m.Relations) || p[2] != strconv.Itoa(relationIndex) {
				return errors.New("correction_pointer")
			}
			switch p[3] {
			case "group_ids":
				t = reflect.TypeOf([]string{})
			case "registration_ordinals":
				t = reflect.TypeOf([]int{})
			default:
				return errors.New("correction_field")
			}
		case len(p) == 6 && p[3] == "families":
			field := p[5]
			if field != "concept_ko" && field != "concept_en" && field != "brevity_caution" && field != "necessary_extra_slots" && field != "linguistic_phenomena" {
				return errors.New("correction_field")
			}
			if field == "necessary_extra_slots" || field == "linguistic_phenomena" {
				t = reflect.TypeOf([]string{})
			}
		case len(p) == 5 && p[3] == "core_event" && (p[4] == "ko" || p[4] == "en"):
		case len(p) == 6 && p[3] == "necessary_slots" && (p[5] == "ko" || p[5] == "en"):
		default:
			return errors.New("correction_field")
		}
		if seen[d.Pointer] {
			return errors.New("correction_duplicate_pointer")
		}
		seen[d.Pointer] = true
		current, err := pointer(n, d.Pointer)
		if err != nil {
			return err
		}
		old, err := tree(d.Old)
		if err != nil {
			return errors.New("correction_value")
		}
		neu, err := tree(d.New)
		if err != nil {
			return errors.New("correction_value")
		}
		if shape(old, t) != nil || shape(neu, t) != nil || !bytes.Equal(compact(current, nil, nil), compact(neu, nil, nil)) || bytes.Equal(compact(old, nil, nil), compact(neu, nil, nil)) {
			return errors.New("correction_value")
		}
		if relationField {
			prior, ok := predecessorRelations[relationIndex]
			if !ok {
				prior = m.Relations[relationIndex]
			}
			if p[3] == "group_ids" {
				var oldIDs []string
				if json.Unmarshal(d.Old, &oldIDs) != nil {
					return errors.New("correction_value")
				}
				prior.IDs = oldIDs
				endpointMasks[relationIndex] |= 1
			} else {
				var oldOrdinals []int
				if json.Unmarshal(d.Old, &oldOrdinals) != nil {
					return errors.New("correction_value")
				}
				prior.Ordinals = oldOrdinals
				endpointMasks[relationIndex] |= 2
			}
			predecessorRelations[relationIndex] = prior
		}
		replacements[current] = compact(old, nil, nil)
	}
	for index, prior := range predecessorRelations {
		if endpointMasks[index] != 3 {
			return errors.New("correction_relation_pair")
		}
		if m.Universe == nil {
			return errors.New("correction_relation_universe")
		}
		if len(prior.IDs) < 2 || len(prior.IDs) > maxGroups || len(prior.IDs) != len(prior.Ordinals) {
			return errors.New("correction_relation_endpoints")
		}
		ids := map[string]bool{}
		for i, id := range prior.IDs {
			o := prior.Ordinals[i]
			if o < m.Universe.Known[0] || o > m.Universe.Known[1] || id != groupID(o) || ids[id] {
				return errors.New("correction_relation_endpoints")
			}
			ids[id] = true
		}
		// Only these two paired endpoint arrays differ; all other relation/provenance
		// fields are copied unchanged and cannot be correction-pointer targets.
	}
	omit := n.get("retained_provisional_metadata_correction")
	if c.Prior != nil {
		replacements[omit] = compact(omit.get("prior_correction_object"), nil, nil)
		omit = nil
	}
	prior := append(compact(n, replacements, omit), '\n')
	if len(prior) != c.PriorBytes || hash(prior) != c.PriorSHA {
		return errors.New("correction_prior_pin")
	}
	return nil
}
func hash(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func groupID(n int) string { return "g_" + hash([]byte(fmt.Sprintf("%s:group:%04d", seed, n)))[:24] }
func familyID(n, f int) string {
	return "f_" + hash([]byte(fmt.Sprintf("%s:family:%04d:%d", seed, n, f)))[:24]
}
func nonempty(s string) bool { return utf8.ValidString(s) && strings.TrimSpace(s) != "" }
func tokens(a []string) bool {
	seen := map[string]bool{}
	for _, v := range a {
		if !tokenPattern.MatchString(v) || strings.Contains(v, ";") || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}
func rangeOK(a []int) bool { return len(a) == 2 && a[0] >= 1 && a[1] >= a[0] && a[1] <= 400 }
func validate(m Metadata, n *node) error {
	if m.Schema != schema {
		return errors.New("schema_name")
	}
	if m.Identity.Seed != seed {
		return errors.New("identity_seed")
	}
	b := m.Boundary
	if b.Utterances || b.Targets || b.Counts || b.Triangle || b.Bins || b.Splits || b.Reference {
		return errors.New("metadata_boundary")
	}
	s := m.Scope
	if s.Full.Groups != 400 || s.Full.Families != 1200 || s.Full.Rows != 2400 || !reflect.DeepEqual(s.Full.Locales, []string{"ko", "en"}) || s.Production.Tranches != 10 || s.Production.Groups != 40 || s.Production.Families != 120 || s.Production.Rows != 240 || !s.NotReduced {
		return errors.New("full_scope")
	}
	t := s.Tranche
	if !rangeOK(t.Range) || t.Range[1]-t.Range[0] != 39 || t.Groups != 40 || t.Families != 120 || t.Rows != 240 || t.Number < 1 || t.Number > 10 || t.Range[0] != (t.Number-1)*40+1 || len(m.Groups) != 40 {
		return errors.New("tranche_scope")
	}
	if k := s.Known; k != nil && (k.Groups != t.Range[1] || k.Families != k.Groups*3 || k.Rows != k.Groups*6) {
		return errors.New("known_provisional_scope")
	}
	for _, p := range m.Sources.Pins {
		if !nonempty(p.Path) || !pinPattern.MatchString(p.SHA) || p.Bytes <= 0 {
			return errors.New("declared_source_pin")
		}
	}
	seen := map[string]bool{}
	ord := map[int]bool{}
	for _, g := range m.Groups {
		if seen[g.ID] {
			return errors.New("duplicate_group_id")
		}
		seen[g.ID] = true
		if g.ID != groupID(g.Ordinal) {
			return errors.New("group_identity")
		}
		if ord[g.Ordinal] || g.Ordinal < t.Range[0] || g.Ordinal > t.Range[1] {
			return errors.New("group_ordinal")
		}
		ord[g.Ordinal] = true
		if !nonempty(g.Core.KO) || !nonempty(g.Core.EN) || !nonempty(g.Domain) || !nonempty(g.Brevity) {
			return errors.New("concept_description")
		}
		if len(g.Slots) != 2 || len(g.Families) != 3 || len(g.Links) != 2 {
			return errors.New("group_cardinality")
		}
		for _, slot := range g.Slots {
			if !nonempty(slot.KO) || !nonempty(slot.EN) || !tokenPattern.MatchString(slot.Slot) {
				return errors.New("necessary_slot")
			}
		}
		frames := map[int]bool{}
		own := map[string]bool{}
		for _, f := range g.Families {
			if seen[f.ID] {
				return errors.New("duplicate_family_id")
			}
			seen[f.ID] = true
			own[f.ID] = true
			if f.Ordinal < 1 || f.Ordinal > 3 || frames[f.Ordinal] || f.ID != familyID(g.Ordinal, f.Ordinal) {
				return errors.New("family_identity")
			}
			frames[f.Ordinal] = true
			if !nonempty(f.KO) || !nonempty(f.EN) || !nonempty(f.Caution) || len(f.Phenomena) == 0 || !tokens(f.Phenomena) || !tokens(f.Extra) {
				return errors.New("concept_description")
			}
		}
		connected := map[string]bool{g.Families[0].ID: true}
		edges := map[string]bool{}
		for _, e := range g.Links {
			if !own[e.From] || !own[e.To] || e.From == e.To {
				return errors.New("unresolved_within_link")
			}
			if !relationPattern.MatchString(e.Relation) {
				return errors.New("relation_type")
			}
			key := e.From + "|" + e.To
			if edges[key] {
				return errors.New("duplicate_within_link")
			}
			edges[key] = true
		}
		for i := 0; i < 3; i++ {
			for _, e := range g.Links {
				if connected[e.From] || connected[e.To] {
					connected[e.From] = true
					connected[e.To] = true
				}
			}
		}
		if len(connected) != 3 {
			return errors.New("within_graph_disconnected")
		}
	}
	for i := t.Range[0]; i <= t.Range[1]; i++ {
		if !ord[i] {
			return errors.New("group_ordinal_gap")
		}
	}
	for _, r := range m.Relations {
		if len(r.IDs) < 2 || len(r.IDs) != len(r.Ordinals) || len(r.IDs) > maxGroups || !relationPattern.MatchString(r.Type) || r.Status != "pending_independent_full_registry_lineage_review" || !nonempty(r.Rationale) || !nonempty(r.Distinction) || !tokens(r.Dimensions) {
			return errors.New("relation_schema")
		}
		seen := map[string]bool{}
		for i, id := range r.IDs {
			if id != groupID(r.Ordinals[i]) || seen[id] {
				return errors.New("relation_identity")
			}
			seen[id] = true
		}
	}
	if u := m.Universe; u != nil {
		if !reflect.DeepEqual(u.Own, t.Range) || !rangeOK(u.Known) || u.Known[0] != 1 || u.Known[1] != t.Range[1] || u.Known[1]%groupsPerTranche != 0 || len(u.External) > maxTranches-1 || !nonempty(u.Status) {
			return errors.New("relation_universe")
		}
		for _, r := range m.Relations {
			for _, o := range r.Ordinals {
				if o < u.Known[0] || o > u.Known[1] {
					return errors.New("relation_outside_declared_universe")
				}
			}
		}
		priorRanges := map[int]bool{}
		for _, e := range u.External {
			if !rangeOK(e.Range) || e.Range[1]-e.Range[0] != groupsPerTranche-1 || (e.Range[0]-1)%groupsPerTranche != 0 || e.Range[0] < u.Known[0] || e.Range[1] > u.Known[1] || e.Range[1] >= t.Range[0] || priorRanges[e.Range[0]] || !pinPattern.MatchString(e.SHA) || !relative(e.Path) {
				return errors.New("external_metadata_declaration")
			}
			priorRanges[e.Range[0]] = true
		}
		if len(u.External) != t.Number-1 {
			return errors.New("relation_universe_external_count")
		}
	}
	return correction(m, n)
}
func readBounded(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, maxInput+1))
	if err != nil {
		return nil, errors.New("input_read")
	}
	if len(b) > maxInput {
		return nil, errors.New("input_size_limit")
	}
	return b, nil
}
func decodePinned(b []byte, pin string) (Metadata, error) {
	var m Metadata
	if len(b) > maxInput {
		return m, errors.New("input_size_limit")
	}
	if !pinPattern.MatchString(pin) {
		return m, errors.New("pin_format")
	}
	if hash(b) != pin {
		return m, errors.New("pin_mismatch")
	}
	if !utf8.Valid(b) {
		return m, errors.New("metadata_utf8")
	}
	n, err := tree(b)
	if err != nil {
		return m, err
	}
	if err = shape(n, reflect.TypeOf(m)); err != nil {
		return m, err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(&m) != nil {
		return m, errors.New("schema_decode")
	}
	if err = validate(m, n); err != nil {
		return m, err
	}
	return m, nil
}
func relative(p string) bool {
	return p != "" && !filepath.IsAbs(p) && filepath.Clean(p) == p && p != "." && p != ".." && !strings.HasPrefix(p, ".."+string(filepath.Separator)) && filepath.Ext(p) == ".json"
}

type Input struct {
	Path        string `json:"path"`
	SHA         string `json:"sha256"`
	Bytes       int    `json:"bytes"`
	Range       []int  `json:"registration_ordinal_range"`
	Declaration string `json:"relation_universe_declaration"`
	Correction  string `json:"retained_correction_check"`
}
type loaded struct {
	meta  Metadata
	input Input
}

// Count and stat-only byte preflight precede every content read, hash and decode.
// Declarations are never paths to auto-open: only explicit -input files load.
type prepared struct {
	path, pin, absolute string
	bytes               int
	info                os.FileInfo
}

func inputCount(n int) error {
	if n < 1 || n > maxTranches {
		return errors.New("input_count_1_to_10_only")
	}
	return nil
}
func inputBudget(a []prepared) error {
	total := 0
	for _, p := range a {
		if p.bytes < 0 || p.bytes > maxTotal-total {
			return errors.New("total_input_size_limit")
		}
		total += p.bytes
	}
	for _, p := range a {
		if p.bytes == 0 || p.bytes > maxInput {
			return errors.New("input_size_limit")
		}
	}
	return nil
}
func loadInputs(inputs []string, preparer func(string) (prepared, error), loader func(prepared) (loaded, error)) ([]loaded, error) {
	if err := inputCount(len(inputs)); err != nil {
		return nil, err
	}
	plans := make([]prepared, 0, len(inputs))
	paths, pins := map[string]bool{}, map[string]bool{}
	for _, spec := range inputs {
		p, err := preparer(spec)
		if err != nil {
			return nil, err
		}
		if paths[p.path] {
			return nil, errors.New("duplicate_input_path")
		}
		if pins[p.pin] {
			return nil, errors.New("duplicate_input_pin")
		}
		paths[p.path], pins[p.pin] = true, true
		plans = append(plans, p)
	}
	if err := inputBudget(plans); err != nil {
		return nil, err
	}
	all := make([]loaded, 0, len(plans))
	for _, p := range plans {
		l, err := loader(p)
		if err != nil {
			return nil, err
		}
		if l.input.Bytes != p.bytes || l.input.Path != p.path || l.input.SHA != p.pin {
			return nil, errors.New("preflight_input_changed")
		}
		all = append(all, l)
	}
	return all, nil
}
func prepare(spec string) (prepared, error) {
	var p prepared
	parts := strings.Split(spec, "=")
	if len(parts) != 2 || !relative(parts[0]) {
		return p, errors.New("relative_metadata_path_required")
	}
	if !pinPattern.MatchString(parts[1]) {
		return p, errors.New("pin_format")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return p, errors.New("working_directory")
	}
	resolved, err := filepath.EvalSymlinks(parts[0])
	if err != nil {
		return p, errors.New("input_open")
	}
	absolute, err := filepath.Abs(resolved)
	if err != nil {
		return p, errors.New("input_open")
	}
	rel, err := filepath.Rel(cwd, absolute)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return p, errors.New("path_outside_working_directory")
	}
	stat, err := os.Stat(absolute)
	if err != nil || !stat.Mode().IsRegular() {
		return p, errors.New("regular_metadata_file_required")
	}
	if stat.Size() < 0 || stat.Size() > maxTotal {
		return p, errors.New("total_input_size_limit")
	}
	p = prepared{path: parts[0], pin: parts[1], absolute: absolute, bytes: int(stat.Size()), info: stat}
	return p, nil
}
func load(p prepared) (loaded, error) {
	var out loaded
	f, err := os.Open(p.absolute)
	if err != nil {
		return out, errors.New("input_open")
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		return out, errors.New("regular_metadata_file_required")
	}
	if p.info == nil || !os.SameFile(p.info, stat) || stat.Size() != int64(p.bytes) {
		return out, errors.New("preflight_input_changed")
	}
	b, err := readBounded(f)
	if err != nil {
		return out, err
	}
	if len(b) != p.bytes {
		return out, errors.New("preflight_input_changed")
	}
	m, err := decodePinned(b, p.pin)
	if err != nil {
		return out, err
	}
	decl := "absent_unknown"
	if m.Universe != nil {
		decl = "present_provisional_supplied_coverage_only"
	}
	corr := "absent"
	if m.Correction != nil {
		corr = "predecessor_sha_and_bytes_reconstructed_exactly"
	}
	out = loaded{m, Input{p.path, p.pin, len(b), m.Scope.Tranche.Range, decl, corr}}
	return out, nil
}

type Lengths struct {
	Groups    int         `json:"group_core_names"`
	Min       int         `json:"minimum_runes"`
	Median    float64     `json:"median_runes"`
	P90       int         `json:"p90_runes"`
	Max       int         `json:"maximum_runes"`
	Histogram map[int]int `json:"rune_length_frequency"`
}

func lengths(a []int) Lengths {
	sort.Ints(a)
	h := map[int]int{}
	for _, n := range a {
		h[n]++
	}
	mid := float64(a[len(a)/2])
	if len(a)%2 == 0 {
		mid = float64(a[len(a)/2-1]+a[len(a)/2]) / 2
	}
	return Lengths{len(a), a[0], mid, a[(len(a)*9+9)/10-1], a[len(a)-1], h}
}

type DuplicateNames struct {
	KOClusters   int    `json:"ko_name_clusters"`
	KOGroups     int    `json:"ko_groups_in_clusters"`
	ENClusters   int    `json:"en_name_clusters"`
	ENGroups     int    `json:"en_groups_in_clusters"`
	PairClusters int    `json:"bilingual_pair_clusters"`
	PairGroups   int    `json:"groups_in_bilingual_pair_clusters"`
	Constraint   string `json:"constraint"`
}

func duplicate(a map[string]int) (int, int) {
	c, g := 0, 0
	for _, n := range a {
		if n > 1 {
			c++
			g += n
		}
	}
	return c, g
}

type ComponentReport struct {
	Count      int         `json:"candidate_review_component_count"`
	Singletons int         `json:"singleton_groups"`
	Max        int         `json:"maximum_groups_per_candidate_component"`
	Sizes      map[int]int `json:"candidate_component_size_frequency"`
	Meaning    string      `json:"meaning"`
}
type Caps struct {
	Tranches int `json:"maximum_supplied_tranches"`
	Each     int `json:"maximum_bytes_per_input"`
	Total    int `json:"maximum_bytes_all_inputs"`
}
type Report struct {
	SourceSchema   string                    `json:"source_metadata_schema"`
	Caps           Caps                      `json:"scope_caps"`
	Provenance     string                    `json:"input_validation_scope"`
	Schema         string                    `json:"schema"`
	Status         string                    `json:"status"`
	Inputs         []Input                   `json:"pinned_metadata_inputs"`
	Groups         int                       `json:"registered_groups"`
	Families       int                       `json:"registered_bilingual_families"`
	Range          []int                     `json:"global_ordinal_range"`
	Within         int                       `json:"within_group_link_edges"`
	Relations      int                       `json:"cross_group_hypotheses"`
	WithinTypes    map[string]int            `json:"within_link_type_counts"`
	RelationTypes  map[string]map[string]int `json:"hypothesis_type_by_review_status_counts"`
	Dimensions     map[string]int            `json:"relation_dimension_incidence"`
	Phenomena      map[string]int            `json:"family_phenomenon_incidence"`
	FramePhenomena map[int]map[string]int    `json:"frame_ordinal_phenomenon_incidence"`
	DirectReport   map[int]int               `json:"frame_ordinal_explicit_report_tag_frequency"`
	CoreLengths    map[string]Lengths        `json:"core_name_rune_lengths_descriptive_only"`
	Duplicates     DuplicateNames            `json:"duplicate_core_name_constraints"`
	Unresolved     map[string]int            `json:"unresolved_relation_counts"`
	Components     ComponentReport           `json:"candidate_relation_review_components"`
	Limits         []string                  `json:"limitations"`
}

func audit(inputs []loaded) (Report, error) {
	r := Report{SourceSchema: schema, Caps: Caps{maxTranches, maxInput, maxTotal}, Provenance: "Only explicitly supplied pinned metadata inputs are opened. Each declared universe is covered exactly by its own tranche and verified supplied external references; source-boundary pins remain declarations only.", Schema: "concept-registry-audit-text-free-v2", Status: "PASS_METADATA_ONLY_PROVISIONAL", WithinTypes: map[string]int{}, RelationTypes: map[string]map[string]int{}, Dimensions: map[string]int{}, Phenomena: map[string]int{}, FramePhenomena: map[int]map[string]int{1: {}, 2: {}, 3: {}}, DirectReport: map[int]int{1: 0, 2: 0, 3: 0}, CoreLengths: map[string]Lengths{}, Unresolved: map[string]int{"within_links": 0, "cross_group_relations": 0, "external_metadata_declarations": 0}}
	if err := inputCount(len(inputs)); err != nil {
		return r, err
	}
	plans := make([]prepared, 0, len(inputs))
	for _, l := range inputs {
		plans = append(plans, prepared{path: l.input.Path, pin: l.input.SHA, bytes: l.input.Bytes})
	}
	if err := inputBudget(plans); err != nil {
		return r, err
	}
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].meta.Scope.Tranche.Range[0] < inputs[j].meta.Scope.Tranche.Range[0]
	})
	groups := map[string]Group{}
	ord := map[int]bool{}
	families := map[string]bool{}
	pins := map[string]loaded{}
	inputPaths := map[string]bool{}
	var ko, en []int
	kn, enNames, pairs := map[string]int{}, map[string]int{}, map[string]int{}
	for _, l := range inputs {
		r.Inputs = append(r.Inputs, l.input)
		if _, ok := pins[l.input.SHA]; ok {
			return r, errors.New("duplicate_input_pin")
		}
		pins[l.input.SHA] = l
		if inputPaths[l.input.Path] {
			return r, errors.New("duplicate_input_path")
		}
		inputPaths[l.input.Path] = true
		for _, g := range l.meta.Groups {
			if _, ok := groups[g.ID]; ok {
				return r, errors.New("duplicate_group_id")
			}
			groups[g.ID] = g
			if ord[g.Ordinal] {
				return r, errors.New("duplicate_group_ordinal")
			}
			ord[g.Ordinal] = true
			r.Groups++
			ko = append(ko, utf8.RuneCountInString(g.Core.KO))
			en = append(en, utf8.RuneCountInString(g.Core.EN))
			k := strings.Join(strings.Fields(g.Core.KO), " ")
			e := strings.Join(strings.Fields(g.Core.EN), " ")
			kn[k]++
			enNames[e]++
			pairs[k+"\x00"+e]++
			for _, f := range g.Families {
				if families[f.ID] {
					return r, errors.New("duplicate_family_id")
				}
				families[f.ID] = true
				r.Families++
				for _, p := range f.Phenomena {
					r.Phenomena[p]++
					r.FramePhenomena[f.Ordinal][p]++
					if p == "explicit_report" {
						r.DirectReport[f.Ordinal]++
					}
				}
			}
			for _, e := range g.Links {
				r.Within++
				r.WithinTypes[e.Relation]++
			}
		}
	}
	if r.Groups != len(inputs)*groupsPerTranche || r.Groups > maxGroups {
		return r, errors.New("join_count_40_steps_to_400_only")
	}
	for i := 1; i <= r.Groups; i++ {
		if !ord[i] {
			return r, errors.New("join_global_ordinal_gap")
		}
	}
	r.Range = []int{1, r.Groups}
	parent := map[string]string{}
	for id := range groups {
		parent[id] = id
	}
	var root func(string) string
	root = func(id string) string {
		if parent[id] != id {
			parent[id] = root(parent[id])
		}
		return parent[id]
	}
	union := func(a, b string) {
		a = root(a)
		b = root(b)
		if a != b {
			if a > b {
				a, b = b, a
			}
			parent[b] = a
		}
	}
	for _, l := range inputs {
		declared := map[int]bool{}
		if u := l.meta.Universe; u != nil {
			if u.Known[1] > r.Groups {
				return r, errors.New("declared_universe_not_supplied")
			}
			if !reflect.DeepEqual(u.Known, []int{1, l.meta.Scope.Tranche.Range[1]}) {
				return r, errors.New("relation_universe")
			}
			for _, g := range l.meta.Groups {
				declared[g.Ordinal] = true
			}
			extPins, extPaths := map[string]bool{}, map[string]bool{}
			for _, ext := range u.External {
				other, ok := pins[ext.SHA]
				if !ok || !reflect.DeepEqual(other.meta.Scope.Tranche.Range, ext.Range) {
					return r, errors.New("unresolved_external_metadata_declaration")
				}
				if ext.SHA == l.input.SHA {
					return r, errors.New("external_metadata_self_reference")
				}
				if ext.Path != other.input.Path {
					return r, errors.New("external_metadata_path")
				}
				if extPins[ext.SHA] || extPaths[ext.Path] {
					return r, errors.New("duplicate_external_metadata_declaration")
				}
				extPins[ext.SHA], extPaths[ext.Path] = true, true
				for _, g := range other.meta.Groups {
					if declared[g.Ordinal] {
						return r, errors.New("overlapping_external_metadata_declaration")
					}
					declared[g.Ordinal] = true
				}
			}
			// Forty-group aligned ranges and at most nine references bound this loop.
			for o := u.Known[0]; o <= u.Known[1]; o++ {
				if !declared[o] || !ord[o] {
					return r, errors.New("declared_universe_incomplete_coverage")
				}
			}
			if len(declared) != u.Known[1]-u.Known[0]+1 {
				return r, errors.New("declared_universe_incomplete_coverage")
			}
		}
		for _, g := range l.meta.Groups {
			for _, e := range g.Links {
				if !families[e.From] || !families[e.To] {
					return r, errors.New("unresolved_within_link")
				}
			}
		}
		for _, e := range l.meta.Relations {
			r.Relations++
			if r.RelationTypes[e.Type] == nil {
				r.RelationTypes[e.Type] = map[string]int{}
			}
			r.RelationTypes[e.Type][e.Status]++
			for _, d := range e.Dimensions {
				r.Dimensions[d]++
			}
			for i, id := range e.IDs {
				if l.meta.Universe != nil && !declared[e.Ordinals[i]] {
					return r, errors.New("relation_outside_supplied_declared_universe")
				}
				g, ok := groups[id]
				if !ok || g.Ordinal != e.Ordinals[i] {
					return r, errors.New("unresolved_cross_group_relation")
				}
				if i > 0 {
					union(e.IDs[0], id)
				}
			}
		}
	}
	sizes := map[string]int{}
	for id := range groups {
		sizes[root(id)]++
	}
	c := ComponentReport{Sizes: map[int]int{}, Meaning: "Undirected connectivity of recorded typed hypotheses is a candidate review set only. It is not confirmed lineage, shared ancestry, semantic independence, or effective sample size."}
	for _, n := range sizes {
		c.Count++
		c.Sizes[n]++
		if n == 1 {
			c.Singletons++
		}
		if n > c.Max {
			c.Max = n
		}
	}
	r.Components = c
	r.CoreLengths["ko"] = lengths(ko)
	r.CoreLengths["en"] = lengths(en)
	kc, kg := duplicate(kn)
	ec, eg := duplicate(enNames)
	pc, pg := duplicate(pairs)
	r.Duplicates = DuplicateNames{kc, kg, ec, eg, pc, pg, "Exact names after whitespace folding flag review constraints only; neither uniqueness nor duplication settles event identity or independence."}
	r.Limits = []string{"Concept metadata only: no utterances, judgments, labels, assigned bins/splits, model calls, Fits, admission, or target-validity inference.", "Core-name rune lengths describe metadata names only; they do not establish short-utterance feasibility, bilingual naturalness, or later observed length.", "Phenomenon and explicit_report counts are descriptive registered tags by family/frame ordinal, not semantic labels, quotas, prevalence, or locale reference values.", "Every recorded relation type and pending hypothesis status is retained in counts; directional endpoint order is checked but components intentionally ignore direction for review connectivity only.", "Relations may be incomplete; absent universe declarations remain absent/unknown. Only supplied pinned group records resolve actual endpoints; present declarations require exact own-plus-external coverage.", "Strict keys and false boundary declarations reject message/utterance schemas; the tool does not semantically prove that a prose concept description contains no realization.", "Correction reconstruction verifies metadata predecessor bytes and pin only; it does not certify editorial quality, external source independence, or study validity."}
	return r, nil
}

type specs []string

func (s *specs) String() string     { return strings.Join(*s, ",") }
func (s *specs) Set(v string) error { *s = append(*s, v); return nil }
func encode(v any) []byte           { b, _ := json.MarshalIndent(v, "", "  "); return append(b, '\n') }
func synthetic(start int) Metadata {
	var m Metadata
	m.Schema = schema
	m.Created = "2026-10-07T00:00:00Z"
	m.Status = "PROVISIONAL_CONCEPT_METADATA_ONLY"
	m.Identity = Identity{seed, "formula", "formula", "opaque namespaces", "global ordinals", "uniqueness is bookkeeping only"}
	m.Scope.Full.Groups = 400
	m.Scope.Full.Families = 1200
	m.Scope.Full.Rows = 2400
	m.Scope.Full.Locales = []string{"ko", "en"}
	m.Scope.Tranche.Range = []int{start, start + 39}
	m.Scope.Tranche.Groups = 40
	m.Scope.Tranche.Families = 120
	m.Scope.Tranche.Rows = 240
	m.Scope.Tranche.Number = (start-1)/40 + 1
	m.Scope.Production.Tranches = 10
	m.Scope.Production.Groups = 40
	m.Scope.Production.Families = 120
	m.Scope.Production.Rows = 240
	m.Scope.NotReduced = true
	m.Required = []string{"independent metadata review"}
	m.Sources.Allowed = []string{"synthetic_fixture_only"}
	m.Relations = []Relation{}
	for i := start; i < start+40; i++ {
		g := Group{Ordinal: i, ID: groupID(i), Domain: "synthetic", Core: Pair{fmt.Sprintf("합성개념%03d", i), fmt.Sprintf("synthetic concept %03d", i)}, Brevity: "hypothesis only", Slots: []Slot{{"object", "합성대상", "synthetic object"}, {"operation", "합성작업", "synthetic operation"}}}
		for f := 1; f <= 3; f++ {
			tags := []string{"synthetic_phenomenon"}
			if f == 1 {
				tags = append(tags, "explicit_report")
			}
			g.Families = append(g.Families, Family{familyID(i, f), f, "SYNTHETIC_KO_CONCEPT_SENTINEL", "SYNTHETIC_EN_CONCEPT_SENTINEL", tags, []string{}, "hypothesis only"})
		}
		g.Links = []Link{{familyID(i, 1), familyID(i, 2), "same_core_event;discourse_variant"}, {familyID(i, 1), familyID(i, 3), "same_core_event;discourse_variant"}}
		m.Groups = append(m.Groups, g)
	}
	return m
}
func synLoad(m Metadata) (loaded, error) {
	b := encode(m)
	pin := hash(b)
	decoded, err := decodePinned(b, pin)
	return loaded{decoded, Input{"synthetic.json", pin, len(b), m.Scope.Tranche.Range, "absent_unknown", "absent"}}, err
}

// Sequential owned fixtures use the canonical existing external-reference schema.
func synTranches(count int) ([]loaded, error) {
	if err := inputCount(count); err != nil {
		return nil, err
	}
	out := make([]loaded, 0, count)
	for i := 1; i <= count; i++ {
		m := synthetic((i-1)*groupsPerTranche + 1)
		m.Scope.Known = &KnownProvisional{i * 40, i * 120, i * 240}
		m.Universe = &Universe{m.Scope.Tranche.Range, []int{1, i * 40}, []ExternalMetadata{}, "PROVISIONAL_NOT_CONFIRMED_LINEAGE"}
		for _, prior := range out {
			m.Universe.External = append(m.Universe.External, ExternalMetadata{prior.meta.Scope.Tranche.Range, prior.input.Path, prior.input.SHA})
		}
		if i > 1 {
			m.Relations = []Relation{{[]string{groupID(1), groupID((i-1)*40 + 1)}, []int{1, (i-1)*40 + 1}, "possible_shared_CLI_ancestor;not_established", "synthetic rationale", "synthetic distinction", "pending_independent_full_registry_lineage_review", []string{"scaffold_overlap"}}}
		}
		l, err := synLoad(m)
		if err != nil {
			return nil, err
		}
		l.input.Path = fmt.Sprintf("synthetic-tranche%02d.json", i)
		l.input.Declaration = "present_provisional_supplied_coverage_only"
		out = append(out, l)
	}
	return out, nil
}
func selfTest() (any, error) {
	type check struct {
		Name     string `json:"name"`
		Status   string `json:"status"`
		Expected string `json:"expected"`
	}
	checks := []check{}
	add := func(name, want string, fn func() error) error {
		err := fn()
		got := "PASS"
		if err != nil {
			got = err.Error()
		}
		if got != want {
			return fmt.Errorf("synthetic_check_failed:%s", name)
		}
		checks = append(checks, check{name, "PASS", want})
		return nil
	}
	for _, count := range []int{1, 2, 3, 10} {
		if err := add(fmt.Sprintf("canonical_%d_tranches_%d_groups", count, count*40), "PASS", func() error {
			a, err := synTranches(count)
			if err != nil {
				return err
			}
			r, err := audit(a)
			if err != nil {
				return err
			}
			reversed := append([]loaded(nil), a...)
			for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
				reversed[i], reversed[j] = reversed[j], reversed[i]
			}
			s, err := audit(reversed)
			if err != nil {
				return err
			}
			if !bytes.Equal(encode(r), encode(s)) || r.Groups != count*40 || r.Families != count*120 || r.Within != count*80 || r.DirectReport[1] != count*40 || r.Components.Max != count || r.Components.Count != count*40-(count-1) {
				return errors.New("aggregate_or_order_or_candidate_semantics")
			}
			if bytes.Contains(encode(r), []byte("SYNTHETIC_")) {
				return errors.New("text_leak")
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	for _, tc := range []struct {
		name, want string
		sizes      []int
	}{
		{"over_2MiB_stat_scope_before_content_loader", "total_input_size_limit", []int{209716, 209716, 209716, 209716, 209716, 209716, 209716, 209716, 209716, 209716}},
		{"over_200KiB_stat_scope_before_content_loader", "input_size_limit", []int{maxInput + 1}},
		{"exact_10_by_200KiB_stat_budget", "PASS", []int{maxInput, maxInput, maxInput, maxInput, maxInput, maxInput, maxInput, maxInput, maxInput, maxInput}},
	} {
		if err := add(tc.name, tc.want, func() error {
			calls := 0
			plans := make([]prepared, len(tc.sizes))
			for i, size := range tc.sizes {
				plans[i] = prepared{path: fmt.Sprintf("synthetic%02d.json", i), pin: fmt.Sprintf("%064x", i+1), bytes: size}
			}
			if tc.want == "PASS" {
				return inputBudget(plans)
			}
			specs := make([]string, len(plans))
			next := 0
			_, err := loadInputs(specs, func(string) (prepared, error) { p := plans[next]; next++; return p, nil }, func(prepared) (loaded, error) { calls++; return loaded{}, errors.New("loader_must_not_run") })
			if calls != 0 {
				return errors.New("content_loader_before_budget")
			}
			return err
		}); err != nil {
			return nil, err
		}
	}
	if err := add("malformed_final_spec_before_any_content_loader", "pin_format", func() error {
		next, calls := 0, 0
		_, err := loadInputs([]string{"fake1", "bad"}, func(string) (prepared, error) {
			next++
			if next == 2 {
				return prepared{}, errors.New("pin_format")
			}
			return prepared{path: "synthetic.json", pin: strings.Repeat("1", 64), bytes: 1}, nil
		}, func(prepared) (loaded, error) { calls++; return loaded{}, errors.New("loader_must_not_run") })
		if calls != 0 {
			return errors.New("content_loader_before_preflight_complete")
		}
		return err
	}); err != nil {
		return nil, err
	}
	for _, tc := range []struct {
		name, want string
		mutate     func([]loaded)
	}{
		{"duplicate_cross_file_group_id", "duplicate_group_id", func(a []loaded) { a[1].meta.Groups[0].ID = a[0].meta.Groups[0].ID }},
		{"duplicate_cross_file_family_id", "duplicate_family_id", func(a []loaded) { a[1].meta.Groups[0].Families[0].ID = a[0].meta.Groups[0].Families[0].ID }},
		{"missing_prior_external_pin", "declared_universe_incomplete_coverage", func(a []loaded) { a[2].meta.Universe.External = a[2].meta.Universe.External[:1] }},
		{"external_wrong_pin", "unresolved_external_metadata_declaration", func(a []loaded) { a[2].meta.Universe.External[1].SHA = strings.Repeat("0", 64) }},
		{"external_wrong_path", "external_metadata_path", func(a []loaded) { a[2].meta.Universe.External[1].Path = "wrong.json" }},
		{"duplicate_external_declaration", "duplicate_external_metadata_declaration", func(a []loaded) {
			a[2].meta.Universe.External = append(a[2].meta.Universe.External, a[2].meta.Universe.External[0])
		}},
		{"earlier_universe_does_not_expand_to_supplied_later_records", "relation_outside_supplied_declared_universe", func(a []loaded) {
			a[0].meta.Relations = []Relation{{[]string{groupID(1), groupID(81)}, []int{1, 81}, "possible_shared_ancestor", "synthetic rationale", "synthetic distinction", "pending_independent_full_registry_lineage_review", nil}}
		}},
	} {
		if err := add(tc.name, tc.want, func() error {
			a, err := synTranches(3)
			if err != nil {
				return err
			}
			tc.mutate(a)
			_, err = audit(a)
			return err
		}); err != nil {
			return nil, err
		}
	}
	if err := add("phantom_declared_universe_has_no_actual_records", "declared_universe_not_supplied", func() error {
		a, err := synTranches(2)
		if err != nil {
			return err
		}
		a[1].meta.Universe.Known = []int{1, 120}
		_, err = audit(a)
		return err
	}); err != nil {
		return nil, err
	}
	if err := add("skipped_middle_tranche", "join_global_ordinal_gap", func() error {
		a, err := synTranches(3)
		if err != nil {
			return err
		}
		_, err = audit([]loaded{a[0], a[2]})
		return err
	}); err != nil {
		return nil, err
	}
	if err := add("missing_first_tranche", "join_global_ordinal_gap", func() error {
		a, err := synTranches(2)
		if err != nil {
			return err
		}
		_, err = audit(a[1:])
		return err
	}); err != nil {
		return nil, err
	}
	if err := add("relation_400_actual_pinned_endpoints_supported", "PASS", func() error {
		a, err := synTranches(10)
		if err != nil {
			return err
		}
		m := a[9].meta
		r := Relation{Type: "possible_shared_UI_ancestor;not_established", Rationale: "synthetic rationale", Distinction: "synthetic distinction", Status: "pending_independent_full_registry_lineage_review", Dimensions: []string{"event"}}
		for o := 1; o <= 400; o++ {
			r.IDs = append(r.IDs, groupID(o))
			r.Ordinals = append(r.Ordinals, o)
		}
		m.Relations = []Relation{r}
		l, err := synLoad(m)
		if err != nil {
			return err
		}
		l.input.Path = a[9].input.Path
		a[9] = l
		report, err := audit(a)
		if err != nil {
			return err
		}
		if report.Components.Count != 1 || report.Components.Max != 400 {
			return errors.New("candidate_component_count")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	for _, field := range []string{"path", "sha", "bytes"} {
		if err := add("loader_preflight_"+field+"_mismatch", "preflight_input_changed", func() error {
			plan := prepared{path: "synthetic.json", pin: strings.Repeat("1", 64), bytes: 1}
			_, err := loadInputs([]string{"fake"}, func(string) (prepared, error) { return plan, nil }, func(prepared) (loaded, error) {
				in := Input{Path: plan.path, SHA: plan.pin, Bytes: plan.bytes}
				switch field {
				case "path":
					in.Path = "different.json"
				case "sha":
					in.SHA = strings.Repeat("2", 64)
				case "bytes":
					in.Bytes = 2
				}
				return loaded{input: in}, nil
			})
			return err
		}); err != nil {
			return nil, err
		}
	}
	base := synthetic(1)
	for _, n := range []int{0, 11} {
		if err := add(fmt.Sprintf("input_count_%d_rejected_before_any_loader", n), "input_count_1_to_10_only", func() error {
			calls := 0
			_, e := loadInputs(make([]string, n), func(string) (prepared, error) { calls++; return prepared{}, errors.New("preparer_must_not_run") }, func(prepared) (loaded, error) { calls++; return loaded{}, errors.New("loader_must_not_run") })
			if calls != 0 {
				return errors.New("loader_called_before_input_bound")
			}
			return e
		}); err != nil {
			return nil, err
		}
	}
	test := func(name, want string, mutate func(*Metadata)) error {
		return add(name, want, func() error { m := synthetic(1); mutate(&m); _, e := synLoad(m); return e })
	}
	if err := add("valid40_and_text_free", "PASS", func() error {
		l, e := synLoad(base)
		if e != nil {
			return e
		}
		r, e := audit([]loaded{l})
		if e != nil {
			return e
		}
		b := encode(r)
		if bytes.Contains(b, []byte("SYNTHETIC_KO_CONCEPT_SENTINEL")) || bytes.Contains(b, []byte("SYNTHETIC_EN_CONCEPT_SENTINEL")) {
			return errors.New("text_leak")
		}
		if r.Groups != 40 || r.Families != 120 || r.DirectReport[1] != 40 {
			return errors.New("aggregate_count")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := add("valid80_join_order_stable_typed_candidate_components", "PASS", func() error {
		a, e := synLoad(base)
		if e != nil {
			return e
		}
		a.input.Path = "synthetic-tranche01.json"
		m := synthetic(41)
		m.Relations = []Relation{{[]string{groupID(1), groupID(41)}, []int{1, 41}, "possible_shared_ancestor;not_established", "synthetic rationale", "synthetic distinction", "pending_independent_full_registry_lineage_review", []string{"scaffold_overlap"}}}
		m.Universe = &Universe{[]int{41, 80}, []int{1, 80}, []ExternalMetadata{{[]int{1, 40}, a.input.Path, a.input.SHA}}, "PROVISIONAL_NOT_CONFIRMED_LINEAGE"}
		b, e := synLoad(m)
		if e != nil {
			return e
		}
		b.input.Path = "synthetic-tranche02.json"
		r, e := audit([]loaded{a, b})
		if e != nil {
			return e
		}
		s, e := audit([]loaded{b, a})
		if e != nil {
			return e
		}
		if !bytes.Equal(encode(r), encode(s)) || r.Components.Count != 79 || r.Components.Max != 2 || r.RelationTypes[m.Relations[0].Type][m.Relations[0].Status] != 1 {
			return errors.New("relation_preservation")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	for _, c := range []struct {
		name, want string
		mutate     func(*Metadata)
	}{
		{"duplicate_group_id", "duplicate_group_id", func(m *Metadata) { m.Groups[1].ID = m.Groups[0].ID }},
		{"duplicate_family_id", "duplicate_family_id", func(m *Metadata) { m.Groups[1].Families[0].ID = m.Groups[0].Families[0].ID }},
		{"wrong_seed", "identity_seed", func(m *Metadata) { m.Identity.Seed = "other" }},
		{"wrong_namespace", "group_identity", func(m *Metadata) { m.Groups[0].ID = "f_" + m.Groups[0].ID[2:] }},
		{"empty_ko_concept", "concept_description", func(m *Metadata) { m.Groups[0].Families[0].KO = " \t" }},
		{"missing_frame", "group_cardinality", func(m *Metadata) { m.Groups[0].Families = m.Groups[0].Families[:2] }},
		{"non40_source_cardinality", "tranche_scope", func(m *Metadata) { m.Groups = m.Groups[:39] }},
		{"unresolved_within_edge", "unresolved_within_link", func(m *Metadata) { m.Groups[0].Links[0].To = familyID(400, 1) }},
		{"metadata_boundary_utterance_true", "metadata_boundary", func(m *Metadata) { m.Boundary.Utterances = true }},
		{"correction_without_prior", "correction_provenance", func(m *Metadata) { m.Correction = &Correction{Batch: 1, Rows: 0, Deltas: []Delta{}} }},
	} {
		if err := test(c.name, c.want, c.mutate); err != nil {
			return nil, err
		}
	}
	if err := add("wrong_pin_before_malformed_json_decode", "pin_mismatch", func() error { _, e := decodePinned([]byte("{malformed"), strings.Repeat("0", 64)); return e }); err != nil {
		return nil, err
	}
	if err := add("correct_pin_exposes_malformed_json", "invalid_json", func() error { b := []byte("{malformed"); _, e := decodePinned(b, hash(b)); return e }); err != nil {
		return nil, err
	}
	if err := add("invalid_utf8_metadata_reject", "metadata_utf8", func() error {
		b := bytes.Replace(encode(base), []byte("synthetic concept 001"), []byte{'s', 0xff}, 1)
		_, e := decodePinned(b, hash(b))
		return e
	}); err != nil {
		return nil, err
	}
	if err := add("duplicate_core_names_flag_constraints_only", "PASS", func() error {
		m := synthetic(1)
		m.Groups[1].Core = m.Groups[0].Core
		l, e := synLoad(m)
		if e != nil {
			return e
		}
		r, e := audit([]loaded{l})
		if e != nil {
			return e
		}
		if r.Duplicates.PairClusters != 1 || r.Duplicates.PairGroups != 2 || r.Components.Count != 40 {
			return errors.New("duplicate_constraint")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := add("metadata_as_utterance_schema_rejected", "schema_unknown_field", func() error {
		b := encode(base)
		b = append([]byte(`{"utterances":[],`), b[2:]...)
		_, e := decodePinned(b, hash(b))
		return e
	}); err != nil {
		return nil, err
	}
	if err := add("nested_message_schema_rejected", "schema_unknown_field", func() error {
		b := encode(base)
		b = bytes.Replace(b, []byte(`"concept_ko":`), []byte(`"message": "synthetic", "concept_ko":`), 1)
		_, e := decodePinned(b, hash(b))
		return e
	}); err != nil {
		return nil, err
	}
	if err := add("duplicate_json_key", "duplicate_json_key", func() error {
		b := encode(base)
		b = append([]byte(`{"schema":"duplicate",`), b[2:]...)
		_, e := decodePinned(b, hash(b))
		return e
	}); err != nil {
		return nil, err
	}
	if err := add("unresolved_cross_relation_actual_universe", "unresolved_cross_group_relation", func() error {
		m := synthetic(1)
		m.Relations = []Relation{{[]string{groupID(1), groupID(41)}, []int{1, 41}, "possible_shared_ancestor", "synthetic rationale", "synthetic distinction", "pending_independent_full_registry_lineage_review", nil}}
		l, e := synLoad(m)
		if e != nil {
			return e
		}
		_, e = audit([]loaded{l})
		return e
	}); err != nil {
		return nil, err
	}
	if err := add("external_pin_not_supplied", "unresolved_external_metadata_declaration", func() error {
		first, e := synLoad(synthetic(1))
		if e != nil {
			return e
		}
		first.input.Path = "synthetic-tranche01.json"
		m := synthetic(41)
		m.Universe = &Universe{[]int{41, 80}, []int{1, 80}, []ExternalMetadata{{[]int{1, 40}, first.input.Path, strings.Repeat("0", 64)}}, "PROVISIONAL"}
		l, e := synLoad(m)
		if e != nil {
			return e
		}
		l.input.Path = "synthetic-tranche02.json"
		_, e = audit([]loaded{first, l})
		return e
	}); err != nil {
		return nil, err
	}
	if err := add("exact_200KiB_accept", "PASS", func() error {
		b := encode(base)
		b = append(b, bytes.Repeat([]byte(" "), maxInput-len(b))...)
		_, e := decodePinned(b, hash(b))
		return e
	}); err != nil {
		return nil, err
	}
	if err := add("over_200KiB_reject_before_decode", "input_size_limit", func() error { _, e := readBounded(bytes.NewReader(bytes.Repeat([]byte(" "), maxInput+1))); return e }); err != nil {
		return nil, err
	}
	if err := add("relative_contained_path_rules", "PASS", func() error {
		if relative(string(filepath.Separator)+"synthetic.json") || relative("../x.json") || relative("a/../x.json") || relative("x.txt") || !relative("a/x.json") {
			return errors.New("path_rules")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := add("retained_correction_exact_prior_pin", "PASS", func() error {
		m := synthetic(1)
		prior := encode(m)
		n, _ := tree(prior)
		prior = append(compact(n, nil, nil), '\n')
		old, _ := json.Marshal(m.Groups[0].Families[0].KO)
		m.Groups[0].Families[0].KO = "SYNTHETIC_CORRECTED_KO_CONCEPT_SENTINEL"
		neu, _ := json.Marshal(m.Groups[0].Families[0].KO)
		m.Correction = &Correction{1, "2026-10-07T00:00:00Z", "synthetic trigger", "synthetic pre-freeze correction", hash(prior), len(prior), correctionSerialization, []Delta{{"/groups/0/families/0/concept_ko", old, neu}}, 0, nil, ""}
		_, e := synLoad(m)
		if e != nil {
			return e
		}
		m.Correction.PriorSHA = strings.Repeat("0", 64)
		_, e = synLoad(m)
		if e == nil || e.Error() != "correction_prior_pin" {
			return errors.New("correction_pin_not_checked")
		}
		m.Correction.PriorSHA = hash(prior)
		m.Correction.Deltas[0].New = json.RawMessage(`"unrelated synthetic value"`)
		_, e = synLoad(m)
		if e == nil || e.Error() != "correction_value" {
			return errors.New("correction_current_value_not_checked")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	for _, alias := range []struct{ name, path string }{
		{"group_leading_zero", "/groups/00/families/0/concept_ko"},
		{"group_plus_zero", "/groups/+0/families/0/concept_ko"},
		{"family_leading_zero", "/groups/0/families/00/concept_ko"},
		{"family_plus_zero", "/groups/0/families/+0/concept_ko"},
	} {
		if err := add("correction_duplicate_location_alias_"+alias.name, "correction_pointer", func() error {
			m := synthetic(1)
			priorTree, _ := tree(encode(m))
			prior := append(compact(priorTree, nil, nil), '\n')
			old, _ := json.Marshal(m.Groups[0].Families[0].KO)
			m.Groups[0].Families[0].KO = "SYNTHETIC_CORRECTED_KO_CONCEPT_SENTINEL"
			neu, _ := json.Marshal(m.Groups[0].Families[0].KO)
			m.Correction = &Correction{1, "2026-10-07T00:00:00Z", "synthetic trigger", "synthetic pre-freeze correction", hash(prior), len(prior), correctionSerialization, []Delta{{"/groups/0/families/0/concept_ko", old, neu}, {alias.path, old, neu}}, 0, nil, ""}
			_, e := synLoad(m)
			return e
		}); err != nil {
			return nil, err
		}
	}
	protocolFixture := func() Metadata {
		m := synthetic(1)
		m.Design.Recruitment = "synthetic recruitment protocol"
		m.Design.NoMandatoryReport = "synthetic no mandatory report protocol"
		m.Design.EssentialScope = "synthetic scope protocol"
		m.Scope.Known = &KnownProvisional{40, 120, 240}
		m.Groups[0].Ledger = &EssentialScopeLedger{[]string{"synthetic optional detail"}, "synthetic core scope"}
		m.Groups[0].Scaffold = &ScaffoldProvenance{"synthetic common provenance", "synthetic construction"}
		m.Sources.Pins = []DeclaredSourcePin{{"synthetic-source.json", strings.Repeat("0", 64), 1}}
		m.Generation = &GenerationProvenance{"synthetic role", "synthetic origin", "synthetic composition", "unassessed", "synthetic inspection limit", "synthetic scaffold policy"}
		return m
	}
	if err := add("explicit_protocol_provenance_schema_accept", "PASS", func() error { _, e := synLoad(protocolFixture()); return e }); err != nil {
		return nil, err
	}
	if err := add("known_provisional_scope_still_bounded", "known_provisional_scope", func() error { m := protocolFixture(); m.Scope.Known.Groups = 80; _, e := synLoad(m); return e }); err != nil {
		return nil, err
	}
	if err := add("declared_source_pin_format_checked_only", "declared_source_pin", func() error { m := protocolFixture(); m.Sources.Pins[0].SHA = "invalid"; _, e := synLoad(m); return e }); err != nil {
		return nil, err
	}
	if err := add("protocol_provenance_rejects_assigned_splits", "schema_unknown_field", func() error {
		b := encode(protocolFixture())
		b = bytes.Replace(b, []byte(`"role":`), []byte(`"assigned_splits":[],"role":`), 1)
		_, e := decodePinned(b, hash(b))
		return e
	}); err != nil {
		return nil, err
	}
	if err := add("prior_correction_object_core_and_slot_exact_reconstruction", "PASS", func() error {
		m := synthetic(1)
		initial := &InitialCorrection{0, "synthetic initial correction status", 0}
		before, _ := tree(encode(m))
		prior := compact(before, nil, nil)
		initialBytes, _ := json.Marshal(initial)
		prior = append(prior[:len(prior)-1], []byte(`,"retained_provisional_metadata_correction":`)...)
		prior = append(prior, initialBytes...)
		prior = append(prior, '}', '\n')
		oldCore, _ := json.Marshal(m.Groups[0].Core.KO)
		oldSlot, _ := json.Marshal(m.Groups[0].Slots[0].KO)
		oldFamily, _ := json.Marshal(m.Groups[0].Families[0].KO)
		m.Groups[0].Core.KO = "synthetic corrected core"
		m.Groups[0].Slots[0].KO = "synthetic corrected necessary slot"
		m.Groups[0].Families[0].KO = "SYNTHETIC_CORRECTED_KO_CONCEPT_SENTINEL"
		newCore, _ := json.Marshal(m.Groups[0].Core.KO)
		newSlot, _ := json.Marshal(m.Groups[0].Slots[0].KO)
		newFamily, _ := json.Marshal(m.Groups[0].Families[0].KO)
		m.Correction = &Correction{1, "2026-10-07T00:00:00Z", "synthetic trigger", "synthetic pre-freeze correction", hash(prior), len(prior), correctionWithPriorSerialization, []Delta{{"/groups/0/core_event/ko", oldCore, newCore}, {"/groups/0/necessary_slots/0/ko", oldSlot, newSlot}, {"/groups/0/families/0/concept_ko", oldFamily, newFamily}}, 0, initial, "synthetic metadata reason"}
		if _, e := synLoad(m); e != nil {
			return e
		}
		m.Correction.Prior.Rows = 1
		if _, e := synLoad(m); e == nil || e.Error() != "correction_provenance" {
			return errors.New("prior_initial_rows_not_checked")
		}
		m.Correction.Prior.Rows = 0
		m.Correction.Deltas = append(m.Correction.Deltas, Delta{"/groups/0/necessary_slots/+0/ko", oldSlot, newSlot})
		if _, e := synLoad(m); e == nil || e.Error() != "correction_pointer" {
			return errors.New("new_slot_alias_not_rejected")
		}
		m.Correction.Deltas = m.Correction.Deltas[:3]
		m.Correction.PriorSHA = strings.Repeat("0", 64)
		if _, e := synLoad(m); e == nil || e.Error() != "correction_prior_pin" {
			return errors.New("prior_object_pin_not_checked")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := add("case_preserved_relation_abbreviations", "PASS", func() error {
		m := synthetic(1)
		for _, typ := range []string{"possible_shared_CLI_ancestor;not_asserted_paraphrase", "possible_shared_UI_ancestor;not_asserted_paraphrase"} {
			m.Relations = append(m.Relations, Relation{[]string{groupID(1), groupID(2)}, []int{1, 2}, typ, "synthetic rationale", "synthetic distinction", "pending_independent_full_registry_lineage_review", []string{"event"}})
		}
		l, e := synLoad(m)
		if e != nil {
			return e
		}
		r, e := audit([]loaded{l})
		if e != nil {
			return e
		}
		for _, relation := range m.Relations {
			if r.RelationTypes[relation.Type][relation.Status] != 1 {
				return errors.New("relation_spelling_not_preserved")
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := add("relation_type_prose_still_rejected", "relation_schema", func() error {
		m := synthetic(1)
		m.Relations = []Relation{{[]string{groupID(1), groupID(2)}, []int{1, 2}, "possible shared CLI ancestor", "synthetic rationale", "synthetic distinction", "pending_independent_full_registry_lineage_review", []string{"event"}}}
		_, e := synLoad(m)
		return e
	}); err != nil {
		return nil, err
	}
	if err := test("lowercase_phenomenon_tags_remain_required", "concept_description", func(m *Metadata) { m.Groups[0].Families[0].Phenomena = []string{"Synthetic_Tag"} }); err != nil {
		return nil, err
	}
	if err := test("lowercase_dimension_tags_remain_required", "relation_schema", func(m *Metadata) {
		m.Relations = []Relation{{[]string{groupID(1), groupID(2)}, []int{1, 2}, "possible_CLI_ancestor", "synthetic rationale", "synthetic distinction", "pending_independent_full_registry_lineage_review", []string{"Event"}}}
	}); err != nil {
		return nil, err
	}
	for _, field := range []string{"contains_sample_utterances", "contains_target_assignments", "contains_expected_semantic_counts", "contains_fixed_semantic_triangle", "contains_assigned_length_bins", "contains_assigned_splits", "contains_training_or_gold_or_reference_material"} {
		if err := add("false_boundary_required_"+field, "metadata_boundary", func() error {
			b := encode(synthetic(1))
			b = bytes.Replace(b, []byte(`"`+field+`": false`), []byte(`"`+field+`": true`), 1)
			_, err := decodePinned(b, hash(b))
			return err
		}); err != nil {
			return nil, err
		}
	}
	return map[string]any{"schema": "concept-registry-audit-synthetic-checks-text-free-v2", "status": "PASS_SYNTHETIC_ONLY", "go_required": "go1.27.1", "go_runtime": runtime.Version(), "process_limits": map[string]any{"GOMAXPROCS": runtime.GOMAXPROCS(0)}, "max_input_bytes_each": maxInput, "max_total_input_bytes": maxTotal, "maximum_inputs": maxTranches, "check_count": len(checks), "checks": checks, "fixtures": "Owned in-memory synthetic concept metadata only; no persistent fixtures or binary required.", "scope": "Only owned synthetic metadata is used by self-test.", "limitations": []string{"Synthetic checks demonstrate structural behavior only.", "No semantic labels, short feasibility, target validity, independence, lineage confirmation, or effective sample size is inferred."}}, nil
}
func main() {
	var inputs specs
	var fake bool
	flags := flag.NewFlagSet("riido-statehint-concept-audit", flag.ContinueOnError)
	// Parser diagnostics can contain arbitrary argument values; keep errors text-free.
	flags.SetOutput(io.Discard)
	flags.Var(&inputs, "input", "clean relative metadata.json=lowercase SHA256 (repeat 1..10; 40-group tranches)")
	flags.BoolVar(&fake, "self-test", false, "run owned in-memory synthetic checks only")
	parseErr := flags.Parse(os.Args[1:])
	if errors.Is(parseErr, flag.ErrHelp) {
		fmt.Fprintln(os.Stdout, "Usage: riido-statehint-concept-audit -input metadata.json=SHA256 [-input metadata-next.json=SHA256 ...]\n       riido-statehint-concept-audit -self-test")
		return
	}
	if parseErr != nil || flags.NArg() != 0 || fake && len(inputs) > 0 {
		fmt.Fprintln(os.Stderr, "audit_rejected:argument_scope")
		os.Exit(2)
	}
	var result any
	var err error
	if fake {
		result, err = selfTest()
	} else {
		var all []loaded
		all, err = loadInputs(inputs, prepare, load)
		if err == nil {
			result, err = audit(all)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "audit_rejected:"+err.Error())
		os.Exit(2)
	}
	if _, err = os.Stdout.Write(encode(result)); err != nil {
		os.Exit(2)
	}
}
