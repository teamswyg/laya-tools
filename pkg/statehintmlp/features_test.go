// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintmlp

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

// Owned v2 zero/probe artifacts expose each contextual bin through
// log(P[class]/P[zero-logit-class]). Casting back to float32 gives bit parity
// with the actual public wide extractor, without changing its private API.
func TestEveryContextualBinExactlyMatchesPublicWideProbes(t *testing.T) {
	texts := []string{"Original alpha completed beta remains", "beta remains alpha completed", "원본 가상 단위는 마쳤고 다른 단계는 남았습니다", "İstanbul CAFÉ e\u0301 １２3 가나🙂 A", "alpha alpha beta alpha !!! A", strings.Repeat("g", MaxTextBytes)}
	var template bytes.Buffer
	if e := statehintwide.NewModel(statehintwide.Contextual).Save(&template); e != nil {
		t.Fatal(e)
	}
	base := template.Bytes()
	header := int(binary.LittleEndian.Uint16(base[6:8]))
	if header != 128 || statehintwide.FeatureBins != FeatureBins {
		t.Fatal("v2 public artifact layout changed; update independent feature probe")
	}
	want := make([][FeatureBins]float32, len(texts))
	var w Workspace
	for i, text := range texts {
		if e := extract(text, &w); e != nil {
			t.Fatal(e)
		}
		want[i] = w.values
	}
	for start := 0; start < FeatureBins; start += 7 {
		data := append([]byte(nil), base...)
		for c := 0; c < 7 && start+c < FeatureBins; c++ {
			offset := header + ((start+c)*IntentCount+c)*4
			binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(1))
		}
		sum := sha256.Sum256(data[:len(data)-32])
		copy(data[len(data)-32:], sum[:])
		probe, e := statehintwide.Load(bytes.NewReader(data))
		if e != nil {
			t.Fatal(e)
		}
		var wideWorkspace statehintwide.Workspace
		for i, text := range texts {
			p, e := probe.Predict(text, &wideWorkspace)
			if e != nil {
				t.Fatal(e)
			}
			for c := 0; c < 7 && start+c < FeatureBins; c++ {
				got := float32(math.Log(p.Probabilities[c] / p.Probabilities[7]))
				if math.Float32bits(got) != math.Float32bits(want[i][start+c]) {
					t.Fatalf("feature mismatch text%d bin%d: %.9g != %.9g", i, start+c, got, want[i][start+c])
				}
			}
		}
	}
}
func TestWorkspaceReuseOrderAndInputBounds(t *testing.T) {
	var a, b Workspace
	if extract("first second third", &a) != nil || extract("third second first", &b) != nil {
		t.Fatal("fixture")
	}
	if a.values == b.values {
		t.Fatal("word bigram order lost")
	}
	if extract("새로운 original input", &a) != nil || extract("새로운 original input", &b) != nil || a.values != b.values || a.count != b.count || !slices.Equal(a.indices[:a.count], b.indices[:b.count]) {
		t.Fatal("workspace retained previous features")
	}
	for _, text := range []string{strings.Repeat("x", MaxTextBytes+1), "zero\x00byte", string([]byte{0xff})} {
		if extract(text, &a) == nil {
			t.Fatal("invalid/oversized UTF8 accepted")
		}
	}
	if extract("valid", nil) == nil {
		t.Fatal("nil workspace accepted")
	}
	if extract(strings.Repeat("x", MaxTextBytes), &a) != nil {
		t.Fatal("exact byte bound rejected")
	}
}
