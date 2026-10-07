// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimsmlp

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"testing"
)

func saved(t *testing.T, m *Model) []byte {
	t.Helper()
	var data bytes.Buffer
	if err := m.Save(&data); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func checksum(data []byte) {
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	copy(data[end:], sum[:])
}

func TestUntrainedAndSyntheticTrainedArtifactsExactlyRoundtrip(t *testing.T) {
	for _, trained := range []bool{false, true} {
		m := NewModel()
		if trained {
			var training TrainingWorkspace
			if _, err := m.Fit(ownedSamples(), &training); err != nil {
				t.Fatal(err)
			}
		}
		data := saved(t, m)
		if len(data) != 131972 || string(data[:4]) != "RCM\x00" || binary.LittleEndian.Uint16(data[6:8]) != 192 {
			t.Fatal("distinct fixed format bounds")
		}
		loaded, err := Load(bytes.NewReader(data))
		if err != nil || *loaded != *m || !bytes.Equal(data, saved(t, loaded)) || loaded.Metadata() != m.Metadata() {
			t.Fatal("full parameters/history/metadata/bytes reload drift", err)
		}
		for _, text := range []string{"Original owned synthetic score fixture.", "직접 만든 합성 추론 자료입니다.", "... 🔧", ""} {
			var a, b Workspace
			pa, ea := m.Predict(text, &a)
			pb, eb := loaded.Predict(text, &b)
			sa, esa := m.Scores(text, &a)
			sb, esb := loaded.Scores(text, &b)
			if ea != nil || eb != nil || esa != nil || esb != nil || pa != pb || sa != sb {
				t.Fatal("full prediction/score reload parity", text)
			}
		}
		if !trained {
			const initialPin = "867e0ebe0805cf62fa69ba3c6a3def9fa2b429fa118ed8146fb5ccbd222b521c"
			got := fmt.Sprintf("%x", sha256.Sum256(data))
			if got != initialPin {
				t.Fatal("deterministic initialization artifact drift", got)
			}
		}
	}
}

func TestArtifactsRejectMeaningfulContractMutationsDespiteValidDigest(t *testing.T) {
	original := saved(t, NewModel())
	for _, tc := range []struct {
		name   string
		change func([]byte)
	}{
		{"linear-parent-RSC", func(b []byte) { copy(b[:4], "RSC\x00") }},
		{"eight-way-parent-RSM", func(b []byte) { copy(b[:4], "RSM\x00") }},
		{"ternary-RQT", func(b []byte) { copy(b[:4], "RQT\x00") }},
		{"version", func(b []byte) { b[4]++ }},
		{"header-size", func(b []byte) { binary.LittleEndian.PutUint16(b[6:8], 65535) }},
		{"input-size", func(b []byte) { binary.LittleEndian.PutUint16(b[8:10], 65535) }},
		{"hidden-size", func(b []byte) { binary.LittleEndian.PutUint16(b[10:12], 65535) }},
		{"head-size", func(b []byte) { b[12] = 255 }},
		{"state-size", func(b []byte) { b[13] = 255 }},
		{"activation", func(b []byte) { b[14]++ }},
		{"initialization", func(b []byte) { b[15]++ }},
		{"seed", func(b []byte) { binary.LittleEndian.PutUint64(b[24:32], 1) }},
		{"feature-schema-hash", func(b []byte) { b[32] ^= 1 }},
		{"head-state-order-hash", func(b []byte) { b[64] ^= 1 }},
		{"numeric-contract-hash", func(b []byte) { b[96] ^= 1 }},
		{"confidence", func(b []byte) { binary.LittleEndian.PutUint64(b[128:136], math.Float64bits(.8)) }},
		{"margin", func(b []byte) { binary.LittleEndian.PutUint64(b[136:144], math.Float64bits(.04)) }},
		{"temperature", func(b []byte) { binary.LittleEndian.PutUint64(b[144:152], math.Float64bits(.65)) }},
		{"untrained-with-samples", func(b []byte) { binary.LittleEndian.PutUint32(b[152:156], 1) }},
		{"trained-without-samples", func(b []byte) { binary.LittleEndian.PutUint64(b[16:24], 40) }},
		{"sample-overflow", func(b []byte) { binary.LittleEndian.PutUint32(b[152:156], math.MaxUint32) }},
		{"step-overflow", func(b []byte) { binary.LittleEndian.PutUint64(b[16:24], math.MaxUint64) }},
		{"inconsistent-fixed-steps", func(b []byte) {
			binary.LittleEndian.PutUint32(b[152:156], 1680)
			binary.LittleEndian.PutUint64(b[16:24], 2080)
		}},
		{"parameter-count", func(b []byte) { binary.LittleEndian.PutUint32(b[156:160], math.MaxUint32) }},
		{"weight-count", func(b []byte) { binary.LittleEndian.PutUint32(b[160:164], math.MaxUint32) }},
		{"reserved-start", func(b []byte) { b[164] = 1 }},
		{"reserved-end", func(b []byte) { b[191] = 1 }},
		{"nan-input", func(b []byte) { binary.LittleEndian.PutUint32(b[192:196], math.Float32bits(float32(math.NaN()))) }},
		{"inf-hidden-bias", func(b []byte) {
			offset := 192 + FeatureBins*HiddenUnits*4
			binary.LittleEndian.PutUint32(b[offset:offset+4], math.Float32bits(float32(math.Inf(1))))
		}},
		{"inf-output", func(b []byte) {
			offset := 192 + (FeatureBins*HiddenUnits+HiddenUnits)*4
			binary.LittleEndian.PutUint32(b[offset:offset+4], math.Float32bits(float32(math.Inf(-1))))
		}},
		{"nan-output-bias", func(b []byte) {
			offset := len(b) - sha256.Size - 4
			binary.LittleEndian.PutUint32(b[offset:offset+4], math.Float32bits(float32(math.NaN())))
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := append([]byte(nil), original...)
			tc.change(bad)
			checksum(bad) // Exercise contract parsing, not merely digest mismatch.
			if _, err := Load(bytes.NewReader(bad)); err != ErrArtifact {
				t.Fatal("unknown/malformed numerical contract accepted", err)
			}
		})
	}
}

type shortWriter struct{}

func (shortWriter) Write(data []byte) (int, error) { return len(data) - 1, nil }

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestArtifactBoundsChecksumsNilPartialIOAndInvalidSave(t *testing.T) {
	m := NewModel()
	original := saved(t, m)
	corrupt := append([]byte(nil), original...)
	corrupt[192] ^= 1
	for _, bad := range [][]byte{nil, original[:191], original[:len(original)-1], append(append([]byte(nil), original...), 0), corrupt} {
		if _, err := Load(bytes.NewReader(bad)); err != ErrArtifact {
			t.Fatal("bad size or checksum accepted", len(bad), err)
		}
	}
	reader := bytes.NewReader(append(append([]byte(nil), original...), bytes.Repeat([]byte{1}, 100)...))
	if _, err := Load(reader); err != ErrArtifact || reader.Len() != 99 {
		t.Fatal("bounded parser overconsumed trailing data", err)
	}
	if _, err := Load(nil); err != ErrArtifact {
		t.Fatal("nil reader accepted", err)
	}
	if m.Save(shortWriter{}) != io.ErrShortWrite || m.Save(nil) != ErrModel {
		t.Fatal("short/nil writer accepted")
	}
	marker := errors.New("synthetic writer failure")
	if m.Save(failingWriter{marker}) != marker {
		t.Fatal("writer error lost")
	}
	for _, block := range []*float32{&m.input[0][0], &m.hiddenBias[0], &m.output[0][0][0], &m.outputBias[0][0]} {
		prior := *block
		*block = float32(math.Inf(1))
		var writer bytes.Buffer
		if m.Save(&writer) != ErrModel || writer.Len() != 0 {
			t.Fatal("invalid model wrote partial bytes")
		}
		*block = prior
	}
	m.steps = 1
	var writer bytes.Buffer
	if m.Save(&writer) != ErrModel || writer.Len() != 0 {
		t.Fatal("untrained inconsistent history was serialized")
	}
}
