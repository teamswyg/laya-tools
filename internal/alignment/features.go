// Package alignment combines cheap lexical evidence with frozen static vectors.
// Each versioned variant has its own coefficient contract; these are not hbin
// coefficients and do not change the existing optional hint model.
package alignment

import (
	"fmt"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/staticembed"
	"math"
	"unicode/utf8"
)

const Schema = "riido-alignment-features-v1"

func Dimension(variant string) int {
	switch variant {
	case "lexical":
		return 16
	case "normalized_cosine":
		return 17
	case "normalized_alignment":
		return 145
	}
	return 0
}
func Features(model *staticembed.Model, variant, query, code string) ([]float64, error) {
	n := Dimension(variant)
	if n == 0 {
		return nil, fmt.Errorf("invalid feature variant")
	}
	if !utf8.ValidString(query) || !utf8.ValidString(code) || !paireval.InScope(paireval.Row{Query: query, Code: code}) {
		return nil, fmt.Errorf("input outside source scope")
	}
	x := make([]float64, n)
	lex := lexicalhint.Features(query, code, false)
	copy(x, lex[:])
	if variant == "lexical" {
		return x, nil
	}
	if model == nil {
		return nil, fmt.Errorf("embedding model required")
	}
	q, _, e := model.Encode(lexicalhint.NormalizeText(query))
	if e != nil {
		return nil, e
	}
	c, _, e := model.Encode(lexicalhint.NormalizeText(code))
	if e != nil {
		return nil, e
	}
	x[16] = staticembed.Cosine(q, c)
	if variant == "normalized_alignment" {
		for i := 0; i < 64; i++ {
			x[17+i] = 8 * float64(q[i]) * float64(c[i])
			x[81+i] = math.Abs(float64(q[i]) - float64(c[i]))
		}
	}
	return x, nil
}
func Prepare(rows []paireval.Row, s paireval.Split, name, variant string, m *staticembed.Model) (pairlearn.Dataset, error) {
	d := pairlearn.Dataset{Split: name, Offsets: []int{0}}
	if name != "development" && name != "validation" {
		return d, fmt.Errorf("holdout preparation prohibited")
	}
	if len(rows) != len(s.Rows) || Dimension(variant) == 0 {
		return d, fmt.Errorf("invalid partition or variant")
	}
	for i, r := range rows {
		if s.Rows[i].Split != name {
			continue
		}
		if r.Label == nil || (*r.Label != 0 && *r.Label != 1) {
			return d, fmt.Errorf("invalid label")
		}
		if !paireval.InScope(r) {
			d.Excluded++
			continue
		}
		x, e := Features(m, variant, r.Query, r.Code)
		if e != nil {
			return d, e
		}
		for j, v := range x {
			if v != 0 {
				d.Indices = append(d.Indices, uint16(j))
				d.Values = append(d.Values, v)
			}
		}
		d.Offsets = append(d.Offsets, len(d.Values))
		d.Labels = append(d.Labels, float64(*r.Label))
		d.Groups = append(d.Groups, s.Rows[i].Group)
	}
	return d, nil
}
