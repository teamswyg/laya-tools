// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only fixed semantic-data study. AI references do not qualify a model.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	anchor           = ".cache/statehint-claims-contrast-study"
	inputBudget      = 8 << 20
	metadataBudget   = 1 << 20
	rowSchema        = "riido-three-claims-training-row-v1"
	strataSchema     = "riido-three-claims-semantic-contrast-strata-v1"
	receiptSchema    = "riido-three-claims-semantic-contrast-writer-receipt-v1"
	parentBytes      = 73988
	nllFloor         = 1e-15
	bootstrapSamples = 10000
)

var errStudy = errors.New("semantic contrast study pinned input or private output invalid")
var errEvaluationAlias = errors.New("pinned input aliases sealed evaluation")

type manifest struct{ parent, comparator string }

func frozenManifest() manifest {
	return manifest{"2e68d46a8452d11721954b75abd9fac97f533b0796e930ae5f79fb050b040709", "b534594a361b8d31bda599bf53b92ee50ecd55821455914514232e7789f381ec"}
}

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
	Schema                     string                   `json:"schema"`
	TrainingSHA                string                   `json:"training_sha256"`
	EvaluationSHA              string                   `json:"evaluation_sha256"`
	EvaluationBytes            int64                    `json:"evaluation_bytes"`
	StrataSHA                  string                   `json:"strata_sha256"`
	RowSchema                  string                   `json:"row_schema"`
	AnnotationSource           string                   `json:"annotation_source"`
	ReferenceStatus            string                   `json:"reference_status"`
	Locales                    [2]string                `json:"locale_order"`
	Heads                      [3]string                `json:"head_order"`
	States                     [3]statehintclaims.State `json:"state_order"`
	EvaluationCounts           counts                   `json:"evaluation_counts"`
	ContrastFamilies           [2][3][3]int             `json:"contrast_reference_families_locale_head_state"`
	ContrastLineages           [2][3][3]int             `json:"contrast_reference_lineages_locale_head_state"`
	NewTrainingFamilies        [2][3][3]int             `json:"new_training_reference_families_locale_head_state"`
	NewTrainingUnknownLineages [2][3]int                `json:"new_training_unknown_lineages_locale_head"`
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

