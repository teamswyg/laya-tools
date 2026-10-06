// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintcorpus checks paired bilingual data without loading a model.
package statehintcorpus

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const MaxFileBytes = 16 << 20
const maxLineBytes = 64 << 10
const maxRows = 2400

var ErrInput = errors.New("corpus structure invalid or over budget")

// Options declares a single partition and the already frozen rubric digest.
// This digest is a caller claim until a driver verifies the rubric file itself.
type Options struct {
	Partition string
	RubricSHA string
}

// Row retains source fields for later author review. Only Text is model input.
// No field or parser result establishes annotation rights or training consent.
type Row struct {
	Schema              string           `json:"schema"`
	ID                  string           `json:"id"`
	Family              string           `json:"family_id"`
	Lineage             string           `json:"leakage_group_id"`
	Partition           string           `json:"partition"`
	Locale              string           `json:"locale"`
	Wording             int              `json:"wording_index"`
	Role                string           `json:"role"`
	Applicable          bool             `json:"semantic_applicable"`
	Text                string           `json:"text"`
	Unit                string           `json:"declared_unit"`
	ClauseScopes        []string         `json:"clause_scope_design"`
	AssertionForms      []string         `json:"assertion_form_design"`
	Expected            statehint.Intent `json:"expected_intent"`
	Ambiguous           bool             `json:"ambiguous"`
	TaskLevelCompletion bool             `json:"task_level_completion"`
	AnnotationSource    string           `json:"annotation_source"`
	OntologyVersion     string           `json:"ontology_version"`
	OntologyFreezeSHA   string           `json:"ontology_freeze_sha256"`
	License             string           `json:"license"`
}

// Family contains metadata only; it permits split auditing without opening
// final-test text or labels embedded in a text file. Labels here are declared
// author metadata, not model outcomes or an independent semantic verification.
type Family struct {
	ID        string           `json:"family_id"`
	Lineage   string           `json:"leakage_group_id"`
	Partition string           `json:"partition"`
	Expected  statehint.Intent `json:"expected_intent"`
}

type Pair struct {
	Family Family
	Rows   [2]int // ko, en; indices into source-order Rows
}

// Corpus intentionally has no public JSON representation: it holds input text.
type Corpus struct {
	rows  []Row
	pairs []Pair
}

// Rows returns a detached snapshot. Callers cannot mutate the validated corpus
// by changing returned rows or their annotation slices.
func (corpus Corpus) Rows() []Row {
	rows := append([]Row(nil), corpus.rows...)
	for i := range rows {
		rows[i].ClauseScopes = append([]string(nil), rows[i].ClauseScopes...)
		rows[i].AssertionForms = append([]string(nil), rows[i].AssertionForms...)
	}
	return rows
}

// Pairs returns detached metadata and row indices in ko,en order.
func (corpus Corpus) Pairs() []Pair {
	return append([]Pair(nil), corpus.pairs...)
}

type Summary struct {
	Rows              int                        `json:"rows"`
	Families          int                        `json:"families"`
	Lineages          int                        `json:"declared_lineages"`
	LocaleRows        [2]int                     `json:"locale_rows_ko_en"`
	IntentFamilies    [statehint.IntentCount]int `json:"intent_families_persisted_order"`
	PartitionFamilies [4]int                     `json:"partition_families_train_validation_calibration_test"`
}

