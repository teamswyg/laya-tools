// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

type evidenceInput struct {
	Start  *int   `json:"start_byte"`
	End    *int   `json:"end_byte"`
	Reason string `json:"reason"`
}
type rawRow struct {
	Schema           string          `json:"schema"`
	ID               string          `json:"id"`
	Family           string          `json:"family_id"`
	Group            string          `json:"leakage_group_id"`
	Split            string          `json:"internal_split"`
	Locale           string          `json:"locale"`
	Text             string          `json:"text"`
	Targets          []string        `json:"targets"`
	RubricSHA        string          `json:"rubric_sha256"`
	AnnotationSource string          `json:"annotation_source"`
	Evidence         []evidenceInput `json:"evidence"`
}
type row struct {
	rawRow
	labels [3]statehintclaims.State
}
type counts struct {
	Rows           int    `json:"rows"`
	Families       int    `json:"families"`
	Lineages       int    `json:"lineages"`
	Locales        [2]int `json:"locales_ko_en"`
	StrataRows     [2]int `json:"strata_rows_contrast_general"`
	StrataFamilies [2]int `json:"strata_families_contrast_general"`
	StrataLineages [2]int `json:"strata_lineages_contrast_general"`
}
type prepared struct {
	rows   []row
	counts counts
	rubric string
}
type stratumRow struct {
	ID      string `json:"id"`
	Family  string `json:"family_id"`
	Group   string `json:"leakage_group_id"`
	Locale  string `json:"locale"`
	Stratum string `json:"stratum"`
}
type strataInput struct {
	Schema string       `json:"schema"`
	Rows   []stratumRow `json:"rows"`
}
type strata struct {
	rows          []stratumRow
	counts        counts
	byID          map[string]int
	rowLineage    []int
	lineageStrata []int
}
type receipt struct {
	Schema           string                   `json:"schema"`
	TrainingSHA      string                   `json:"training_sha256"`
	EvaluationSHA    string                   `json:"evaluation_sha256"`
	EvaluationBytes  int64                    `json:"evaluation_bytes"`
	StrataSHA        string                   `json:"strata_sha256"`
	RowSchema        string                   `json:"row_schema"`
	AnnotationSource string                   `json:"annotation_source"`
	ReferenceStatus  string                   `json:"reference_status"`
	Locales          [2]string                `json:"locale_order"`
	Heads            [3]string                `json:"head_order"`
	States           [3]statehintclaims.State `json:"state_order"`
	EvaluationCounts counts                   `json:"evaluation_counts"`
	ContrastFamilies [2][3][3]int             `json:"contrast_reference_families_locale_head_state"`
	ContrastLineages [2][3][3]int             `json:"contrast_reference_lineages_locale_head_state"`
}

func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func validSHA(s string) bool {
	if len(s) != sha256.Size*2 {
		return false
	}
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == sha256.Size && hex.EncodeToString(b) == s
}
func local(s string) bool {
	return len(s) <= 4096 && filepath.IsLocal(s) && s != "." && filepath.Clean(s) == s && !strings.Contains(s, "\\")
}
func bounded(s string, n int) bool { return len(s) <= n && utf8.ValidString(s) }
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
func localeIndex(s string) int {
	if s == "ko" {
		return 0
	}
	if s == "en" {
		return 1
	}
	return -1
}
func stratumIndex(s string) int {
	if s == "contrast" {
		return 0
	}
	if s == "general" {
		return 1
	}
	return -1
}

// Reject duplicate keys as well as unknown fields; encoding/json otherwise
// silently accepts contradictory repeated schema, targets or evidence fields.
func uniqueJSON(data []byte) bool {
	d := json.NewDecoder(bytes.NewReader(data))
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
		if depth > 32 {
			return errStudy
		}
		defer func() { depth-- }()
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
					return errStudy
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
			return errStudy
		}
		_, e = d.Token()
		return e
	}
	if value() != nil {
		return false
	}
	_, e := d.Token()
	return e == io.EOF
}
func jsonShape(value any, t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		if value == nil {
			return true
		}
		return jsonShape(value, t.Elem())
	}
	switch t.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok || len(object) != t.NumField() {
			return false
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			v, ok := object[name]
			if !ok || !jsonShape(v, field.Type) {
				return false
			}
		}
	case reflect.Array, reflect.Slice:
		array, ok := value.([]any)
		if !ok || t.Kind() == reflect.Array && len(array) != t.Len() {
			return false
		}
		for _, v := range array {
			if !jsonShape(v, t.Elem()) {
				return false
			}
		}
	}
	return true
}
func decodeStrict(data []byte, v any) error {
	if !utf8.Valid(data) || !uniqueJSON(data) {
		return errStudy
	}
	var shape any
	if json.Unmarshal(data, &shape) != nil || !jsonShape(shape, reflect.TypeOf(v).Elem()) {
		return errStudy
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return errStudy
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errStudy
	}
	return nil
}