// Inspect every input before reading any bytes. Another pinned role must not
// refer to the evaluation leaf by either pathname or hardlink identity, even
// if that role would later reject its schema or digest. This is also required
// in actual mode so pre-fit admission cannot expose evaluation through a role.
func preflightInputs(root *os.Root, p pins) (os.FileInfo, error) {
	evaluation, e := inspectInput(root, p.evaluationName, inputBudget)
	if e != nil {
		return nil, errStudy
	}
	for _, input := range []struct {
		name         string
		limit, exact int
	}{{p.parentName, parentBytes, parentBytes}, {p.trainingName, inputBudget, 0}, {p.strataName, metadataBudget, 0}, {p.receiptName, metadataBudget, 0}, {p.comparatorName, parentBytes, parentBytes}} {
		if filepath.Clean(input.name) == filepath.Clean(p.evaluationName) {
			return nil, errEvaluationAlias
		}
		st, e := inspectInput(root, input.name, input.limit)
		if e != nil {
			return nil, errStudy
		}
		if os.SameFile(evaluation, st) {
			return nil, errEvaluationAlias
		}
		if input.exact > 0 && st.Size() != int64(input.exact) {
			return nil, errStudy
		}
	}
	return evaluation, nil
}
func readEvaluation(root *os.Root, p pins, before os.FileInfo) ([]byte, error) {
	f, e := openInput(root, p.evaluationName, inputBudget, 0)
	if e != nil {
		return nil, errStudy
	}
	st, e := f.Stat()
	if e != nil || !sameSnapshot(before, st) {
		f.Close()
		return nil, errStudy
	}
	b, re := io.ReadAll(io.LimitReader(f, inputBudget+1))
	after, se := f.Stat()
	ce := f.Close()
	if re != nil || se != nil || ce != nil || !sameSnapshot(before, after) || len(b) != int(before.Size()) || digest(b) != p.evaluationSHA {
		return nil, errStudy
	}
	return b, nil
}
func modelShape(data []byte, steps uint64) bool {
	if len(data) != parentBytes || parentBytes != statehintclaims.ArtifactBytes || string(data[:4]) != "RSC\x00" {
		return false
	}
	end := len(data) - sha256.Size
	s := sha256.Sum256(data[:end])
	var schema [64]byte
	copy(schema[:], statehintclaims.FeatureSchema)
	var order bytes.Buffer
	for _, h := range statehintclaims.Heads() {
		order.WriteString(h.String())
		order.WriteByte(0)
	}
	for _, state := range statehintclaims.States() {
		order.WriteString(string(state))
		order.WriteByte(0)
	}
	orderSHA := sha256.Sum256(order.Bytes())
	contractSHA := sha256.Sum256([]byte("feature-major-nine-columns;mean-three-categorical-ce;fresh-zero-adamw;bias-no-decay;fixed-T1;v1"))
	return bytes.Equal(s[:], data[end:]) && binary.LittleEndian.Uint16(data[4:6]) == 1 && binary.LittleEndian.Uint16(data[6:8]) == 192 && binary.LittleEndian.Uint16(data[8:10]) == 2048 && bytes.Equal(data[10:16], []byte{3, 3, 1, 0, 0, 0}) && binary.LittleEndian.Uint64(data[16:24]) == steps && binary.LittleEndian.Uint64(data[24:32]) == 1729 && bytes.Equal(data[32:96], schema[:]) && bytes.Equal(data[96:128], orderSHA[:]) && bytes.Equal(data[128:160], contractSHA[:]) && binary.LittleEndian.Uint64(data[160:168]) == math.Float64bits(.9) && binary.LittleEndian.Uint64(data[168:176]) == math.Float64bits(.05) && binary.LittleEndian.Uint64(data[176:184]) == math.Float64bits(1) && bytes.Equal(data[184:192], make([]byte, 8))
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
			if r.NewTrainingUnknownLineages[l][h] < 15 || r.NewTrainingUnknownLineages[l][h] > 40 {
				return receipt{}, errStudy
			}
			for k := range r.ContrastFamilies[l][h] {
				if r.ContrastFamilies[l][h][k] < 20 || r.ContrastFamilies[l][h][k] > 120 || r.ContrastLineages[l][h][k] < 10 || r.ContrastLineages[l][h][k] > 40 || r.ContrastLineages[l][h][k] > r.ContrastFamilies[l][h][k] || r.NewTrainingFamilies[l][h][k] < 30 || r.NewTrainingFamilies[l][h][k] > 120 {
					return receipt{}, errStudy
				}
			}
			if r.ContrastFamilies[l][h][0]+r.ContrastFamilies[l][h][1]+r.ContrastFamilies[l][h][2] != 120 || r.NewTrainingFamilies[l][h][0]+r.NewTrainingFamilies[l][h][1]+r.NewTrainingFamilies[l][h][2] != 120 {
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

type localeReport struct {
	Rows                int           `json:"rows"`
	ReferenceCounts     [3][3]int     `json:"reference_counts_head_true_false_unknown"`
	ReferenceFamilies   [3][3]int     `json:"reference_family_support_head_true_false_unknown"`
	ReferenceLineages   [3][3]int     `json:"reference_lineage_support_head_true_false_unknown"`
	RawConfusion        [3][3][3]int  `json:"raw_confusion_head_target_winner"`
	MeanCE              float64       `json:"mean_three_head_cross_entropy"`
	HeadCE              [3]float64    `json:"mean_cross_entropy_by_head"`
	TargetCE            [3][3]float64 `json:"mean_cross_entropy_by_head_target"`
	TargetCEDefined     [3][3]bool    `json:"cross_entropy_by_head_target_defined"`
	UnknownTargetCE     float64       `json:"mean_unknown_target_cross_entropy"`
	UnknownTargetCount  int           `json:"unknown_target_count"`
	TrueProposed        [3]int        `json:"gated_true_proposed"`
	TrueCorrect         [3]int        `json:"gated_true_correct"`
	TrueFalsePositive   [3]int        `json:"gated_true_false_positive"`
	TrueOnFalse         [3]int        `json:"gated_true_on_false_reference"`
	TrueOnUnknown       [3]int        `json:"gated_true_on_unknown_reference"`
	TrueCorrectFamilies [3]int        `json:"gated_true_correct_family_support"`
	TrueCorrectLineages [3]int        `json:"gated_true_correct_lineage_support"`
	Precision           [3]float64    `json:"gated_true_precision"`
	PrecisionDefined    [3]bool       `json:"gated_true_precision_defined"`
	UnknownReasons      [3][5]int     `json:"emitted_unknown_reasons"`
	ClippedTargets      [3]int        `json:"target_probabilities_clipped_at_floor"`
}
type evaluationReport struct {
	Rows             int             `json:"rows"`
	MeanCE           float64         `json:"mean_three_head_cross_entropy"`
	ProbabilityFloor float64         `json:"cross_entropy_probability_floor"`
	ReasonOrder      [5]string       `json:"unknown_reason_order"`
	Locales          [2]localeReport `json:"locales_ko_en"`
}
type observation struct {
	prediction statehintclaims.Prediction
	score      statehintclaims.ScoreResult
}

func observe(m *statehintclaims.Model, p prepared) ([]observation, error) {
	r := make([]observation, len(p.rows))
	var w statehintclaims.Workspace
	for i, row := range p.rows {
		prediction, e := m.Predict(row.Text, &w)
		if e != nil || prediction.Source != statehintclaims.Learned || prediction.TrainingSteps != m.TrainingSteps() {
			return nil, errStudy
		}
		score, e := m.Scores(row.Text, &w)
		if e != nil || score.WordCount == 0 {
			return nil, errStudy
		}
		r[i] = observation{prediction, score}
	}
	return r, nil
}
func rowCE(r row, o observation) (float64, error) {
	var total float64
	for h, t := range r.labels {
		k, ok := statehintclaims.StateIndex(t)
		if !ok {
			return 0, errStudy
		}
		p := o.score.Probabilities[h][k]
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return 0, errStudy
		}
		total -= math.Log(math.Max(p, nllFloor)) / 3
	}
	return total, nil
}
func evaluate(p prepared, obs []observation) (evaluationReport, error) {
	if len(obs) != len(p.rows) || len(obs) == 0 {
		return evaluationReport{}, errStudy
	}
	r := evaluationReport{Rows: len(obs), ProbabilityFloor: nllFloor, ReasonOrder: [5]string{"semantic_unknown", "low_confidence", "low_margin", "untrained", "no_word_content"}}
	var groups [2][3][3][]string
	var correct [2][3][]string
	for i, input := range p.rows {
		l := localeIndex(input.Locale)
		report := &r.Locales[l]
		report.Rows++
		for h, head := range obs[i].prediction.Heads {
			k, ok := statehintclaims.StateIndex(input.labels[h])
			winner, wok := statehintclaims.StateIndex(head.Winner)
			_, sok := statehintclaims.StateIndex(head.State)
			prob := obs[i].score.Probabilities[h][k]
			if !ok || !wok || !sok || head.Head != statehintclaims.Head(h).String() || head.Probabilities != obs[i].score.Probabilities[h] || math.IsNaN(prob) || math.IsInf(prob, 0) || prob < 0 || prob > 1 {
				return evaluationReport{}, errStudy
			}
			ce := -math.Log(math.Max(prob, nllFloor))
			report.ReferenceCounts[h][k]++
			report.ReferenceFamilies[h][k]++
			groups[l][h][k] = append(groups[l][h][k], input.Group)
			report.RawConfusion[h][k][winner]++
			report.MeanCE += ce / 3
			report.HeadCE[h] += ce
			report.TargetCE[h][k] += ce
			if prob < nllFloor {
				report.ClippedTargets[h]++
			}
			if k == 2 {
				report.UnknownTargetCount++
				report.UnknownTargetCE += ce
			}
			if head.State == statehintclaims.Unknown {
				found := false
				for j, reason := range r.ReasonOrder {
					if head.UnknownReason == reason {
						report.UnknownReasons[h][j]++
						found = true
						break
					}
				}
				if !found {
					return evaluationReport{}, errStudy
				}
			} else if head.UnknownReason != "" {
				return evaluationReport{}, errStudy
			}
			if head.State == statehintclaims.True {
				report.TrueProposed[h]++
				if k == 0 {
					report.TrueCorrect[h]++
					report.TrueCorrectFamilies[h]++
					correct[l][h] = append(correct[l][h], input.Group)
				} else {
					report.TrueFalsePositive[h]++
					if k == 1 {
						report.TrueOnFalse[h]++
					} else {
						report.TrueOnUnknown[h]++
					}
				}
			}
		}
	}
	for l := range r.Locales {
		report := &r.Locales[l]
		if report.Rows == 0 {
			return evaluationReport{}, errStudy
		}
		r.MeanCE += report.MeanCE
		report.MeanCE /= float64(report.Rows)
		for h := range report.HeadCE {
			report.HeadCE[h] /= float64(report.Rows)
			report.TrueCorrectLineages[h] = unique(correct[l][h])
			if report.TrueProposed[h] > 0 {
				report.PrecisionDefined[h] = true
				report.Precision[h] = float64(report.TrueCorrect[h]) / float64(report.TrueProposed[h])
			}
			for k := range report.TargetCE[h] {
				report.ReferenceLineages[h][k] = unique(groups[l][h][k])
				if report.ReferenceCounts[h][k] > 0 {
					report.TargetCEDefined[h][k] = true
					report.TargetCE[h][k] /= float64(report.ReferenceCounts[h][k])
				}
			}
		}
		if report.UnknownTargetCount > 0 {
			report.UnknownTargetCE /= float64(report.UnknownTargetCount)
		}
	}
	r.MeanCE /= float64(r.Rows)
	return r, nil
}

type pairedReport struct {
	Difference           float64    `json:"candidate_minus_matched_float_mean_three_head_ce"`
	StrataDifference     [2]float64 `json:"mean_delta_contrast_general"`
	ConfidenceInterval   [2]float64 `json:"paired_lineage_bootstrap_percentile_95pct_ci"`
	Lineages             [2]int     `json:"resampled_lineages_contrast_general"`
	Weights              [2]int     `json:"fixed_strata_weights_contrast_general"`
	Resamples            int        `json:"resamples"`
	Seed                 int64      `json:"seed"`
	Unit                 string     `json:"resampling_unit"`
	PercentileConvention string     `json:"percentile_convention"`
}
type lineageDelta struct {
	sum     float64
	rows    int
	stratum int
}

func paired(p prepared, s strata, candidate, comparator []observation) (pairedReport, error) {
	if len(candidate) != len(p.rows) || len(comparator) != len(p.rows) {
		return pairedReport{}, errStudy
	}
	groups := make([]lineageDelta, len(s.lineageStrata))
	for i, stratum := range s.lineageStrata {
		groups[i].stratum = stratum
	}
	for i, r := range p.rows {
		index, ok := s.byID[r.ID]
		if !ok {
			return pairedReport{}, errStudy
		}
		a, e := rowCE(r, candidate[i])
		if e != nil {
			return pairedReport{}, e
		}
		b, e := rowCE(r, comparator[i])
		if e != nil {
			return pairedReport{}, e
		}
		g := &groups[s.rowLineage[index]]
		g.rows++
		g.sum += a - b
	}
	var values [2][]float64
	for _, g := range groups {
		if g.rows != 6 || g.stratum < 0 {
			return pairedReport{}, errStudy
		}
		values[g.stratum] = append(values[g.stratum], g.sum/float64(g.rows))
	}
	return bootstrap(values)
}
func bootstrap(values [2][]float64) (pairedReport, error) {
	r := pairedReport{Weights: [2]int{2, 1}, Resamples: bootstrapSamples, Seed: 1729, Unit: "paired_leakage_group_mean_of_six_rows", PercentileConvention: "linear_interpolation_at_q_times_n_minus_one_q_0.025_0.975"}
	for s, v := range values {
		if len(v) == 0 {
			return pairedReport{}, errStudy
		}
		r.Lineages[s] = len(v)
		for _, x := range v {
			if math.IsNaN(x) || math.IsInf(x, 0) {
				return pairedReport{}, errStudy
			}
			r.StrataDifference[s] += x
		}
		r.StrataDifference[s] /= float64(len(v))
	}
	r.Difference = (2*r.StrataDifference[0] + r.StrataDifference[1]) / 3
	random := rand.New(rand.NewPCG(1729, 1729^0x9e3779b97f4a7c15))
	draws := make([]float64, bootstrapSamples)
	for i := range draws {
		var means [2]float64
		for s, v := range values {
			for range len(v) {
				means[s] += v[random.IntN(len(v))]
			}
			means[s] /= float64(len(v))
		}
		draws[i] = (2*means[0] + means[1]) / 3
	}
	sort.Float64s(draws)
	r.ConfidenceInterval = [2]float64{interpolatedPercentile(draws, .025), interpolatedPercentile(draws, .975)}
	return r, nil
}

// sorted is nonempty, sorted and finite; q is in [0,1]. Bootstrap fixes its
// endpoints at .025/.975, interpolating at q*(N-1), never nearest rank.
func interpolatedPercentile(sorted []float64, q float64) float64 {
	at := q * float64(len(sorted)-1)
	low := int(at)
	high := min(low+1, len(sorted)-1)
	return sorted[low] + (at-float64(low))*(sorted[high]-sorted[low])
}

type progressReport struct {
	PrimaryCIUpperBelowZero   bool       `json:"primary_ci_upper_below_zero"`
	UnknownCELower            [2]bool    `json:"unknown_target_ce_lower_ko_en"`
	NoAddedTrueFalsePositives [2][3]bool `json:"no_increased_gated_true_false_positive_locale_head"`
	ResearchProgress          bool       `json:"research_progress_all_frozen_conditions"`
}

func progress(delta pairedReport, candidate, comparator evaluationReport) progressReport {
	r := progressReport{PrimaryCIUpperBelowZero: delta.ConfidenceInterval[1] < 0}
	r.ResearchProgress = r.PrimaryCIUpperBelowZero
	for l := range r.UnknownCELower {
		a, b := candidate.Locales[l], comparator.Locales[l]
		r.UnknownCELower[l] = a.UnknownTargetCount > 0 && b.UnknownTargetCount == a.UnknownTargetCount && a.UnknownTargetCE < b.UnknownTargetCE
		r.ResearchProgress = r.ResearchProgress && r.UnknownCELower[l]
		for h := range r.NoAddedTrueFalsePositives[l] {
			r.NoAddedTrueFalsePositives[l][h] = a.TrueFalsePositive[h] <= b.TrueFalsePositive[h]
			r.ResearchProgress = r.ResearchProgress && r.NoAddedTrueFalsePositives[l][h]
		}
	}
	return r
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errStudy
	}
	if e := root.Mkdir(".cache", 0700); e != nil && !errors.Is(e, os.ErrExist) {
		return nil, errStudy
	}
	cache, e := openDirectory(root, ".cache")
	if e != nil {
		return nil, errStudy
	}
	defer cache.Close()
	base := filepath.Base(anchor)
	if e := cache.Mkdir(base, 0700); e != nil && !errors.Is(e, os.ErrExist) {
		return nil, errStudy
	}
	study, e := openDirectory(cache, base)
	if e != nil {
		return nil, errStudy
	}
	defer study.Close()
	st, e := study.Stat(".")
	if e != nil || st.Mode().Perm() != 0700 {
		return nil, errStudy
	}
	child := filepath.Base(name)
	if study.Mkdir(child, 0700) != nil {
		return nil, errStudy
	}
	out, e := openDirectory(study, child)
	if e != nil {
		return nil, errStudy
	}
	st, e = out.Stat(".")
	if e != nil || st.Mode().Perm() != 0700 {
		out.Close()
		return nil, errStudy
	}
	return out, nil
}
func writeBytes(root *os.Root, name string, data []byte) error {
	f, e := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errStudy
	}
	n, e := f.Write(data)
	if e == nil && n != len(data) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil || ce != nil {
		return errStudy
	}
	return nil
}

