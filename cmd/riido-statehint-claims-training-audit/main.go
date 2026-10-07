// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only, text-free audit of one immutable public training corpus.
package main

import (
	"bufio"
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
	"sort"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"
)

const (
	trainingSHA = "c5b5a0adbc721d73213d4565c6a7c430094f31b8e7502fb5249e7ed40a3da837"
	rubricSHA   = "7d908277fe4624374f7d8cc9318f1a228e2128b595ee2bc94fe214e5f258d0b5"
	inputBudget = 8 << 20
)

var errAudit = errors.New("training audit rejected pinned public input")

// Fixture injection is available to package tests only; external CLI pins are fixed.
type manifest struct {
	sha                    string
	rows, families, groups int
}

func frozenManifest() manifest { return manifest{trainingSHA, 1680, 840, 280} }

type evidence struct {
	Start  *int   `json:"start_byte"`
	End    *int   `json:"end_byte"`
	Reason string `json:"reason"`
}
type inputRow struct {
	Schema     string     `json:"schema"`
	ID         string     `json:"id"`
	Family     string     `json:"family_id"`
	Group      string     `json:"leakage_group_id"`
	Split      string     `json:"internal_split"`
	Locale     string     `json:"locale"`
	Text       string     `json:"text"`
	Targets    []string   `json:"targets"`
	Rubric     string     `json:"rubric_sha256"`
	Annotation string     `json:"annotation_source"`
	Evidence   []evidence `json:"evidence"`
}
type row struct {
	inputRow
	labels [3]int
	locale int
}
type fraction struct {
	Count int     `json:"count"`
	Mean  float64 `json:"mean"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
}
type duplicates struct {
	Clusters            int `json:"clusters"`
	RedundantRows       int `json:"redundant_rows"`
	ConflictingClusters int `json:"conflicting_target_clusters"`
	ConflictingPairs    int `json:"conflicting_target_pairs"`
}
type report struct {
	Schema                 string          `json:"schema"`
	TrainingSHA            string          `json:"public_training_sha256"`
	LocaleOrder            [2]string       `json:"locale_order"`
	HeadOrder              [3]string       `json:"head_order"`
	StateOrder             [3]string       `json:"state_order"`
	Rows                   int             `json:"rows"`
	Families               int             `json:"families"`
	Groups                 int             `json:"groups"`
	LocaleRows             [2]int          `json:"locale_rows"`
	GroupsWithLocale       [2]int          `json:"groups_with_locale"`
	HeadStates             [2][3][3]int    `json:"locale_head_state_rows"`
	GroupSupports          [2][3][3]int    `json:"locale_head_state_group_supports"`
	LengthBounds           [4]int          `json:"text_utf8_byte_bucket_upper_bounds"`
	LengthRows             [4]int          `json:"text_utf8_byte_bucket_rows"`
	LocaleLengthRows       [2][4]int       `json:"locale_text_utf8_byte_bucket_rows"`
	LengthStates           [4][2][3][3]int `json:"bucket_locale_head_state_rows"`
	RuneBounds             [4]int          `json:"text_unicode_rune_bucket_upper_bounds"`
	RuneRows               [4]int          `json:"text_unicode_rune_bucket_rows"`
	LocaleRuneRows         [2][4]int       `json:"locale_text_unicode_rune_bucket_rows"`
	RuneStates             [4][2][3][3]int `json:"rune_bucket_locale_head_state_rows"`
	EmptyEvidence          [3][3]int       `json:"head_state_empty_evidence_rows"`
	NonemptyEvidence       [3][3]int       `json:"head_state_nonempty_evidence_rows"`
	SpanFractions          [3][3]fraction  `json:"head_state_span_byte_fraction"`
	PairDifferences        [3]int          `json:"paired_locale_target_difference_families"`
	DuplicateNormalization string          `json:"duplicate_normalization"`
	Duplicates             duplicates      `json:"normalized_text_duplicates"`
	Limitations            string          `json:"limitations"`
}

func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func local(s string) bool {
	return len(s) <= 4096 && filepath.IsLocal(s) && s != "." && filepath.Clean(s) == s && !strings.Contains(s, "\\")
}
func identifier(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
func boundary(s string, i int) bool {
	return i >= 0 && i <= len(s) && (i == len(s) || utf8.RuneStart(s[i]))
}
func index(s string, order []string) int {
	for i, v := range order {
		if s == v {
			return i
		}
	}
	return -1
}

// A root-bound component walk rejects normalized path aliases, symlinks and hardlinks.
func readPinned(root *os.Root, name, supplied string, pin manifest) ([]byte, error) {
	if supplied != pin.sha || !local(name) {
		return nil, errAudit
	}
	parts := strings.Split(name, string(filepath.Separator))
	current := root
	var owned *os.Root
	defer func() {
		if owned != nil {
			owned.Close()
		}
	}()
	for _, part := range parts[:len(parts)-1] {
		before, e := current.Lstat(part)
		if e != nil || !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
			return nil, errAudit
		}
		next, e := current.OpenRoot(part)
		if e != nil {
			return nil, errAudit
		}
		after, e := next.Stat(".")
		if e != nil || !os.SameFile(before, after) {
			next.Close()
			return nil, errAudit
		}
		if owned != nil {
			owned.Close()
		}
		owned, current = next, next
	}
	leaf := parts[len(parts)-1]
	before, e := current.Lstat(leaf)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > inputBudget {
		return nil, errAudit
	}
	st, ok := before.Sys().(*syscall.Stat_t)
	if !ok || st.Nlink != 1 {
		return nil, errAudit
	}
	f, e := current.Open(leaf)
	if e != nil {
		return nil, errAudit
	}
	opened, e := f.Stat()
	if e != nil || !sameFile(before, opened) {
		f.Close()
		return nil, errAudit
	}
	b, re := io.ReadAll(io.LimitReader(f, inputBudget+1))
	after, se := f.Stat()
	ce := f.Close()
	if re != nil || se != nil || ce != nil || !sameFile(before, after) || len(b) != int(before.Size()) || len(b) > inputBudget || digest(b) != pin.sha {
		return nil, errAudit
	}
	return b, nil
}
func sameFile(a, b os.FileInfo) bool {
	if a == nil || b == nil || !os.SameFile(a, b) || a.Size() != b.Size() || a.Mode() != b.Mode() || !a.ModTime().Equal(b.ModTime()) {
		return false
	}
	as, aok := a.Sys().(*syscall.Stat_t)
	bs, bok := b.Sys().(*syscall.Stat_t)
	return aok && bok && as.Nlink == 1 && bs.Nlink == 1
}

// encoding/json permits duplicate keys and missing fields; reject both recursively.
func decodeStrict(b []byte, v any) error {
	if !utf8.Valid(b) {
		return errAudit
	}
	d := json.NewDecoder(bytes.NewReader(b))
	depth := 0
	var value func() error
	value = func() error {
		t, e := d.Token()
		if e != nil {
			return e
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		depth++
		defer func() { depth-- }()
		if depth > 16 {
			return errAudit
		}
		switch delim {
		case '{':
			seen := make(map[string]bool)
			for d.More() {
				key, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := key.(string)
				if !ok || seen[s] {
					return errAudit
				}
				seen[s] = true
				if e := value(); e != nil {
					return e
				}
			}
		case '[':
			for d.More() {
				if e := value(); e != nil {
					return e
				}
			}
		default:
			return errAudit
		}
		_, e = d.Token()
		return e
	}
	if value() != nil {
		return errAudit
	}
	if _, e := d.Token(); e != io.EOF {
		return errAudit
	}
	var shape any
	if json.Unmarshal(b, &shape) != nil || !completeShape(shape, reflect.TypeOf(v).Elem()) {
		return errAudit
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return errAudit
	}
	return nil
}
func completeShape(v any, t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		return v != nil && completeShape(v, t.Elem())
	}
	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok || len(m) != t.NumField() {
			return false
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			x, ok := m[f.Tag.Get("json")]
			if !ok || !completeShape(x, f.Type) {
				return false
			}
		}
	case reflect.Slice:
		a, ok := v.([]any)
		if !ok {
			return false
		}
		for _, x := range a {
			if !completeShape(x, t.Elem()) {
				return false
			}
		}
	default:
		if v == nil {
			return false
		}
	}
	return true
}

func prepare(b []byte, pin manifest) ([]row, error) {
	if digest(b) != pin.sha || len(b) == 0 || len(b) > inputBudget || !utf8.Valid(b) {
		return nil, errAudit
	}
	rows := make([]row, 0, pin.rows)
	s := bufio.NewScanner(bytes.NewReader(b))
	s.Buffer(make([]byte, 4096), 64<<10)
	for s.Scan() {
		var r row
		if len(rows) >= pin.rows || decodeStrict(s.Bytes(), &r.inputRow) != nil {
			return nil, errAudit
		}
		r.locale = index(r.Locale, []string{"ko", "en"})
		if r.Schema != "riido-three-claims-training-row-v1" || r.Split != "fit" || r.Annotation != "ai_semantic_reference" || r.Rubric != rubricSHA || r.locale < 0 || !identifier(r.ID) || !identifier(r.Family) || !identifier(r.Group) || r.ID != r.Family+"-"+r.Locale || len(r.Targets) != 3 || len(r.Evidence) != 3 || len(r.Text) > 4096 || strings.TrimSpace(r.Text) == "" {
			return nil, errAudit
		}
		word := false
		for _, c := range r.Text {
			word = word || unicode.IsLetter(c) || unicode.IsNumber(c)
		}
		if !word {
			return nil, errAudit
		}
		for h, target := range r.Targets {
			r.labels[h] = index(target, []string{"true", "false", "unknown"})
			ev := r.Evidence[h]
			if r.labels[h] < 0 || ev.Start == nil || ev.End == nil || *ev.Start > *ev.End || !boundary(r.Text, *ev.Start) || !boundary(r.Text, *ev.End) || len(ev.Reason) > 512 || !utf8.ValidString(ev.Reason) || strings.TrimSpace(ev.Reason) == "" {
				return nil, errAudit
			}
			if *ev.Start == *ev.End && (r.labels[h] != 1 || *ev.Start != 0) || *ev.End > *ev.Start && strings.TrimSpace(r.Text[*ev.Start:*ev.End]) == "" {
				return nil, errAudit
			}
		}
		rows = append(rows, r)
	}
	if s.Err() != nil || len(rows) != pin.rows {
		return nil, errAudit
	}
	return rows, nil
}

func summarize(rows []row, pin manifest) (report, error) {
	r := report{Schema: "riido-three-claims-training-audit-v1", TrainingSHA: pin.sha, LocaleOrder: [2]string{"ko", "en"}, HeadOrder: [3]string{"response_requested", "current_activity_claimed", "completion_claimed"}, StateOrder: [3]string{"true", "false", "unknown"}, LengthBounds: [4]int{64, 128, 256, 4096}, RuneBounds: [4]int{16, 32, 64, 4096}, DuplicateNormalization: "Unicode lowercase; Unicode whitespace collapse; no compatibility normalization", Limitations: "Training-only AI references; aggregate audit provides no model accuracy, independence or product qualification evidence."}
	// Sorting bounded metadata avoids maps in the aggregate passes.
	order := make([]int, len(rows))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return rows[order[i]].ID < rows[order[j]].ID })
	for i := 1; i < len(order); i++ {
		if rows[order[i-1]].ID == rows[order[i]].ID {
			return report{}, errAudit
		}
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := rows[order[i]], rows[order[j]]
		if a.Family != b.Family {
			return a.Family < b.Family
		}
		return a.locale < b.locale
	})
	for i := 0; i < len(order); i += 2 {
		if i+1 >= len(order) {
			return report{}, errAudit
		}
		a, b := rows[order[i]], rows[order[i+1]]
		if a.Family != b.Family || a.Group != b.Group || a.locale != 0 || b.locale != 1 {
			return report{}, errAudit
		}
		r.Families++
		for h := range 3 {
			if a.labels[h] != b.labels[h] {
				r.PairDifferences[h]++
			}
		}
	}
	sort.Slice(order, func(i, j int) bool { return rows[order[i]].Group < rows[order[j]].Group })
	for i := 0; i < len(order); i += 6 {
		if i+5 >= len(order) {
			return report{}, errAudit
		}
		var seen [2][3][3]bool
		var locales [2]int
		group := rows[order[i]].Group
		for j := i; j < i+6; j++ {
			x := rows[order[j]]
			if x.Group != group {
				return report{}, errAudit
			}
			locales[x.locale]++
			for h, s := range x.labels {
				seen[x.locale][h][s] = true
			}
		}
		if locales != [2]int{3, 3} || i+6 < len(order) && rows[order[i+6]].Group == group {
			return report{}, errAudit
		}
		r.Groups++
		for l := range 2 {
			r.GroupsWithLocale[l]++
			for h := range 3 {
				for s := range 3 {
					if seen[l][h][s] {
						r.GroupSupports[l][h][s]++
					}
				}
			}
		}
	}
	if r.Families != pin.families || r.Groups != pin.groups {
		return report{}, errAudit
	}
	type duplicateRow struct {
		normalized string
		labels     [3]int
	}
	normalized := make([]duplicateRow, len(rows))
	for i, x := range rows {
		r.Rows++
		r.LocaleRows[x.locale]++
		bucket := 0
		for len(x.Text) > r.LengthBounds[bucket] {
			bucket++
		}
		r.LengthRows[bucket]++
		r.LocaleLengthRows[x.locale][bucket]++
		runeBucket := 0
		runes := utf8.RuneCountInString(x.Text)
		for runes > r.RuneBounds[runeBucket] {
			runeBucket++
		}
		r.RuneRows[runeBucket]++
		r.LocaleRuneRows[x.locale][runeBucket]++
		for h, s := range x.labels {
			r.HeadStates[x.locale][h][s]++
			r.LengthStates[bucket][x.locale][h][s]++
			r.RuneStates[runeBucket][x.locale][h][s]++
			span := *x.Evidence[h].End - *x.Evidence[h].Start
			if span == 0 {
				r.EmptyEvidence[h][s]++
			} else {
				r.NonemptyEvidence[h][s]++
			}
			v := float64(span) / float64(len(x.Text))
			f := &r.SpanFractions[h][s]
			if f.Count == 0 || v < f.Min {
				f.Min = v
			}
			if v > f.Max {
				f.Max = v
			}
			f.Count++
			f.Mean += v
		}
		normalized[i] = duplicateRow{strings.ToLower(strings.Join(strings.Fields(x.Text), " ")), x.labels}
	}
	if r.LocaleRows != [2]int{pin.families, pin.families} {
		return report{}, errAudit
	}
	for h := range 3 {
		for s := range 3 {
			f := &r.SpanFractions[h][s]
			if f.Count > 0 {
				f.Mean /= float64(f.Count)
			}
		}
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].normalized < normalized[j].normalized })
	for i := 0; i < len(normalized); {
		j := i + 1
		for j < len(normalized) && normalized[j].normalized == normalized[i].normalized {
			j++
		}
		if j-i > 1 {
			r.Duplicates.Clusters++
			r.Duplicates.RedundantRows += j - i - 1
			conflicting := false
			for a := i; a < j; a++ {
				for b := a + 1; b < j; b++ {
					if normalized[a].labels != normalized[b].labels {
						r.Duplicates.ConflictingPairs++
						conflicting = true
					}
				}
			}
			if conflicting {
				r.Duplicates.ConflictingClusters++
			}
		}
		i = j
	}
	return r, nil
}

func run(args []string, dst io.Writer, root *os.Root, pin manifest) error {
	f := flag.NewFlagSet("riido-statehint-claims-training-audit", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	name := f.String("training", "", "relative path to the pinned public training JSONL")
	sha := f.String("training-sha256", "", "required immutable public training SHA-256")
	if f.Parse(args) != nil || f.NArg() != 0 || *sha != pin.sha {
		return errAudit
	}
	b, e := readPinned(root, *name, *sha, pin)
	if e != nil {
		return errAudit
	}
	rows, e := prepare(b, pin)
	if e != nil {
		return errAudit
	}
	r, e := summarize(rows, pin)
	if e != nil {
		return errAudit
	}
	return json.NewEncoder(dst).Encode(r)
}
func main() {
	root, e := os.OpenRoot(".")
	if e != nil {
		fmt.Fprintln(os.Stderr, errAudit)
		os.Exit(1)
	}
	defer root.Close()
	if e := run(os.Args[1:], os.Stdout, root, frozenManifest()); e != nil {
		fmt.Fprintln(os.Stderr, errAudit)
		os.Exit(1)
	}
}