// Read checks a single JSONL partition. It never opens another partition,
// modifies input, evaluates predictions, or certifies semantic labels/rights.
func Read(in io.Reader, options Options) (Corpus, Summary, error) {
	var corpus Corpus
	if in == nil || partitionIndex(options.Partition) < 0 || !digest(options.RubricSHA) {
		return corpus, Summary{}, ErrInput
	}
	limited := &io.LimitedReader{R: in, N: MaxFileBytes + 1}
	scanner := bufio.NewScanner(limited)
	scanner.Buffer(make([]byte, 4096), maxLineBytes)
	for scanner.Scan() {
		if len(corpus.rows) == maxRows {
			return Corpus{}, Summary{}, ErrInput
		}
		row, err := decodeRow(scanner.Bytes())
		if err != nil || !validRow(row, options) {
			return Corpus{}, Summary{}, ErrInput
		}
		corpus.rows = append(corpus.rows, row)
	}
	if scanner.Err() != nil || limited.N <= 0 || len(corpus.rows) == 0 || len(corpus.rows)%2 != 0 {
		return Corpus{}, Summary{}, ErrInput
	}
	indices := make([]int, len(corpus.rows))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool { return corpus.rows[indices[i]].ID < corpus.rows[indices[j]].ID })
	for i := 1; i < len(indices); i++ {
		if corpus.rows[indices[i-1]].ID == corpus.rows[indices[i]].ID {
			return Corpus{}, Summary{}, ErrInput
		}
	}
	hashes := make([][32]byte, len(corpus.rows))
	for i, row := range corpus.rows {
		hashes[i] = sha256.Sum256([]byte(row.Text))
	}
	sort.Slice(hashes, func(i, j int) bool { return bytes.Compare(hashes[i][:], hashes[j][:]) < 0 })
	for i := 1; i < len(hashes); i++ {
		if hashes[i-1] == hashes[i] {
			return Corpus{}, Summary{}, ErrInput
		}
	}
	sort.Slice(indices, func(i, j int) bool {
		a, b := corpus.rows[indices[i]], corpus.rows[indices[j]]
		if a.Family != b.Family {
			return a.Family < b.Family
		}
		return a.Locale < b.Locale
	})
	metadata := make([]Family, 0, len(indices)/2)
	for i := 0; i < len(indices); i += 2 {
		a, b := corpus.rows[indices[i]], corpus.rows[indices[i+1]]
		// Alphabetic locale sorting is en,ko. Each family has exactly this pair.
		if a.Family != b.Family || a.Locale != "en" || b.Locale != "ko" || a.Lineage != b.Lineage || a.Expected != b.Expected || a.Unit != b.Unit || a.Ambiguous != b.Ambiguous || a.OntologyVersion != b.OntologyVersion || (i+2 < len(indices) && corpus.rows[indices[i+2]].Family == a.Family) {
			return Corpus{}, Summary{}, ErrInput
		}
		family := Family{ID: a.Family, Lineage: a.Lineage, Partition: a.Partition, Expected: a.Expected}
		metadata = append(metadata, family)
		corpus.pairs = append(corpus.pairs, Pair{Family: family, Rows: [2]int{indices[i+1], indices[i]}})
	}
	summary, err := ValidateMetadata(metadata)
	if err != nil {
		return Corpus{}, Summary{}, err
	}
	return corpus, summary, nil
}

// ValidateMetadata rejects repeated family IDs and cross-partition declared
// lineages. It does not invent independence from distinct identifiers. It is
// safe to call on the complete family manifest before reading held-out text.
func ValidateMetadata(families []Family) (Summary, error) {
	var summary Summary
	if len(families) == 0 || len(families) > maxRows/2 {
		return summary, ErrInput
	}
	order := make([]int, len(families))
	for i, family := range families {
		class, ok := statehint.IntentIndex(family.Expected)
		partition := partitionIndex(family.Partition)
		if !opaque(family.ID) || !opaque(family.Lineage) || !ok || partition < 0 {
			return Summary{}, ErrInput
		}
		order[i] = i
		summary.IntentFamilies[class]++
		summary.PartitionFamilies[partition]++
	}
	sort.Slice(order, func(i, j int) bool { return families[order[i]].ID < families[order[j]].ID })
	for i := 1; i < len(order); i++ {
		if families[order[i-1]].ID == families[order[i]].ID {
			return Summary{}, ErrInput
		}
	}
	sort.Slice(order, func(i, j int) bool { return families[order[i]].Lineage < families[order[j]].Lineage })
	for i, index := range order {
		if i == 0 || families[order[i-1]].Lineage != families[index].Lineage {
			summary.Lineages++
		} else if families[order[i-1]].Partition != families[index].Partition {
			return Summary{}, ErrInput
		}
	}
	summary.Families = len(families)
	summary.Rows = 2 * len(families)
	summary.LocaleRows = [2]int{len(families), len(families)}
	return summary, nil
}

// MatchMetadata requires every parsed pair to match the manifest partition
// exactly. It neither reads nor returns text for the other partitions.
func MatchMetadata(corpus Corpus, families []Family, partition string) error {
	if partitionIndex(partition) < 0 || len(corpus.pairs) == 0 {
		return ErrInput
	}
	if _, err := ValidateMetadata(families); err != nil {
		return err
	}
	declared := make([]Family, 0, len(corpus.pairs))
	for _, family := range families {
		if family.Partition == partition {
			declared = append(declared, family)
		}
	}
	if len(declared) != len(corpus.pairs) {
		return ErrInput
	}
	sort.Slice(declared, func(i, j int) bool { return declared[i].ID < declared[j].ID })
	for i, pair := range corpus.pairs {
		if pair.Family != declared[i] {
			return ErrInput
		}
	}
	return nil
}