type recipe struct {
	Epochs         int        `json:"epochs"`
	Batch          int        `json:"batch_size"`
	Rate           float64    `json:"learning_rate"`
	Decay          float64    `json:"weight_decay"`
	Seed           int64      `json:"seed"`
	Temperature    float64    `json:"temperature"`
	Gate           [2]float64 `json:"confidence_margin"`
	Initialization string     `json:"initialization"`
}

func fixedRecipe() recipe {
	return recipe{40, 32, .001, .01, 1729, 1, [2]float64{.9, .05}, "parent_float_copy_fresh_adamw"}
}

type modelReport struct {
	SHA                string `json:"sha256"`
	ArtifactBytes      int    `json:"artifact_bytes"`
	TrainingSteps      uint64 `json:"training_steps"`
	InitializationSeed int64  `json:"initialization_seed"`
	ReadOnlyExact      bool   `json:"read_only_exact_save_parity"`
}
type candidateReport struct {
	modelReport
	ReloadByteExact       bool `json:"reload_exact_byte_parity"`
	ReloadPredictExact    bool `json:"reload_exact_prediction_parity"`
	ReloadScoresExact     bool `json:"reload_exact_score_parity"`
	PredictionCalls       int  `json:"evaluation_prediction_calls"`
	ScoreCalls            int  `json:"evaluation_score_calls"`
	ReloadPredictionCalls int  `json:"reload_prediction_calls"`
	ReloadScoreCalls      int  `json:"reload_score_calls"`
}
type studyReport struct {
	Schema               string                        `json:"schema"`
	Status               string                        `json:"status"`
	TrainingSHA          string                        `json:"training_sha256"`
	EvaluationSHA        string                        `json:"evaluation_sha256"`
	StrataSHA            string                        `json:"strata_sha256"`
	ReceiptSHA           string                        `json:"writer_receipt_sha256"`
	ReferenceStatus      string                        `json:"reference_status"`
	TrainingCounts       counts                        `json:"training_counts"`
	EvaluationCounts     counts                        `json:"evaluation_counts"`
	Heads                [3]string                     `json:"head_order"`
	States               [3]statehintclaims.State      `json:"state_order"`
	Recipe               recipe                        `json:"fixed_recipe"`
	Parent               modelReport                   `json:"original_float_parent"`
	Comparator           modelReport                   `json:"fixed_matched_float_comparator"`
	Candidate            candidateReport               `json:"candidate"`
	Fit                  statehintclaims.WarmFitReport `json:"one_float_warm_fit"`
	FitLossMeaning       string                        `json:"mean_fit_loss_meaning"`
	FitCalls             int                           `json:"actual_new_fit_calls"`
	CandidateEvaluation  evaluationReport              `json:"candidate_final_evaluation"`
	ComparatorEvaluation evaluationReport              `json:"matched_float_final_evaluation"`
	Primary              pairedReport                  `json:"primary_paired_comparison"`
	Progress             progressReport                `json:"frozen_research_progress_rule"`
	Qualified            bool                          `json:"semantic_quality_qualified"`
	QualificationChanged bool                          `json:"existing_qualification_rules_changed"`
	Selection            bool                          `json:"selection_performed"`
	Calibration          bool                          `json:"calibration_performed"`
	CalTestAccessed      bool                          `json:"original_or_existing_calibration_test_accessed"`
	StateWrites          int                           `json:"application_state_writes"`
	FitBeforeEvaluation  bool                          `json:"fit_completed_before_evaluation_body_and_comparator_load"`
	ElapsedNS            int64                         `json:"fit_evaluation_save_reload_nanoseconds"`
	GoHeap               uint64                        `json:"go_heap_alloc_bytes_not_os_rss"`
	OSRSSMeasured        bool                          `json:"os_rss_measured_in_process"`
	OSRSSMeasurement     string                        `json:"os_rss_measurement"`
	CPUProfile           bool                          `json:"private_cpu_profile_written"`
}

