// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 teamswyg contributors.
package hintlearn

import (
	"math"
	"reflect"
	"slices"
	"sort"
)

// FeatureScratchError is a fixed internal diagnostic without supplied text.
type FeatureScratchError string

func (e FeatureScratchError) Error() string { return string(e) }

const (
	ErrFeatureScratchSize     FeatureScratchError = "feature_scratch_size"
	ErrFeatureScratchCapacity FeatureScratchError = "feature_scratch_capacity"
)

// FeatureScratch owns one mutable cross-feature buffer. It is constructor-local,
// not a concurrent cache. Copies share the mutable buffer. Features returns a
// borrowed compact row: copy it before another successful use of this scratch.
// The independent Features implementation remains the feature-bit oracle.
type FeatureScratch struct{ cross []Feature }

func validScratchSize(slots int) bool {
	// Size() observes the Go layout without unsafe or a platform-size assumption.
	maxInt := int(^uint(0) >> 1)
	return slots >= 0 && slots <= maxInt/int(reflect.TypeOf(Feature{}).Size())
}
func featureCrossSlots(qt, dt []string) (int, error) {
	if len(qt) == 0 || len(dt) == 0 {
		return 0, nil
	}
	maxInt := int(^uint(0) >> 1)
	if len(qt) > (maxInt-1)/len(dt) {
		return 0, ErrFeatureScratchSize
	}
	slots := len(qt)*len(dt) + 1
	if !validScratchSize(slots) {
		return 0, ErrFeatureScratchSize
	}
	return slots, nil
}

// FeatureCardinality counts distinct cross indices, including features whose
// signed values will cancel to zero. Index0 is the final overlap feature, not
// a constant bias. crossSlots reserves the full unsorted cross product plus it.
// This prepass repeats tokenization/hash work; callers must charge that cost.
// Callers enforce text bounds before using these internal helpers.
func FeatureCardinality(query, document string) (count, crossSlots int, err error) {
	qw, dw := words(query), words(document)
	qt, dt := terms(qw), terms(dw)
	crossSlots, err = featureCrossSlots(qt, dt)
	if err != nil || crossSlots == 0 {
		return 0, crossSlots, err
	}
	var seen [Dimension / 64]uint64
	for _, q := range qt {
		prefix := hashFrom(hash(q), "\x00")
		for _, d := range dt {
			index := 1 + int(hashFrom(prefix, d)%(Dimension-1))
			word, mask := index/64, uint64(1)<<uint(index%64)
			if seen[word]&mask == 0 {
				seen[word] |= mask
				count++
			}
		}
	}
	return count + 1, crossSlots, nil
}

// NewFeatureScratch allocates exactly one reserved buffer. Zero capacity is
// valid for empty-term rows; negative or allocation-overflow sizes are rejected.
func NewFeatureScratch(crossSlots int) (*FeatureScratch, error) {
	if !validScratchSize(crossSlots) {
		return nil, ErrFeatureScratchSize
	}
	return &FeatureScratch{cross: make([]Feature, crossSlots)}, nil
}

// Features retains the original generation, unstable sort.Slice comparator,
// duplicate accumulation and final overlap order. Capacity affects storage only.
// Insufficient capacity fails before changing the previous borrowed row.
func (s *FeatureScratch) Features(query, document string) ([]Feature, error) {
	if s == nil {
		return nil, ErrFeatureScratchSize
	}
	qw, dw := words(query), words(document)
	qt, dt := terms(qw), terms(dw)
	slots, err := featureCrossSlots(qt, dt)
	if err != nil {
		return nil, err
	}
	if slots == 0 {
		return nil, nil
	}
	if slots > cap(s.cross) {
		return nil, ErrFeatureScratchCapacity
	}
	fs := s.cross[:0]
	scale := 1 / math.Sqrt(float64(len(qt)*len(dt)))
	for _, q := range qt {
		prefix := hashFrom(hash(q), "\x00")
		for _, d := range dt {
			h := hashFrom(prefix, d)
			v := scale
			if h>>63 != 0 {
				v = -v
			}
			fs = append(fs, Feature{1 + int(h%(Dimension-1)), v})
		}
	}
	sort.Slice(fs, func(i, j int) bool { return fs[i].Index < fs[j].Index })
	n := 0
	for _, f := range fs {
		if n > 0 && fs[n-1].Index == f.Index {
			fs[n-1].Value += f.Value
		} else {
			fs[n] = f
			n++
		}
	}
	fs = fs[:n]
	sort.Strings(qw)
	qw = slices.Compact(qw)
	sort.Strings(dw)
	dw = slices.Compact(dw)
	hits := 0
	for _, q := range qw {
		if slices.Contains(dw, q) {
			hits++
		}
	}
	fs = append(fs, Feature{0, float64(hits) / float64(max(1, len(qw)))})
	return fs, nil
}
