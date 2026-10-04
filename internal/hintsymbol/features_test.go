// SPDX-License-Identifier: Apache-2.0
package hintsymbol

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
)

// Owned synthetic text only: these controls call feature extraction, never a
// corpus, source behavior API, model, scorer, trainer or benchmark.
func symbolBitsEqual(a, b []Feature) bool {
	if len(a) != len(b) || (a == nil) != (b == nil) {
		return false
	}
	for i := range a {
		if a[i].Index != b[i].Index || math.Float64bits(a[i].Value) != math.Float64bits(b[i].Value) {
			return false
		}
	}
	return true
}

func legacyBitsEqual(a []Feature, b []hintlearn.Feature) bool {
	if len(a) != len(b) || (a == nil) != (b == nil) {
		return false
	}
	for i := range a {
		if a[i].Index != b[i].Index || math.Float64bits(a[i].Value) != math.Float64bits(b[i].Value) {
			return false
		}
	}
	return true
}

func TestNoSymbolLegacyBits(t *testing.T) {
	var distinct [MaxWords]string
	for i := range distinct {
		distinct[i] = fmt.Sprintf("term%d", i)
	}
	distinctWords := strings.Join(distinct[:], " ")
	fixtures := []struct{ name, query, document string }{
		{"empty-query", "", "alpha"},
		{"empty-document", "alpha", ""},
		{"punctuation-only", "!!!🙂", "alpha"},
		{"ordered-words", "retain green then remove amber", "remove amber then retain green"},
		{"case-and-unicode", "Å β 한글 １２", "å Β 한글 １２"},
		{"separators", "left\x00right, high_low; short-long", "right left high low short long"},
		{"repeated-words", strings.Repeat("repeat ", MaxWords), strings.Repeat("repeat ", MaxWords)},
		{"distinct-word-boundary", distinctWords, distinctWords},
		{"byte-boundary", strings.Repeat("a", MaxBytes), strings.Repeat("é", MaxBytes/2)},
	}
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			got, err := Extract(f.query, f.document)
			if err != nil || !legacyBitsEqual(got, hintlearn.Features(f.query, f.document)) {
				t.Fatalf("legacy indices, value bits, order or nilness changed: %v", err)
			}
		})
	}
}

func TestComparisonAndIntervalInformationLoss(t *testing.T) {
	fixtures := []struct{ name, a, b string }{
		{"less-boundary", "value < ceiling", "value <= ceiling"},
		{"greater-boundary", "value > floor", "value >= floor"},
		{"direction", "value < threshold", "value > threshold"},
		{"closed-open", "[lower upper]", "(lower upper)"},
		{"left-right-open", "[lower upper)", "(lower upper]"},
	}
	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			// The old representation's equality is the information-loss control,
			// not evidence that any new feature knows which behavior is correct.
			oldA, oldB := hintlearn.Features("inspect boundary", f.a), hintlearn.Features("inspect boundary", f.b)
			if len(oldA) != len(oldB) {
				t.Fatal("fixture no longer demonstrates legacy information loss")
			}
			for i := range oldA {
				if oldA[i].Index != oldB[i].Index || math.Float64bits(oldA[i].Value) != math.Float64bits(oldB[i].Value) {
					t.Fatal("fixture no longer demonstrates legacy information loss")
				}
			}
			for _, side := range []string{"query", "document"} {
				t.Run(side, func(t *testing.T) {
					qa, da, qb, db := "inspect boundary", f.a, "inspect boundary", f.b
					if side == "query" {
						qa, da, qb, db = f.a, "inspect boundary", f.b, "inspect boundary"
					}
					a, ae := Extract(qa, da)
					b, be := Extract(qb, db)
					if ae != nil || be != nil || symbolBitsEqual(a, b) {
						t.Fatalf("retained symbols failed to distinguish this bounded fixture: %v, %v", ae, be)
					}
				})
			}
		})
	}
}

func TestLongestComparisonsAndLiteralWordIsolation(t *testing.T) {
	got, err := tokenize("a<=b a>=b a<b a>b [a,b) (a,b] gte lte lt gt")
	want := []string{
		"a", "\x01<=", "b", "a", "\x01>=", "b", "a", "\x01<", "b", "a", "\x01>", "b",
		"\x01[", "a", "b", "\x01)", "\x01(", "a", "b", "\x01]", "gte", "lte", "lt", "gt",
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("comparison pairs must be one token and brackets retain their side: got %q, error %v", got, err)
	}
	for _, f := range []struct{ symbol, literal string }{{"<", "lt"}, {"<=", "lte"}, {">", "gt"}, {">=", "gte"}} {
		a, ae := Extract("inspect relation", "a "+f.symbol+" b")
		b, be := Extract("inspect relation", "a "+f.literal+" b")
		if ae != nil || be != nil || symbolBitsEqual(a, b) {
			t.Fatalf("symbol %q aliases ordinary word %q: %v, %v", f.symbol, f.literal, ae, be)
		}
	}
	// A raw control prefix is punctuation, not a way to inject an internal tag.
	words, err := tokenize("gte \x01gte GT")
	if err != nil || !reflect.DeepEqual(words, []string{"gte", "gte", "gt"}) {
		t.Fatalf("raw ordinary words entered the symbol namespace: %q, %v", words, err)
	}
}