func canonicalModel(data []byte, steps uint64) (*statehintclaims.Model, error) {
	m, e := statehintclaims.Load(bytes.NewReader(data))
	if e != nil || m.TrainingSteps() != steps || m.InitializationSeed() != 1729 || m.Temperature() != 1 {
		return nil, errStudy
	}
	var b bytes.Buffer
	if m.Save(&b) != nil || !bytes.Equal(data, b.Bytes()) {
		return nil, errStudy
	}
	return m, nil
}
func modelSummary(m *statehintclaims.Model, data []byte) modelReport {
	return modelReport{digest(data), len(data), m.TrainingSteps(), m.InitializationSeed(), true}
}
func saveCandidate(out *os.Root, m *statehintclaims.Model, p prepared, obs []observation) (candidateReport, error) {
	var b bytes.Buffer
	if m.Save(&b) != nil || b.Len() != parentBytes {
		return candidateReport{}, errStudy
	}
	stored, e := readPinned(out, "candidate.rsc", digest(b.Bytes()), parentBytes, parentBytes)
	if e != nil {
		return candidateReport{}, errStudy
	}
	loaded, e := canonicalModel(stored, 4240)
	if e != nil {
		return candidateReport{}, e
	}
	round, e := observe(loaded, p)
	if e != nil || len(round) != len(obs) {
		return candidateReport{}, errStudy
	}
	for i := range round {
		if round[i] != obs[i] {
			return candidateReport{}, errStudy
		}
	}
	return candidateReport{modelReport: modelSummary(m, b.Bytes()), ReloadByteExact: true, ReloadPredictExact: true, ReloadScoresExact: true, PredictionCalls: len(obs), ScoreCalls: len(obs), ReloadPredictionCalls: len(obs), ReloadScoreCalls: len(obs)}, nil
}