// os.Root prevents escape. Each component's Lstat/SameFile pair additionally
// rejects symlinks and replacement between inspection and opening.
func openDirectory(root *os.Root, name string) (*os.Root, error) {
	before, e := root.Lstat(name)
	if e != nil || !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
		return nil, errStudy
	}
	child, e := root.OpenRoot(name)
	if e != nil {
		return nil, errStudy
	}
	after, e := child.Stat(".")
	if e != nil || !after.IsDir() || !os.SameFile(before, after) {
		child.Close()
		return nil, errStudy
	}
	return child, nil
}
func openInput(root *os.Root, name string, limit, exact int) (*os.File, error) {
	if !local(name) {
		return nil, errStudy
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
		next, e := openDirectory(current, part)
		if e != nil {
			return nil, errStudy
		}
		if owned != nil {
			owned.Close()
		}
		owned, current = next, next
	}
	leaf := parts[len(parts)-1]
	before, e := current.Lstat(leaf)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > int64(limit) || exact > 0 && before.Size() != int64(exact) {
		return nil, errStudy
	}
	f, e := current.Open(leaf)
	if e != nil {
		return nil, errStudy
	}
	after, e := f.Stat()
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) || before.Size() != after.Size() {
		f.Close()
		return nil, errStudy
	}
	return f, nil
}
func readPinned(root *os.Root, name, sha string, limit, exact int) ([]byte, error) {
	if !validSHA(sha) {
		return nil, errStudy
	}
	f, e := openInput(root, name, limit, exact)
	if e != nil {
		return nil, errStudy
	}
	b, re := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	st, se := f.Stat()
	ce := f.Close()
	if re != nil || se != nil || ce != nil || len(b) < 1 || len(b) > limit || exact > 0 && len(b) != exact || st.Size() != int64(len(b)) || digest(b) != sha {
		return nil, errStudy
	}
	return b, nil
}

// inspectInput never opens the leaf. Check validates a sealed evaluation's
// existence, path and declared size without reading any body byte.
func inspectInput(root *os.Root, name string, limit int) (os.FileInfo, error) {
	if !local(name) {
		return nil, errStudy
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
		next, e := openDirectory(current, part)
		if e != nil {
			return nil, errStudy
		}
		if owned != nil {
			owned.Close()
		}
		owned, current = next, next
	}
	st, e := current.Lstat(parts[len(parts)-1])
	if e != nil || !st.Mode().IsRegular() || st.Size() < 1 || st.Size() > int64(limit) {
		return nil, errStudy
	}
	return st, nil
}
func sameSnapshot(a, b os.FileInfo) bool {
	return a != nil && b != nil && os.SameFile(a, b) && a.Size() == b.Size() && a.Mode() == b.Mode() && a.ModTime().Equal(b.ModTime())
}

