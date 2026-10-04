// SPDX-License-Identifier: Apache-2.0
// Package hintprepared provides an explicitly selected, unverified ordering hint.
// It prepares existing hintlearn features once; it changes no router defaults,
// activates no model, certifies no claim and authorizes no action.
//
// The internal/hintlearn dependency includes training code at package/build
// level. This API calls only Features and exposes no Fit/Encode/Decode operation.
// A View's wire format/dimension does not prove feature-schema compatibility,
// quality, source clearance or license qualification; callers must establish
// those separately. Scores are not probabilities.
package hintprepared

import (
	"errors"
	"math"
	"strings"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/pkg/hintweights"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const Kind = "prepared_hint_v1"

var (
	ErrInput    = errors.New("hintprepared_invalid_input")
	ErrPrepared = errors.New("hintprepared_invalid_owner")
	ErrMismatch = errors.New("hintprepared_input_mismatch")
	ErrView     = errors.New("hintprepared_invalid_view")
	ErrFeature  = errors.New("hintprepared_invalid_feature")
	ErrScore    = errors.New("hintprepared_nonfinite_score")
)

// Prepared owns immutable cloned raw text and a tight flat feature array.
// It binds exact request/count/ordered text, independent of IDs and provenance.
// Copies share immutable storage. Rank has no shared scratch or score cache;
// concurrent read-only use with an immutable View returns independent values.
// The zero value is invalid. Features and underlying slices are not exposed.
type Prepared struct {
	request  string
	texts    [shortclaim.MaxCandidates]string
	offsets  [shortclaim.MaxCandidates + 1]uint32
	features []hintweights.Feature
	count    int
	ready    bool
}

// Prepare first revalidates every exported Prepared field, including normalized
// forms, provenance, ASCII identifiers and unused slots, preserving its fixed
// public diagnostics. Features then uses raw
// text with its existing tokenization, hash, duplicate accumulation and order.
// There is no caller budget. Validated byte/count bounds and <=8192 features per
// row bound retained data; temporary construction is not a peak-memory limit.
// Logical retained data is tight feature backing + cloned text byte lengths;
// owner headers/arrays/padding are separate. These are not allocator, heap or RSS
// measurements, and exclude caller inputs, model/View and returned value objects.
func Prepare(p shortclaim.Prepared) (*Prepared, error) {
	if err := shortclaim.ValidatePrepared(p); err != nil {
		return nil, err
	}
	var rows [shortclaim.MaxCandidates][]hintlearn.Feature
	var offsets [shortclaim.MaxCandidates + 1]uint32
	total := 0
	for i := 0; i < p.Count; i++ {
		fs := hintlearn.Features(p.Request, p.Candidates[i].Text)
		if len(fs) > hintweights.Dimension {
			return nil, ErrFeature
		}
		for _, f := range fs {
			if f.Index < 0 || f.Index >= hintweights.Dimension || math.IsNaN(f.Value) || math.IsInf(f.Value, 0) {
				return nil, ErrFeature
			}
		}
		rows[i] = fs
		total += len(fs)
		offsets[i+1] = uint32(total) // 8*8192 can be 65536; never uint16 offsets.
	}
	out := &Prepared{request: strings.Clone(p.Request), offsets: offsets, count: p.Count, features: make([]hintweights.Feature, total)}
	for i := 0; i < p.Count; i++ {
		out.texts[i] = strings.Clone(p.Candidates[i].Text)
		for j, f := range rows[i] {
			out.features[int(offsets[i])+j] = hintweights.Feature{Index: f.Index, Value: f.Value}
		}
	}
	out.ready = true
	return out, nil
}

// Rank retains every current candidate. Returned indices refer to current,
// so ID/provenance-only changes require no feature rebuilding. A changed raw
// request/count/ordered text returns ErrMismatch, without rebuild or fallback.
// A new View always recomputes all scores. Ties preserve current input positions.
// Constructor-validated opaque input avoids repeating public-field validation.
// Any error returns the zero Ranking. No caller buffer or persistent state is
// mutated; returned fixed arrays can be retained or modified independently.
func (p *Prepared) Rank(current shortclaim.ValidatedInput, view *hintweights.View) (shortclaim.Ranking, error) {
	if p == nil || !p.ready || p.count < 1 || p.count > shortclaim.MaxCandidates {
		return shortclaim.Ranking{}, ErrPrepared
	}
	now := current.Prepared()
	if now.Count < 1 || now.Count > shortclaim.MaxCandidates {
		return shortclaim.Ranking{}, ErrInput
	}
	if now.Request != p.request || now.Count != p.count {
		return shortclaim.Ranking{}, ErrMismatch
	}
	for i := 0; i < p.count; i++ {
		if now.Candidates[i].Text != p.texts[i] {
			return shortclaim.Ranking{}, ErrMismatch
		}
	}
	if _, err := view.Coefficient(0); err != nil {
		return shortclaim.Ranking{}, ErrView
	}
	out := shortclaim.Ranking{Kind: Kind, Count: p.count}
	for i := 0; i < p.count; i++ {
		s, err := view.Score(p.features[p.offsets[i]:p.offsets[i+1]])
		if err != nil {
			return shortclaim.Ranking{}, ErrFeature
		}
		if math.IsNaN(s) || math.IsInf(s, 0) {
			return shortclaim.Ranking{}, ErrScore
		}
		out.Scores[i], out.Order[i] = s, i
	}
	for i := 1; i < p.count; i++ {
		for j := i; j > 0; j-- {
			a, b := out.Order[j-1], out.Order[j]
			if out.Scores[a] > out.Scores[b] || (out.Scores[a] == out.Scores[b] && a < b) {
				break
			}
			out.Order[j-1], out.Order[j] = b, a
		}
	}
	return out, nil
}