type pins struct{ parentName, parentSHA, comparatorName, comparatorSHA, trainingName, trainingSHA, evaluationName, evaluationSHA, strataName, strataSHA, receiptName, receiptSHA string }

func fitStudy(root, out *os.Root, parent *statehintclaims.Model, parentData []byte, train prepared, s strata, receipt receipt, p pins, evaluationSnapshot os.FileInfo, cpuProfile bool) (result studyReport, err error) {
	started := time.Now()
	phase := "before_fit"
	fitCalls := 0
	defer func() {
		if err != nil && fitCalls > 0 {
			record := struct {
				Schema      string `json:"schema"`
				Status      string `json:"status"`
				FitCalls    int    `json:"actual_new_fit_calls"`
				Phase       string `json:"failure_phase"`
				TrainingSHA string `json:"training_sha256"`
				ParentSHA   string `json:"parent_sha256"`
				NoRetry     bool   `json:"no_implicit_retry"`
			}{"riido-three-claims-semantic-contrast-fit-consumption-v1", "failed_after_fit_invocation_consumed", fitCalls, phase, p.trainingSHA, p.parentSHA, true}
			encoded, e := json.MarshalIndent(record, "", "  ")
			if e != nil || writeBytes(out, "fit-consumption.json", append(encoded, '\n')) != nil {
				err = errStudy
			}
		}
	}()
	var profile *os.File
	finishProfile := func() error {
		if profile == nil {
			return nil
		}
		pprof.StopCPUProfile()
		syncErr := profile.Sync()
		closeErr := profile.Close()
		profile = nil
		if syncErr != nil || closeErr != nil {
			return errStudy
		}
		return nil
	}
	if cpuProfile {
		profile, err = out.OpenFile("cpu.pprof", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return result, errStudy
		}
		if pprof.StartCPUProfile(profile) != nil {
			profile.Close()
			profile = nil
			return result, errStudy
		}
		defer func() {
			if finishProfile() != nil {
				phase = "cpu_profile_finish"
				err = errStudy
			}
		}()
	}
	samples := make([]statehintclaims.Sample, len(train.rows))
	for i, r := range train.rows {
		samples[i] = statehintclaims.Sample{Text: r.Text, Targets: r.labels}
	}
	candidate := parent.Clone()
	var workspace statehintclaims.TrainingWorkspace
	beforeFit, e := inspectInput(root, p.evaluationName, inputBudget)
	if e != nil || !sameSnapshot(evaluationSnapshot, beforeFit) {
		return result, errStudy
	}
	phase = "fixed_warm_fit"
	fitCalls = 1
	fit, e := candidate.WarmFit(parent, samples, &workspace)
	if e != nil || fit.Samples != 1680 || fit.Epochs != 40 || fit.Batches != 2120 || fit.BaseTrainingSteps != 2120 || fit.NewOptimizerSteps != 2120 || fit.TrainingSteps != 4240 || fit.Seed != 1729 || fit.Initialization != fixedRecipe().Initialization || candidate.TrainingSteps() != 4240 || candidate.InitializationSeed() != 1729 {
		return result, errStudy
	}
	// Persist the one completed continuation before exposing fresh evaluation.
	// A later malformed/evaluation failure preserves this fit and cannot retry
	// through the same fresh output directory.
	var candidateBytes bytes.Buffer
	phase = "candidate_save_before_evaluation"
	if candidate.Save(&candidateBytes) != nil || candidateBytes.Len() != parentBytes || writeBytes(out, "candidate.rsc", candidateBytes.Bytes()) != nil {
		return result, errStudy
	}
	// This is the sole fit. Only after success may any fresh evaluation body be
	// read, or the fixed comparator be deserialized. Failure does not retry.
	phase = "evaluation_body_read_and_pin"
	data, e := readEvaluation(root, p, evaluationSnapshot)
	if e != nil {
		return result, errStudy
	}
	phase = "evaluation_schema_parse"
	evaluation, e := prepare(data, "dev")
	if e != nil {
		return result, errStudy
	}
	phase = "evaluation_reference_support_audit"
	if validateEvaluation(evaluation, train, s, receipt) != nil {
		return result, errStudy
	}
	phase = "comparator_read_and_pin"
	comparatorData, e := readPinned(root, p.comparatorName, p.comparatorSHA, parentBytes, parentBytes)
	if e != nil {
		return result, errStudy
	}
	phase = "comparator_numeric_load"
	comparator, e := canonicalModel(comparatorData, 4240)
	if e != nil {
		return result, errStudy
	}
	phase = "evaluation_predictions_and_aggregate_diagnostics"
	candidateObs, e := observe(candidate, evaluation)
	if e != nil {
		return result, errStudy
	}
	comparatorObs, e := observe(comparator, evaluation)
	if e != nil {
		return result, errStudy
	}
	candidateEval, e := evaluate(evaluation, candidateObs)
	if e != nil {
		return result, errStudy
	}
	comparatorEval, e := evaluate(evaluation, comparatorObs)
	if e != nil {
		return result, errStudy
	}
	delta, e := paired(evaluation, s, candidateObs, comparatorObs)
	if e != nil || delta.Lineages != [2]int{40, 20} {
		return result, errStudy
	}
	phase = "candidate_reload_prediction_score_byte_parity"
	saved, e := saveCandidate(out, candidate, evaluation, candidateObs)
	if e != nil {
		return result, errStudy
	}
	phase = "read_only_parent_comparator_parity"
	var unchanged bytes.Buffer
	if parent.Save(&unchanged) != nil || !bytes.Equal(parentData, unchanged.Bytes()) {
		return result, errStudy
	}
	unchanged.Reset()
	if comparator.Save(&unchanged) != nil || !bytes.Equal(comparatorData, unchanged.Bytes()) {
		return result, errStudy
	}
	phase = "cpu_profile_finish"
	if finishProfile() != nil {
		return result, errStudy
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	result = studyReport{Schema: "riido-three-claims-semantic-contrast-study-report-v1", Status: "ai_reference_research_only_not_human_or_product_evidence", TrainingSHA: p.trainingSHA, EvaluationSHA: p.evaluationSHA, StrataSHA: p.strataSHA, ReceiptSHA: p.receiptSHA, ReferenceStatus: "ai_authored_independently_ai_reviewed_human_pending", TrainingCounts: train.counts, EvaluationCounts: s.counts, States: statehintclaims.States(), Recipe: fixedRecipe(), Parent: modelSummary(parent, parentData), Comparator: modelSummary(comparator, comparatorData), Candidate: saved, Fit: fit, FitLossMeaning: "mean_online_fit_trajectory_loss_across_all_epochs_not_final_evaluation_ce", FitCalls: 1, CandidateEvaluation: candidateEval, ComparatorEvaluation: comparatorEval, Primary: delta, Progress: progress(delta, candidateEval, comparatorEval), FitBeforeEvaluation: true, ElapsedNS: time.Since(started).Nanoseconds(), GoHeap: memory.HeapAlloc, OSRSSMeasurement: "external_process_measurement_required_go_heap_is_not_os_rss", CPUProfile: cpuProfile}
	for i, h := range statehintclaims.Heads() {
		result.Heads[i] = h.String()
	}
	phase = "aggregate_report_write"
	encoded, e := json.MarshalIndent(result, "", "  ")
	if e != nil || writeBytes(out, "report.json", append(encoded, '\n')) != nil {
		return studyReport{}, errStudy
	}
	return result, nil
}
func run(args []string, output, errorOutput io.Writer) error {
	return runWithManifest(args, output, errorOutput, frozenManifest())
}

// Tests provide an owned synthetic manifest, never actual pinned model bytes.
func runWithManifest(args []string, output, errorOutput io.Writer, m manifest) error {
	f := flag.NewFlagSet("riido-statehint-claims-contrast-study", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	var p pins
	f.StringVar(&p.parentName, "parent", "", "original float parent")
	f.StringVar(&p.parentSHA, "parent-sha256", "", "parent exact SHA")
	f.StringVar(&p.comparatorName, "comparator", "", "fixed matched float")
	f.StringVar(&p.comparatorSHA, "comparator-sha256", "", "comparator exact SHA")
	f.StringVar(&p.trainingName, "training", "", "fixed training JSONL")
	f.StringVar(&p.trainingSHA, "training-sha256", "", "training exact SHA")
	f.StringVar(&p.evaluationName, "evaluation", "", "fresh sealed evaluation JSONL")
	f.StringVar(&p.evaluationSHA, "evaluation-sha256", "", "evaluation exact SHA")
	f.StringVar(&p.strataName, "strata", "", "text-free evaluation metadata")
	f.StringVar(&p.strataSHA, "strata-sha256", "", "metadata exact SHA")
	f.StringVar(&p.receiptName, "receipt", "", "sealed text-free writer receipt")
	f.StringVar(&p.receiptSHA, "receipt-sha256", "", "receipt exact SHA")
	check := f.Bool("check", false, "pins/schema/counts; evaluation body remains unopened")
	name := f.String("out", "", "fresh private study directory")
	profile := f.Bool("cpu-profile", false, "private CPU profile")
	if e := f.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errorOutput, "riido-statehint-claims-contrast-study: required --parent/--parent-sha256 --comparator/--comparator-sha256 --training/--training-sha256 --evaluation/--evaluation-sha256 --strata/--strata-sha256 --receipt/--receipt-sha256; --check OR --out .cache/statehint-claims-contrast-study/NEW [--cpu-profile]. Check uses a sealed text-free receipt and never opens evaluation, loads a model, predicts or fits. Actual mode performs exactly one fixed WarmFit before evaluation access or comparator Load. AI-reference research only.")
			return e
		}
		return errStudy
	}
	if f.NArg() != 0 || p.parentSHA != m.parent || p.comparatorSHA != m.comparator || *check && (*name != "" || *profile) || !*check && (!local(*name) || filepath.Dir(*name) != anchor) {
		return errStudy
	}
	for _, pair := range [][2]string{{p.parentName, p.parentSHA}, {p.comparatorName, p.comparatorSHA}, {p.trainingName, p.trainingSHA}, {p.evaluationName, p.evaluationSHA}, {p.strataName, p.strataSHA}, {p.receiptName, p.receiptSHA}} {
		if !local(pair[0]) || !validSHA(pair[1]) {
			return errStudy
		}
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errStudy
	}
	defer root.Close()
	evaluationSnapshot, e := preflightInputs(root, p)
	if e != nil {
		return e
	}
	parentData, e := readPinned(root, p.parentName, p.parentSHA, parentBytes, parentBytes)
	if e != nil || !modelShape(parentData, 2120) {
		return errStudy
	}
	trainingData, e := readPinned(root, p.trainingName, p.trainingSHA, inputBudget, 0)
	if e != nil {
		return errStudy
	}
	train, e := prepare(trainingData, "fit")
	if e != nil || !trainingScope(train.counts) {
		return errStudy
	}
	strataData, e := readPinned(root, p.strataName, p.strataSHA, metadataBudget, 0)
	if e != nil {
		return errStudy
	}
	s, e := prepareStrata(strataData, train)
	if e != nil {
		return errStudy
	}
	receiptData, e := readPinned(root, p.receiptName, p.receiptSHA, metadataBudget, 0)
	if e != nil {
		return errStudy
	}
	writerReceipt, e := validateReceipt(receiptData, p.trainingSHA, p.evaluationSHA, p.strataSHA, s)
	if e != nil {
		return errStudy
	}
	if evaluationSnapshot.Size() != writerReceipt.EvaluationBytes {
		return errStudy
	}
	if *check {
		comparatorData, e := readPinned(root, p.comparatorName, p.comparatorSHA, parentBytes, parentBytes)
		if e != nil || !modelShape(comparatorData, 4240) {
			return errStudy
		}
		return json.NewEncoder(output).Encode(struct {
			Status     string `json:"status"`
			Training   counts `json:"training_counts"`
			Evaluation counts `json:"evaluation_receipt_counts"`
		}{"checked pinned training/metadata/receipt and model containers; no Load, Predict, Fit, evaluation body access or outputs", train.counts, s.counts})
	}
	parent, e := canonicalModel(parentData, 2120)
	if e != nil {
		return errStudy
	}
	private, e := privateOutput(root, *name)
	if e != nil {
		return errStudy
	}
	defer private.Close()
	r, e := fitStudy(root, private, parent, parentData, train, s, writerReceipt, p, evaluationSnapshot, *profile)
	if e != nil {
		return errStudy
	}
	return json.NewEncoder(output).Encode(struct {
		Status    string `json:"status"`
		Qualified bool   `json:"semantic_quality_qualified"`
		Progress  bool   `json:"research_progress"`
	}{r.Status, false, r.Progress.ResearchProgress})
}
func main() {
	runtime.GOMAXPROCS(2)
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "semantic contrast study failed; check pinned canonical inputs and fresh private output")
		os.Exit(1)
	}
}
