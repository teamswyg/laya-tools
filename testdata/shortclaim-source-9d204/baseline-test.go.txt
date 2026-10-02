package shortclaim

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func baselineTestPrepared(t testing.TB, request string, texts ...string) Prepared {
	t.Helper()
	var candidates [MaxCandidates]Candidate
	for i, text := range texts {
		candidates[i] = Candidate{ID: "candidate-" + strconv.Itoa(i), Text: text}
	}
	p, err := Validate(Input{Schema: Schema, Request: request, Provenance: "authored-baseline-control", Candidates: candidates[:len(texts)]})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func baselineTestRanks(t testing.TB, p Prepared) [4]Ranking {
	t.Helper()
	ranks, err := Baselines(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, ranking := range ranks {
		if ranking.Count != p.Count {
			t.Fatalf("candidate count changed: %d", ranking.Count)
		}
		var seen [MaxCandidates]bool
		for i := 0; i < ranking.Count; i++ {
			index := ranking.Order[i]
			if index < 0 || index >= p.Count || seen[index] {
				t.Fatalf("nonpermutation in %s", ranking.Kind)
			}
			seen[index] = true
			if math.IsNaN(ranking.Scores[index]) || math.IsInf(ranking.Scores[index], 0) {
				t.Fatalf("nonfinite score in %s", ranking.Kind)
			}
			if i > 0 && ranking.Scores[ranking.Order[i-1]] < ranking.Scores[index] {
				t.Fatalf("score order reversed in %s", ranking.Kind)
			}
			if i > 0 && ranking.Scores[ranking.Order[i-1]] == ranking.Scores[index] && ranking.Order[i-1] > index {
				t.Fatalf("unstable tie in %s", ranking.Kind)
			}
		}
		for i := ranking.Count; i < MaxCandidates; i++ {
			if ranking.Order[i] != 0 || ranking.Scores[i] != 0 {
				t.Fatalf("nonzero unused slot in %s", ranking.Kind)
			}
		}
	}
	return ranks
}

func TestBaselinesRetainAllAndStableTies(t *testing.T) {
	texts := []string{"same unknown words", "same unknown words", "same unknown words", "same unknown words", "same unknown words", "same unknown words", "same unknown words", "same unknown words"}
	p := baselineTestPrepared(t, "no matching tokens", texts...)
	original := p
	ranks := baselineTestRanks(t, p)
	if p != original {
		t.Fatal("prepared input was mutated")
	}
	wantKinds := [4]string{FixedOrderKind, BM25Kind, LexicalOrderedKind, NarrowRuleKind}
	for i, ranking := range ranks {
		if ranking.Kind != wantKinds[i] {
			t.Fatalf("kind %d = %q", i, ranking.Kind)
		}
		for j := 0; j < ranking.Count; j++ {
			if ranking.Order[j] != j || ranking.Scores[j] != 0 {
				t.Fatalf("zero-evidence tie reordered in %s", ranking.Kind)
			}
		}
	}
	if ranks[3].FallbackReason != "unsupported_rule_request" {
		t.Fatal("opaque request did not disclose fallback")
	}
}

func TestBaselinesDoNotUseIdentifiersOrProvenance(t *testing.T) {
	p := baselineTestPrepared(t, "keep active entries", "keep expired entries", "keep active entries", "remove expired entries")
	want := baselineTestRanks(t, p)
	p.Provenance = "different-source-role"
	for i := 0; i < p.Count; i++ {
		p.Candidates[i].ID = "target-looking-" + strconv.Itoa(p.Count-i)
	}
	if got := baselineTestRanks(t, p); got != want {
		t.Fatal("metadata changed text-only rankings")
	}
}

func TestBaselinesRejectForgedPrepared(t *testing.T) {
	p := baselineTestPrepared(t, "short request", "short candidate")
	controls := []struct {
		name string
		edit func(*Prepared)
	}{
		{"zero_count", func(p *Prepared) { p.Count = 0 }},
		{"negative_count", func(p *Prepared) { p.Count = -1 }},
		{"excess_count", func(p *Prepared) { p.Count = MaxCandidates + 1 }},
		{"unknown_schema", func(p *Prepared) { p.Schema = "other" }},
		{"forged_query", func(p *Prepared) { p.NormalizedRequest = "a different query" }},
		{"forged_candidate", func(p *Prepared) { p.Candidates[0].NormalizedText = "a different candidate" }},
		{"unused_slot", func(p *Prepared) { p.Candidates[1] = p.Candidates[0] }},
		{"raw_overflow", func(p *Prepared) { p.Request = strings.Repeat("x", MaxTextBytes+1) }},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			mutant := p
			control.edit(&mutant)
			got, err := Baselines(mutant)
			if err == nil || got != [4]Ranking{} {
				t.Fatal("invalid prepared input yielded rankings")
			}
		})
	}
}

