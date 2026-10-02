package shortclaim_test

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func input(request string, texts ...string) shortclaim.Input {
	in := shortclaim.Input{Schema: shortclaim.Schema, Request: request, Provenance: "authored-validated-test"}
	for i, text := range texts {
		in.Candidates = append(in.Candidates, shortclaim.Candidate{ID: string(rune('a' + i)), Text: text})
	}
	return in
}

var kinds = [4]string{shortclaim.FixedOrderKind, shortclaim.BM25Kind, shortclaim.LexicalOrderedKind, shortclaim.NarrowRuleKind}

func TestValidatedRanksPreserveAllScoresOrderAndFallback(t *testing.T) {
	cases := []struct {
		name string
		in   shortclaim.Input
	}{
		{"eight_ties", input("unknown words", "same words", "same words", "same words", "same words", "same words", "same words", "same words", "same words")},
		{"repeated_order", input("red red blue red red blue", "red blue", "blue red", "red red blue")},
		{"unicode_camel", input("KeepActive 항목; remove EXPIRED entries", "removeActive entries", "retainActive 항목", "keep EXPIRED entries")},
		{"rule_negation", input("behavior-v1: if active then keep else remove", "behavior-v1: if active then remove else keep", "behavior-v1: if not active then remove else keep")},
		{"rule_clauses", input("behavior-v1: if active then keep else remove; if ready then accept else reject", "behavior-v1: if ready then accept else reject; if active then keep else remove", "behavior-v1: if active then keep else remove; if ready then accept else reject")},
		{"request_fallback", input("ordinary prose about if active then keep", "keep active", "remove active")},
		{"candidate_fallback", input("behavior-v1: if active then keep else remove", "behavior-v1: if active then keep else remove", "unsupported syntax")},
		{"word_bound", input(strings.TrimSpace(strings.Repeat("same ", 32)), "same")},
		{"byte_bound", input(strings.Repeat("x", 512), strings.Repeat("x", 512))},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			prepared, err := shortclaim.Validate(c.in)
			if err != nil {
				t.Fatal(err)
			}
			validated, err := shortclaim.ValidateInput(c.in)
			if err != nil || validated.Prepared() != prepared {
				t.Fatal("constructor changed prepared value")
			}
			wire, err := json.Marshal(c.in)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := shortclaim.LoadValidated(strings.NewReader(string(wire)))
			if err != nil || loaded != validated {
				t.Fatal("wire and direct constructors differ")
			}
			for _, kind := range kinds {
				want, err := shortclaim.Rank(prepared, kind)
				if err != nil {
					t.Fatal(err)
				}
				got, err := validated.Rank(kind)
				if err != nil || got != want {
					t.Fatalf("ranking changed for %s", kind)
				}
			}
		})
	}
}

func TestValidatedConstructorsReturnZeroOnInvalidInput(t *testing.T) {
	base := input("keep records", "keep records")
	bad := []shortclaim.Input{
		input("", "keep records"), input(strings.Repeat("x", 513), "keep records"),
		input(strings.TrimSpace(strings.Repeat("word ", 33)), "keep records"),
		input("keep records", ""), input("keep records"), input("keep records", string([]byte{0xff})),
	}
	duplicate := input("keep records", "keep records", "remove records")
	duplicate.Candidates[1].ID = duplicate.Candidates[0].ID
	bad = append(bad, duplicate)
	for i, in := range bad {
		_, want := shortclaim.Validate(in)
		v, err := shortclaim.ValidateInput(in)
		if want == nil || err != want || v != (shortclaim.ValidatedInput{}) {
			t.Fatalf("invalid direct case %d changed contract", i)
		}
	}
	wire, _ := json.Marshal(base)
	for _, raw := range []string{
		string(wire) + strings.Repeat(" ", shortclaim.MaxJSONBytes),
		strings.Replace(string(wire), `"request":"keep records"`, `"request":"keep records","request":"private-test-marker"`, 1),
		strings.Replace(string(wire), `"request":"keep records"`, `"request":"\ud800"`, 1),
		strings.Replace(string(wire), `"schema":`, `"Schema":`, 1),
	} {
		_, want := shortclaim.Load(strings.NewReader(raw))
		v, err := shortclaim.LoadValidated(strings.NewReader(raw))
		if want == nil || err != want || v != (shortclaim.ValidatedInput{}) {
			t.Fatal("invalid wire changed contract")
		}
		if strings.Contains(err.Error(), "private-test-marker") {
			t.Fatal("raw input leaked")
		}
	}
	if v, err := shortclaim.LoadValidated(nil); err != shortclaim.ErrRead || v != (shortclaim.ValidatedInput{}) {
		t.Fatal("nil reader")
	}
}