func prepare(data []byte, split string) (prepared, error) {
	var p prepared
	if len(data) < 1 || len(data) > inputBudget || !utf8.Valid(data) || (split != "fit" && split != "dev") {
		return p, errStudy
	}
	s := bufio.NewScanner(bytes.NewReader(data))
	s.Buffer(make([]byte, 4096), 64<<10)
	var features statehintwide.Workspace
	for s.Scan() {
		line := bytes.TrimSpace(s.Bytes())
		if len(line) == 0 {
			return prepared{}, errStudy
		}
		if len(p.rows) >= 2048 {
			return prepared{}, errStudy
		}
		var r row
		if decodeStrict(line, &r.rawRow) != nil || r.Schema != rowSchema || r.Split != split || r.AnnotationSource != "ai_semantic_reference" || !validSHA(r.RubricSHA) || !identifier(r.ID) || !identifier(r.Family) || !identifier(r.Group) || localeIndex(r.Locale) < 0 || r.ID != r.Family+"-"+r.Locale || len(r.Targets) != 3 || len(r.Evidence) != 3 || !bounded(r.Text, statehintclaims.MaxTextBytes) || strings.TrimSpace(r.Text) == "" {
			return prepared{}, errStudy
		}
		if p.rubric == "" {
			p.rubric = r.RubricSHA
		} else if p.rubric != r.RubricSHA {
			return prepared{}, errStudy
		}
		for h, target := range r.Targets {
			label, ok := statehintclaims.ParseState(target)
			e := r.Evidence[h]
			if !ok || e.Start == nil || e.End == nil || *e.Start > *e.End || !boundary(r.Text, *e.Start) || !boundary(r.Text, *e.End) || !bounded(e.Reason, 512) || strings.TrimSpace(e.Reason) == "" {
				return prepared{}, errStudy
			}
			if label != statehintclaims.False && *e.End <= *e.Start || label == statehintclaims.False && *e.Start == *e.End && *e.Start != 0 {
				return prepared{}, errStudy
			}
			if *e.End > *e.Start && strings.TrimSpace(r.Text[*e.Start:*e.End]) == "" {
				return prepared{}, errStudy
			}
			r.labels[h] = label
		}
		v, e := statehintwide.ExtractContextual(r.Text, &features)
		if e != nil || v.WordCount() == 0 {
			return prepared{}, errStudy
		}
		p.rows = append(p.rows, r)
	}
	if s.Err() != nil || len(p.rows) == 0 {
		return prepared{}, errStudy
	}
	meta := make([]stratumRow, len(p.rows))
	for i, r := range p.rows {
		meta[i] = stratumRow{ID: r.ID, Family: r.Family, Group: r.Group, Locale: r.Locale}
	}
	c, e := metadataCounts(meta, false)
	if e != nil {
		return prepared{}, errStudy
	}
	p.counts = c
	return p, nil
}