func TestBaselinesBM25DocumentFrequencyAndQueryRepetition(t *testing.T) {
	p := baselineTestPrepared(t, "common rare", "common common", "rare", "common")
	ranks := baselineTestRanks(t, p)
	if ranks[1].Order[0] != 1 {
		t.Fatal("rare document term should outrank common-only terms")
	}
	// Here rare has df=1, tf=1, document length=1, average length=4/3.
	wantRare := math.Log(1+(3-1+0.5)/(1+0.5)) * 2.2 / (1 + 1.2*(0.25+0.75/(4.0/3.0)))
	if math.Abs(ranks[1].Scores[1]-wantRare) > 1e-12 {
		t.Fatalf("BM25 rare-term score %.15f, want %.15f", ranks[1].Scores[1], wantRare)
	}
	repeated := baselineTestPrepared(t, "common rare rare", "common common", "rare", "common")
	if got := baselineTestRanks(t, repeated)[1]; got != ranks[1] {
		t.Fatal("query repetition multiplied BM25 evidence")
	}
}

func TestBaselinesLexicalOrderAndRepeatedTokens(t *testing.T) {
	p := baselineTestPrepared(t, "red blue green", "green blue red", "red blue green", "red red red", "unrelated")
	ranking := baselineTestRanks(t, p)[2]
	if ranking.Order[0] != 1 || ranking.Scores[1] != 1 || ranking.Scores[0] != 0.5 || ranking.Scores[2] != 1.0/6.0 || ranking.Scores[3] != 0 {
		t.Fatal("word recall and ordered-bigram evidence did not distinguish order")
	}
	p = baselineTestPrepared(t, "red red blue red red blue", "red red blue", "red blue", "blue red")
	ranking = baselineTestRanks(t, p)[2]
	if ranking.Scores[0] <= ranking.Scores[1] || ranking.Scores[1] != ranking.Scores[2] {
		t.Fatal("duplicate ordered pairs were counted inconsistently")
	}
	p = baselineTestPrepared(t, "single", "single single", "other")
	ranking = baselineTestRanks(t, p)[2]
	if ranking.Scores[0] != 1 || ranking.Scores[1] != 0 {
		t.Fatal("one-word query requires word-only recall")
	}
}

func TestBaselinesNarrowRulesNegationAndActionDirection(t *testing.T) {
	p := baselineTestPrepared(t,
		"behavior-v1: if active then keep rows else remove rows",
		"behavior-v1: if active then remove rows else keep rows",
		"behavior-v1: if not active then remove rows else keep rows",
		"behavior-v1: if expired then keep rows else remove rows",
	)
	ranking := baselineTestRanks(t, p)[3]
	if ranking.FallbackReason != "" || ranking.Order[0] != 1 || ranking.Scores[1] <= 1 || ranking.Scores[0] >= 0.001 || ranking.Scores[2] >= 0.001 {
		t.Fatal("explicit condition negation or action direction was lost")
	}
	query, ok := baselineRuleParse(p.Request)
	negation, negOK := baselineRuleParse(p.Candidates[1].Text)
	if !ok || !negOK || query != negation {
		t.Fatal("negated condition and swapped branches are not canonical equivalents")
	}
}

func TestBaselinesNarrowRulesOperandAndBoundaryOrder(t *testing.T) {
	p := baselineTestPrepared(t,
		"behavior-v1: if count <= limit then accept else reject",
		"behavior-v1: if limit <= count then accept else reject",
		"behavior-v1: if count < limit then accept else reject",
		"behavior-v1: if count <= limit then accept else reject",
	)
	ranking := baselineTestRanks(t, p)[3]
	if ranking.FallbackReason != "" || ranking.Order[0] != 2 || ranking.Scores[0] >= 0.001 || ranking.Scores[1] >= 0.001 {
		t.Fatal("operand order or inclusive boundary was ignored")
	}
}

func TestBaselinesNarrowRulesClauseOrderAndActionNegation(t *testing.T) {
	p := baselineTestPrepared(t,
		"behavior-v1: if active then keep else remove; if ready then accept else reject",
		"behavior-v1: if ready then accept else reject; if active then keep else remove",
		"behavior-v1: if active then keep else remove; if ready then accept else reject",
	)
	ranking := baselineTestRanks(t, p)[3]
	if ranking.FallbackReason != "" || ranking.Order[0] != 1 || ranking.Scores[0] >= 0.001 {
		t.Fatal("ordered clauses were treated as a bag")
	}
	p = baselineTestPrepared(t,
		"behavior-v1: if active then not remove else keep",
		"behavior-v1: if active then keep else keep",
		"behavior-v1: if active then not remove else keep",
	)
	ranking = baselineTestRanks(t, p)[3]
	if ranking.FallbackReason != "" || ranking.Order[0] != 1 || ranking.Scores[0] >= 0.001 {
		t.Fatal("action negation was guessed to be an antonym")
	}
}

