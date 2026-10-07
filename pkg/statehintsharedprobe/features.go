// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintsharedprobe trains a shared linear head on frozen Laya features.
// It cannot infer from text without the original feature-producing backbone.
package statehintsharedprobe

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

const (
	IntentCount       = 8
	FeatureDimensions = 1024
	RowBytes          = IntentCount * FeatureDimensions * 4
	MaxRows           = 2400
	FeatureSchema     = "laya-v1-prefinal-8x1024-fp32le-shared-no-bias-v1"
	BaseSHA           = "891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c"
	InstructionSHA    = "255725ec00a4526fefedfad729ddd75acc9681181b81fc0f8f96b89b0a400fd3"
)

var (
	ErrFeatures = errors.New("frozen feature row invalid or outside bounds")
	ErrModel    = errors.New("shared feature head invalid")
	ErrTraining = errors.New("shared feature training invalid")
	ErrArtifact = errors.New("shared feature head artifact invalid")
)

type Row [IntentCount][FeatureDimensions]float32
type Workspace struct {
	Row    Row
	bytes  [RowBytes]byte
	logits [IntentCount]float64
}

// Store borrows a ReaderAt. The caller owns its lifetime, immutable bytes and
// provenance checks. No full-file allocation or process-global cache is made.
type Store struct {
	reader io.ReaderAt
	rows   int
}

func NewStore(reader io.ReaderAt, rows int) (*Store, error) {
	if reader == nil || rows < 1 || rows > MaxRows {
		return nil, ErrFeatures
	}
	return &Store{reader, rows}, nil
}
func (s *Store) Rows() int {
	if s == nil {
		return 0
	}
	return s.rows
}
func (s *Store) ReadRow(index int, w *Workspace) error {
	if s == nil || w == nil || index < 0 || index >= s.rows {
		return ErrFeatures
	}
	n, e := s.reader.ReadAt(w.bytes[:], int64(index)*RowBytes)
	if n != RowBytes || e != nil && e != io.EOF {
		return ErrFeatures
	}
	offset := 0
	for option := range w.Row {
		for feature := range w.Row[option] {
			v := math.Float32frombits(binary.LittleEndian.Uint32(w.bytes[offset : offset+4]))
			if !finite(float64(v)) {
				return ErrFeatures
			}
			w.Row[option][feature] = v
			offset += 4
		}
	}
	return nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func validRow(r *Row) bool {
	if r == nil {
		return false
	}
	for _, v := range r {
		for _, x := range v {
			if !finite(float64(x)) {
				return false
			}
		}
	}
	return true
}