func TestValidatedZeroAndJSONForgeCannotRank(t *testing.T) {
	var zero shortclaim.ValidatedInput
	for _, kind := range kinds {
		if got, err := zero.Rank(kind); err != shortclaim.ErrPrepared || got != (shortclaim.Ranking{}) {
			t.Fatal("zero input ranked")
		}
	}
	if err := json.Unmarshal([]byte(`{"ready":true,"prepared":{"Count":1,"NormalizedRequest":"forged"}}`), &zero); err != nil {
		t.Fatal(err)
	}
	if got, err := zero.Rank(shortclaim.LexicalOrderedKind); err != shortclaim.ErrPrepared || got != (shortclaim.Ranking{}) {
		t.Fatal("wire forged private ready state")
	}
	valid, err := shortclaim.ValidateInput(input("keep records", "keep records"))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := valid.Rank("private-test-marker"); err != shortclaim.ErrBaselineKind || got != (shortclaim.Ranking{}) {
		t.Fatal("unknown selector contract")
	}
	b, err := json.Marshal(valid)
	if err != nil || string(b) != "{}" {
		t.Fatal("opaque input exported raw fields")
	}
}

func TestValidatedOwnsValueAndPreparedCopies(t *testing.T) {
	in := input("keep active records", "remove active records", "keep active records")
	v, err := shortclaim.ValidateInput(in)
	if err != nil {
		t.Fatal(err)
	}
	want, err := v.Rank(shortclaim.LexicalOrderedKind)
	if err != nil {
		t.Fatal(err)
	}
	original := v.Prepared()
	in.Request = "changed request"
	in.Candidates[0] = shortclaim.Candidate{ID: "changed", Text: "changed"}
	copy := v.Prepared()
	copy.NormalizedRequest = "forged"
	copy.Candidates[0] = shortclaim.PreparedCandidate{ID: "changed", Text: "changed", NormalizedText: "changed"}
	copy.Count = 8
	if v.Prepared() != original {
		t.Fatal("caller mutation changed internal value")
	}
	if got, err := v.Rank(shortclaim.LexicalOrderedKind); err != nil || got != want {
		t.Fatal("ranking changed after copy mutation")
	}
	if got, err := shortclaim.Rank(copy, shortclaim.LexicalOrderedKind); err == nil || got != (shortclaim.Ranking{}) {
		t.Fatal("public Prepared no longer revalidates")
	}
	other := v
	other = shortclaim.ValidatedInput{}
	if other == v || v.Prepared() != original {
		t.Fatal("opaque value copies share mutable state")
	}
}

func TestPublicRankStillRejectsForgedNormalForms(t *testing.T) {
	v, err := shortclaim.ValidateInput(input("keep records", "keep records"))
	if err != nil {
		t.Fatal(err)
	}
	for _, edit := range []func(*shortclaim.Prepared){
		func(p *shortclaim.Prepared) { p.NormalizedRequest = "forged" },
		func(p *shortclaim.Prepared) { p.Candidates[0].NormalizedText = "forged" },
		func(p *shortclaim.Prepared) { p.Candidates[1] = p.Candidates[0] },
		func(p *shortclaim.Prepared) { p.Count = 0 },
		func(p *shortclaim.Prepared) { p.Request = strings.Repeat("x", 513) },
	} {
		p := v.Prepared()
		edit(&p)
		for _, kind := range kinds {
			if got, err := shortclaim.Rank(p, kind); err == nil || got != (shortclaim.Ranking{}) {
				t.Fatal("forged public Prepared accepted")
			}
		}
	}
}

func TestValidatedConcurrentRankUsesOwnedScratch(t *testing.T) {
	v, err := shortclaim.ValidateInput(input("keep active records", "keep inactive records", "keep active records", "remove records"))
	if err != nil {
		t.Fatal(err)
	}
	var want [4]shortclaim.Ranking
	for i, kind := range kinds {
		want[i], err = v.Rank(kind)
		if err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for i, kind := range kinds {
				if got, err := v.Rank(kind); err != nil || got != want[i] {
					t.Error("shared ranking state")
				}
			}
		})
	}
	wg.Wait()
}