func TestInputBoundaryAcceptanceAndAtomicRejection(t *testing.T) {
	valid := []struct{ name, text string }{
		{"512-ascii-bytes", strings.Repeat("a", MaxBytes)},
		{"512-utf8-bytes", strings.Repeat("é", MaxBytes/2)},
		{"32-words", strings.TrimSpace(strings.Repeat("a ", MaxWords))},
		{"64-symbol-tokens", strings.Repeat("[", MaxTokens)},
		{"64-long-comparisons", strings.Repeat("<=", MaxTokens)},
		{"32-words-and-32-symbols", strings.Repeat("a[", MaxWords)},
	}
	for _, f := range valid {
		for _, side := range []string{"query", "document"} {
			t.Run(f.name+"/"+side, func(t *testing.T) {
				query, document := f.text, "alpha"
				if side == "document" {
					query, document = document, query
				}
				got, err := Extract(query, document)
				if err != nil || len(got) == 0 {
					t.Fatalf("inclusive boundary rejected: %v", err)
				}
			})
		}
	}
	invalid := []struct{ name, text string }{
		{"513-ascii-bytes", strings.Repeat("a", MaxBytes+1)},
		{"514-utf8-bytes", strings.Repeat("é", MaxBytes/2+1)},
		{"33-words", strings.TrimSpace(strings.Repeat("a ", MaxWords+1))},
		{"65-symbol-tokens", strings.Repeat("[", MaxTokens+1)},
		{"65-long-comparisons", strings.Repeat("<=", MaxTokens+1)},
		{"65-mixed-tokens", strings.Repeat("a[", MaxWords) + "]"},
		{"invalid-leading-byte", string([]byte{0xff})},
		{"truncated-utf8", string([]byte{0xe2, 0x82})},
		{"utf8-surrogate", string([]byte{0xed, 0xa0, 0x80})},
	}
	for _, f := range invalid {
		for _, other := range []string{"alpha", ""} {
			for _, side := range []string{"query", "document"} {
				name := f.name + "/" + side
				if other == "" {
					name += "/empty-other"
				}
				t.Run(name, func(t *testing.T) {
					query, document := f.text, other
					if side == "document" {
						query, document = document, query
					}
					got, err := Extract(query, document)
					if err != ErrInput || got != nil {
						t.Fatalf("invalid input must return ErrInput and no partial features: len=%d, %v", len(got), err)
					}
				})
			}
		}
	}
}

func TestDeterministicCompactOrderAndCallerOwnership(t *testing.T) {
	fixtures := []struct{ query, document string }{
		{"alpha < beta", "omega [ value )"},
		{strings.Repeat("a[", MaxWords), strings.Repeat("b>", MaxWords)},
		{"repeat repeat repeat", "repeat repeat repeat"},
	}
	for _, f := range fixtures {
		want, err := Extract(f.query, f.document)
		if err != nil || len(want) == 0 || len(want) > Dimension {
			t.Fatalf("bounded feature extraction: %v", err)
		}
		// Legacy storage order is sorted compact cross indices, then overlap0.
		// Globally sorting index0 first would break exact legacy compatibility.
		for i, feature := range want[:len(want)-1] {
			if feature.Index < 1 || feature.Index >= Dimension || (i > 0 && feature.Index <= want[i-1].Index) || math.IsNaN(feature.Value) || math.IsInf(feature.Value, 0) {
				t.Fatal("cross features are not finite, bounded and strictly ordered")
			}
		}
		overlap := want[len(want)-1]
		if overlap.Index != 0 || overlap.Value < 0 || overlap.Value > 1 || math.IsNaN(overlap.Value) {
			t.Fatal("terminal overlap feature changed")
		}
		first, err := Extract(f.query, f.document)
		if err != nil || !symbolBitsEqual(first, want) {
			t.Fatal("repeat extraction changed feature bits or order", err)
		}
		first[0] = Feature{Index: -1, Value: math.NaN()}
		first[len(first)-1] = Feature{Index: -2, Value: math.Inf(1)}
		next, err := Extract(f.query, f.document)
		if err != nil || !symbolBitsEqual(next, want) {
			t.Fatal("caller mutation reached another extraction", err)
		}
		_, err = Extract("different <= relation", "different ( interval ]")
		if err != nil || !symbolBitsEqual(next, want) {
			t.Fatal("another extraction mutated retained caller storage", err)
		}
	}
	// Independent callers also own their returned slices under concurrent use.
	want, err := Extract("alpha < beta", "omega [ value )")
	if err != nil {
		t.Fatal(err)
	}
	const callers = 8
	var wg sync.WaitGroup
	errors := make(chan error, callers)
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				got, err := Extract("alpha < beta", "omega [ value )")
				if err != nil || !symbolBitsEqual(got, want) {
					errors <- fmt.Errorf("concurrent caller storage changed: %v", err)
					return
				}
				got[0] = Feature{Index: -1, Value: math.NaN()}
			}
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