// Metadata is used only for split validation and lineage aggregation. Model
// samples and scoring calls below receive text and supervised targets alone.
func metadataCounts(rows []stratumRow, stratified bool) (counts, error) {
	var c counts
	ids := make(map[string]bool, len(rows))
	families := make(map[string][]stratumRow, len(rows)/2)
	groups := make(map[string][]stratumRow, len(rows)/6)
	for _, r := range rows {
		l := localeIndex(r.Locale)
		s := stratumIndex(r.Stratum)
		if !identifier(r.ID) || !identifier(r.Family) || !identifier(r.Group) || l < 0 || r.ID != r.Family+"-"+r.Locale || ids[r.ID] || stratified && s < 0 || !stratified && r.Stratum != "" {
			return counts{}, errStudy
		}
		ids[r.ID] = true
		families[r.Family] = append(families[r.Family], r)
		groups[r.Group] = append(groups[r.Group], r)
		c.Rows++
		c.Locales[l]++
		if stratified {
			c.StrataRows[s]++
		}
	}
	for _, f := range families {
		if len(f) != 2 || f[0].Locale == f[1].Locale || f[0].Group != f[1].Group || f[0].Stratum != f[1].Stratum {
			return counts{}, errStudy
		}
		c.Families++
		if stratified {
			c.StrataFamilies[stratumIndex(f[0].Stratum)]++
		}
	}
	for _, g := range groups {
		if len(g) != 6 {
			return counts{}, errStudy
		}
		for _, r := range g {
			if r.Stratum != g[0].Stratum {
				return counts{}, errStudy
			}
		}
		c.Lineages++
		if stratified {
			c.StrataLineages[stratumIndex(g[0].Stratum)]++
		}
	}
	return c, nil
}
func trainingScope(c counts) bool {
	return c.Rows == 1680 && c.Families == 840 && c.Lineages == 280 && c.Locales == [2]int{840, 840}
}
func evaluationScope(c counts) bool {
	return c == counts{Rows: 360, Families: 180, Lineages: 60, Locales: [2]int{180, 180}, StrataRows: [2]int{240, 120}, StrataFamilies: [2]int{120, 60}, StrataLineages: [2]int{40, 20}}
}
func prepareStrata(data []byte, train prepared) (strata, error) {
	var input strataInput
	if len(data) > metadataBudget || decodeStrict(data, &input) != nil || input.Schema != strataSchema {
		return strata{}, errStudy
	}
	c, e := metadataCounts(input.Rows, true)
	if e != nil || !evaluationScope(c) {
		return strata{}, errStudy
	}
	used := make(map[string]bool, 3*len(train.rows))
	for _, r := range train.rows {
		used["i:"+r.ID] = true
		used["f:"+r.Family] = true
		used["g:"+r.Group] = true
	}
	result := strata{rows: input.Rows, counts: c, byID: make(map[string]int, len(input.Rows))}
	groupStrata := make(map[string]int, c.Lineages)
	for i, r := range input.Rows {
		if used["i:"+r.ID] || used["f:"+r.Family] || used["g:"+r.Group] {
			return strata{}, errStudy
		}
		result.byID[r.ID] = i
		groupStrata[r.Group] = stratumIndex(r.Stratum)
	}
	groups := make([]string, 0, len(groupStrata))
	for group := range groupStrata {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	groupIndex := make(map[string]int, len(groups))
	for i, group := range groups {
		groupIndex[group] = i
		result.lineageStrata = append(result.lineageStrata, groupStrata[group])
	}
	result.rowLineage = make([]int, len(input.Rows))
	for i, r := range input.Rows {
		result.rowLineage[i] = groupIndex[r.Group]
	}
	return result, nil
}
func validateReceipt(data []byte, trainingSHA, evaluationSHA, strataSHA string, s strata) (receipt, error) {
	var r receipt
	if len(data) > metadataBudget || decodeStrict(data, &r) != nil || r.Schema != receiptSchema || r.TrainingSHA != trainingSHA || r.EvaluationSHA != evaluationSHA || r.EvaluationBytes < 1 || r.EvaluationBytes > inputBudget || r.StrataSHA != strataSHA || r.RowSchema != rowSchema || r.AnnotationSource != "ai_semantic_reference" || r.ReferenceStatus != "ai_authored_independently_ai_reviewed_human_pending" || r.EvaluationCounts != s.counts || r.Locales != [2]string{"ko", "en"} || r.Heads != headOrder() || r.States != statehintclaims.States() {
		return receipt{}, errStudy
	}
	for l := range r.ContrastFamilies {
		for h := range r.ContrastFamilies[l] {
			for k := range r.ContrastFamilies[l][h] {
				if r.ContrastFamilies[l][h][k] < 20 || r.ContrastFamilies[l][h][k] > 120 || r.ContrastLineages[l][h][k] < 10 || r.ContrastLineages[l][h][k] > 40 || r.ContrastLineages[l][h][k] > r.ContrastFamilies[l][h][k] {
					return receipt{}, errStudy
				}
			}
			if r.ContrastFamilies[l][h][0]+r.ContrastFamilies[l][h][1]+r.ContrastFamilies[l][h][2] != 120 {
				return receipt{}, errStudy
			}
		}
	}
	return r, nil
}
func headOrder() [3]string {
	var r [3]string
	for i, h := range statehintclaims.Heads() {
		r[i] = h.String()
	}
	return r
}
func validateEvaluation(p, train prepared, s strata, receipt receipt) error {
	if p.counts.Rows != 360 || p.counts.Families != 180 || p.counts.Lineages != 60 || p.counts.Locales != [2]int{180, 180} || p.rubric != train.rubric {
		return errStudy
	}
	var support [2][3][3]int
	var groupSupport [2][3][3][]string
	for _, r := range p.rows {
		i, ok := s.byID[r.ID]
		if !ok {
			return errStudy
		}
		m := s.rows[i]
		if r.Family != m.Family || r.Group != m.Group || r.Locale != m.Locale {
			return errStudy
		}
		if m.Stratum == "contrast" {
			l := localeIndex(r.Locale)
			for h, t := range r.labels {
				k, _ := statehintclaims.StateIndex(t)
				support[l][h][k]++
				groupSupport[l][h][k] = append(groupSupport[l][h][k], r.Group)
			}
		}
	}
	for l := range support {
		for h := range support[l] {
			for k := range support[l][h] {
				if support[l][h][k] != receipt.ContrastFamilies[l][h][k] || unique(groupSupport[l][h][k]) != receipt.ContrastLineages[l][h][k] {
					return errStudy
				}
			}
		}
	}
	return nil
}
func unique(v []string) int {
	sort.Strings(v)
	n := 0
	for i, s := range v {
		if i == 0 || s != v[i-1] {
			n++
		}
	}
	return n
}