func TestBaselinesNarrowRulesWholeRequestFallback(t *testing.T) {
	request := "behavior-v1: if active then keep else remove"
	unsupported := []string{
		"ordinary prose about if active then keep else remove",
		"quoted behavior-v1: if active then keep else remove",
		"behavior-v1: if active then keep else remove unexplained trailing words",
		"behavior-v1: if active then keep else remove;",
		"behavior-v1: if active and ready then keep else remove",
		"behavior-v1: if count<=limit then accept else reject",
		"behavior-v1: if count = limit then accept else reject",
		"behavior-v1: if active then execute else remove",
		"behavior-v1: if 상태 then keep else remove",
		"Behavior-v1: if active then keep else remove",
		"behavior-v1: if not then keep else remove",
		"behavior-v1: if not then keep else remove rows",
		"behavior-v1: if active then keep else remove else keep",
		"behavior-v1: ",
	}
	for i, text := range unsupported {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			p := baselineTestPrepared(t, request, request, text, "behavior-v1: if active then remove else keep")
			ranks := baselineTestRanks(t, p)
			if ranks[3].FallbackReason != "unsupported_rule_candidate" || ranks[3].Order != ranks[1].Order || ranks[3].Scores != ranks[1].Scores {
				t.Fatal("unsupported candidate did not fall back to whole BM25 ranking")
			}
			p = baselineTestPrepared(t, text, request, text)
			ranks = baselineTestRanks(t, p)
			if ranks[3].FallbackReason != "unsupported_rule_request" || ranks[3].Order != ranks[1].Order || ranks[3].Scores != ranks[1].Scores {
				t.Fatal("unsupported request was partially parsed")
			}
		})
	}
}

func TestBaselinesBoundsAndSingleCandidate(t *testing.T) {
	query := strings.TrimSpace(strings.Repeat("same ", MaxNormalizedWords))
	p := baselineTestPrepared(t, query, query)
	ranks := baselineTestRanks(t, p)
	for _, ranking := range ranks {
		if ranking.Count != 1 || ranking.Order[0] != 0 {
			t.Fatal("single candidate was dropped")
		}
	}
	p = baselineTestPrepared(t, "조건 유지", "조건 유지", "조건 제거", "unrelated")
	baselineTestRanks(t, p)
}

func TestRankMatchesComparisonAndRejectsUnknown(t *testing.T) {
	for _, p := range []Prepared{
		baselineTestPrepared(t, "ordinary request", "ordinary candidate", "unrelated candidate"),
		baselineTestPrepared(t, "behavior-v1: if active then keep else remove", "behavior-v1: if active then remove else keep", "behavior-v1: if not active then remove else keep"),
	} {
		for _, want := range baselineTestRanks(t, p) {
			got, err := Rank(p, want.Kind)
			if err != nil || got != want {
				t.Fatalf("selected %s differs from comparison: %v", want.Kind, err)
			}
		}
		got, err := Rank(p, "unknown")
		if err == nil || err.Error() != "shortclaim_unknown_baseline" || got != (Ranking{}) {
			t.Fatal("unknown control yielded a ranking")
		}
		p.NormalizedRequest = "forged text"
		for _, kind := range [4]string{FixedOrderKind, BM25Kind, LexicalOrderedKind, NarrowRuleKind} {
			if got, err := Rank(p, kind); err == nil || got != (Ranking{}) {
				t.Fatalf("selected %s accepted forged input", kind)
			}
		}
	}
}

func TestBaselinesConcurrentImmutableInput(t *testing.T) {
	p := baselineTestPrepared(t, "short ordering request", "ordering short request", "short ordering request", "unrelated candidate")
	want := baselineTestRanks(t, p)
	for i := 0; i < 16; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			for j := 0; j < 4; j++ {
				got, err := Rank(p, want[j].Kind)
				if err != nil || got != want[j] {
					t.Fatalf("concurrent %s changed ranking", want[j].Kind)
				}
			}
		})
	}
}

func baselineBenchmarkPrepared(b testing.TB) Prepared {
	return baselineTestPrepared(b,
		"behavior-v1: if count <= limit then accept rows else reject rows",
		"behavior-v1: if count < limit then accept rows else reject rows",
		"behavior-v1: if count <= limit then reject rows else accept rows",
		"behavior-v1: if count <= limit then accept rows else reject rows",
		"behavior-v1: if not count <= limit then reject rows else accept rows",
		"behavior-v1: if limit <= count then accept rows else reject rows",
		"behavior-v1: if count >= limit then accept rows else reject rows",
		"behavior-v1: if count == limit then accept rows else reject rows",
		"behavior-v1: if count != limit then accept rows else reject rows",
	)
}

func BenchmarkBaselinesEightCandidates(b *testing.B) {
	p := baselineBenchmarkPrepared(b)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := Baselines(p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRankEightCandidates(b *testing.B) {
	p := baselineBenchmarkPrepared(b)
	for _, kind := range [4]string{FixedOrderKind, BM25Kind, LexicalOrderedKind, NarrowRuleKind} {
		b.Run(kind, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if _, err := Rank(p, kind); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