func validRow(row Row, options Options) bool {
	_, ok := statehint.IntentIndex(row.Expected)
	if !ok || row.Schema != "statehint-v4-original-train-seed-row-v1" || !opaque(row.ID) || !opaque(row.Family) || !opaque(row.Lineage) || row.Partition != options.Partition || row.Locale != "ko" && row.Locale != "en" || row.Wording != 1 || row.Role != "prose" || !row.Applicable || row.TaskLevelCompletion || row.OntologyFreezeSHA != options.RubricSHA || row.License != "Apache-2.0" || row.OntologyVersion != "statehint-intent-scope-v4-1200x2-v1" {
		return false
	}
	if !boundedText(row.Text, statehint.MaxTextBytes) || !boundedText(row.Unit, 1024) || !boundedText(row.AnnotationSource, 1024) || row.Ambiguous && row.Expected != statehint.Unclear {
		return false
	}
	return tags(row.ClauseScopes, []string{"current_unit", "substep", "other_work", "reference", "unknown"}) && (len(row.AssertionForms) == 0 && row.Expected == statehint.Unclear || tags(row.AssertionForms, []string{"asserted", "ongoing", "remaining", "planned", "conditional", "negated", "quoted", "questioned", "requested"}))
}

func boundedText(value string, limit int) bool {
	return strings.TrimSpace(value) != "" && len(value) <= limit && utf8.ValidString(value) && !strings.ContainsRune(value, 0)
}

func tags(values, allowed []string) bool {
	if len(values) == 0 || len(values) > len(allowed) {
		return false
	}
	for i, value := range values {
		found := false
		for _, candidate := range allowed {
			found = found || value == candidate
		}
		if !found {
			return false
		}
		for _, previous := range values[:i] {
			if previous == value {
				return false
			}
		}
	}
	return true
}

func opaque(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func digest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

func partitionIndex(value string) int {
	for i, candidate := range [4]string{"train", "validation", "calibration", "test"} {
		if value == candidate {
			return i
		}
	}
	return -1
}

// A fixed key set rejects duplicate keys, case aliases, nulls and additions.
// encoding/json's decoded logical strings are used; escaped Unicode is not
// required to preserve its original byte spelling.
func decodeRow(data []byte) (Row, error) {
	var row Row
	if !utf8.Valid(data) {
		return row, ErrInput
	}
	keys := [20]string{"schema", "id", "family_id", "leakage_group_id", "partition", "locale", "wording_index", "role", "semantic_applicable", "text", "declared_unit", "clause_scope_design", "assertion_form_design", "expected_intent", "ambiguous", "task_level_completion", "annotation_source", "ontology_version", "ontology_freeze_sha256", "license"}
	var seen [20]bool
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return row, ErrInput
	}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return Row{}, ErrInput
		}
		index := -1
		for i, candidate := range keys {
			if key == candidate {
				index = i
				break
			}
		}
		if index < 0 || seen[index] {
			return Row{}, ErrInput
		}
		var raw json.RawMessage
		if decoder.Decode(&raw) != nil || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return Row{}, ErrInput
		}
		var target any
		switch index {
		case 0:
			target = &row.Schema
		case 1:
			target = &row.ID
		case 2:
			target = &row.Family
		case 3:
			target = &row.Lineage
		case 4:
			target = &row.Partition
		case 5:
			target = &row.Locale
		case 6:
			target = &row.Wording
		case 7:
			target = &row.Role
		case 8:
			target = &row.Applicable
		case 9:
			target = &row.Text
		case 10:
			target = &row.Unit
		case 11:
			target = &row.ClauseScopes
		case 12:
			target = &row.AssertionForms
		case 13:
			target = &row.Expected
		case 14:
			target = &row.Ambiguous
		case 15:
			target = &row.TaskLevelCompletion
		case 16:
			target = &row.AnnotationSource
		case 17:
			target = &row.OntologyVersion
		case 18:
			target = &row.OntologyFreezeSHA
		case 19:
			target = &row.License
		}
		if json.Unmarshal(raw, target) != nil {
			return Row{}, ErrInput
		}
		seen[index] = true
	}
	for _, required := range seen {
		if !required {
			return Row{}, ErrInput
		}
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return Row{}, ErrInput
	}
	if _, err = decoder.Token(); err != io.EOF {
		return Row{}, ErrInput
	}
	return row, nil
}
