// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"math"
	"strings"
	"testing"
)

// Authored controls are entirely synthetic. No public request, JWT token,
// fixture, Want, observed outcome, model or native verifier is used.
func retrievalSyntheticCaptions() [8]string {
	return [8]string{"green blue", "red green", "red", "purple", "purple", "purple", "purple", "purple"}
}
func TestRetrievalJaccardKnownSetArithmetic(t *testing.T) {
	captions := retrievalSyntheticCaptions()
	got, code := rankRetrievalText("red red green", captions, retrievalJaccard)
	if code != retrievalOK {
		t.Fatalf("qualification: %s", code)
	}
	if got.Order != [8]int{1, 2, 0, 3, 4, 5, 6, 7} || got.Scores[0] != 1.0/3.0 || got.Scores[1] != 1 || got.Scores[2] != 0.5 {
		t.Fatal("unique set intersection/union or stable zero ties changed")
	}
	for i := 3; i < 8; i++ {
		if got.Scores[i] != 0 {
			t.Fatal("absent words earned score")
		}
	}
	repeated := captions
	repeated[1] = "red green red green"
	same, code := rankRetrievalText("red green green", repeated, retrievalJaccard)
	if code != retrievalOK || same.Order != got.Order || same.Scores != got.Scores {
		t.Fatal("set score depended on multiplicity")
	}
}
func TestRetrievalBM25AnalyticSparseHitAndQueryMultiplicity(t *testing.T) {
	captions := [8]string{"needle", "other", "other", "other", "other", "other", "other", "other"}
	got, code := rankRetrievalText("needle", captions, retrievalBM25)
	// Eight unit-length documents and df1 reduce the single hit to log(6).
	if code != retrievalOK || got.Order != [8]int{0, 1, 2, 3, 4, 5, 6, 7} || math.Abs(got.Scores[0]-math.Log(6)) > 1e-12 {
		t.Fatal("analytic sparse-hit reference changed")
	}
	for i := 1; i < 8; i++ {
		if got.Scores[i] != 0 {
			t.Fatal("nonmatching document earned BM25 score")
		}
	}
	duplicate, code := rankRetrievalText("needle needle", captions, retrievalBM25)
	if code != retrievalOK || duplicate.Order != got.Order || duplicate.Scores != got.Scores {
		t.Fatal("query repetition counted twice")
	}
}
func TestRetrievalStableTiesAndCallerOwnership(t *testing.T) {
	var captions [8]string
	for i := range captions {
		captions[i] = "same words"
	}
	original := captions
	for _, kind := range [2]retrievalKind{retrievalBM25, retrievalJaccard} {
		got, code := rankRetrievalText("same words", captions, kind)
		if code != retrievalOK || got.Order != [8]int{0, 1, 2, 3, 4, 5, 6, 7} {
			t.Fatal("equal-score ties lost source order")
		}
		for _, v := range got.Scores {
			if v != got.Scores[0] {
				t.Fatal("identical caption scores differ")
			}
		}
		got.Order[0] = 7
		got.Scores[0] = -1
		next, code := rankRetrievalText("same words", captions, kind)
		if code != retrievalOK || next.Order[0] != 0 || next.Scores[0] < 0 || captions != original {
			t.Fatal("caller value mutation leaked across calls")
		}
	}
}
func TestRetrievalBM25DocumentFrequencyAndLengthReference(t *testing.T) {
	captions := [8]string{"needle needle", "needle", "other", "other", "other", "other", "other", "other"}
	got, code := rankRetrievalText("needle", captions, retrievalBM25)
	// df2 and nine tokens across eight docs give idf=log3.6, denominator3.9
	// for the double hit and2.1 for the single hit; these are independent closed
	// rational references, not saved task outcomes or regenerated Wants.
	idf := math.Log(3.6)
	if code != retrievalOK || got.Order != [8]int{0, 1, 2, 3, 4, 5, 6, 7} || math.Abs(got.Scores[0]-idf*4.4/3.9) > 1e-12 || math.Abs(got.Scores[1]-idf*2.2/2.1) > 1e-12 {
		t.Fatal("document TF/length/df arithmetic changed")
	}
}
func TestRetrievalNormalizationAndHardBounds(t *testing.T) {
	if retrievalNormalize("XMLParser2.foo BAR_baz") != "xml parser2 foo bar baz" {
		t.Fatal("CamelCase/full punctuation normalization changed")
	}
	if retrievalNormalize("École 42_Δelta") != "école 42 δelta" {
		t.Fatal("Unicode letter/digit normalization changed")
	}
	w, code := retrievalPrepare(strings.Repeat("x", 512))
	if code != retrievalOK || w.Count != 1 {
		t.Fatal("exact raw and normalized byte boundary rejected")
	}
	if _, code := retrievalPrepare(strings.Repeat("x", 513)); code != retrievalUnknownText {
		t.Fatal("raw byte limit ignored")
	}
	if _, code := retrievalPrepare(string([]byte{0xff})); code != retrievalUnknownUnicode {
		t.Fatal("invalid UTF8 repaired")
	}
	if _, code := retrievalPrepare("!!!"); code != retrievalUnknownNormalized {
		t.Fatal("empty normalized request fabricated")
	}
	if _, code := retrievalPrepare(strings.TrimSpace(strings.Repeat("x ", 32))); code != retrievalOK {
		t.Fatal("32-word boundary rejected")
	}
	if _, code := retrievalPrepare(strings.TrimSpace(strings.Repeat("x ", 33))); code != retrievalUnknownNormalized {
		t.Fatal("33-word request truncated")
	}
	if _, code := retrievalPrepare(strings.Repeat("Ⱥ", 256)); code != retrievalUnknownNormalized {
		t.Fatal("Unicode lowercasing expansion exceeded byte budget silently")
	}
	captions := retrievalSyntheticCaptions()
	captions[7] = ""
	out, code := rankRetrievalText("red green", captions, retrievalBM25)
	if code != retrievalUnknownText || out.Order != [8]int{0, 1, 2, 3, 4, 5, 6, 7} {
		t.Fatal("missing eighth caption became successful partial ranking")
	}
	if _, code := rankRetrievalText("red green", retrievalSyntheticCaptions(), retrievalKind(255)); code != retrievalUnknownKind {
		t.Fatal("unsupported method fell back silently")
	}
}
