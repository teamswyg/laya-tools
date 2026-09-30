package hintsearch

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
)

func helperBaseline(n int) []int {
	base := make([]int, n)
	for i := range base {
		base[i] = n - i - 1
	}
	return base
}

func checkHelperOrder(t testing.TB, base, got []int) {
	t.Helper()
	if len(got) != len(base) || got[0] != base[0] {
		t.Fatal("lost candidates or first baseline candidate")
	}
	ranks := make([]int, len(base))
	for i, d := range got {
		if d < 0 || d >= len(base) || ranks[d] != 0 {
			t.Fatal("invalid helper permutation")
		}
		ranks[d] = i + 1
	}
	for i, d := range base {
		if ranks[d] > min(len(base), 2*(i+1)-1) {
			t.Fatal("helper exceeded baseline unit-inspection bound")
		}
	}
}

// This reference looks up each candidate's independent one-based rank in each
// field, rather than accumulating the implementation's traversal order.
func referenceHelperRRF(full, names Ranking) []int {
	type scored struct {
		d     int
		score float64
	}
	var ranked []scored
	for d := range full.Order {
		score := 0.
		for _, field := range []Ranking{full, names} {
			if field.Scores[d] <= 0 {
				continue
			}
			rank := slices.Index(field.Order, d) + 1
			score += 1 / float64(60+rank)
		}
		if score > 0 {
			ranked = append(ranked, scored{d, score})
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score != ranked[j].score {
			return ranked[i].score > ranked[j].score
		}
		return ranked[i].d < ranked[j].d
	})
	out := make([]int, min(PathHelperHintLimit, len(ranked)))
	for i := range out {
		out[i] = ranked[i].d
	}
	return out
}

func TestPathHelpersFrozenBM25Rules(t *testing.T) {
	paths := []string{"z/HTTPReader.go", "a/read_file.go", "src/HTTPRequest.go", "src/HTTP_Response.go", "doc/Reader.go", "테스트/파일.go", "unrelated.txt"}
	h, err := NewPathHelpers(paths)
	if err != nil {
		t.Fatal(err)
	}
	base := helperBaseline(len(paths))
	for _, query := range []string{"http reader", "HTTPRequest", "read_file", "테스트", "read file HTTP_Response", "absentword"} {
		got, err := h.Rank(query, base)
		if err != nil || len(got) != 3 {
			t.Fatal(err)
		}
		normalized, full, names := make([]string, len(paths)), make([]string, len(paths)), make([]string, len(paths))
		for i, p := range paths {
			normalized[i] = lexicalhint.NormalizeText(p)
			full[i] = strings.ToLower(p) + " " + lexicalhint.NormalizeText(p)
			name := p[strings.LastIndex(p, "/")+1:]
			names[i] = strings.ToLower(name) + " " + lexicalhint.NormalizeText(name)
		}
		ni, _ := NewPathTextIndex(paths, normalized)
		fi, _ := NewPathTextIndex(paths, full)
		bi, _ := NewPathTextIndex(paths, names)
		nq := lexicalhint.NormalizeText(query)
		jq := strings.ToLower(query) + " " + nq
		nr, _ := ni.RankLongInto(nq, Ranking{})
		fr, _ := fi.RankLongInto(jq, Ranking{})
		br, _ := bi.RankLongInto(jq, Ranking{})
		var positive []int
		for _, d := range nr.Order {
			if nr.Scores[d] > 0 && len(positive) < PathHelperHintLimit {
				positive = append(positive, d)
			}
		}
		wantNormal, _ := InterleavePathBaselineFirst(base, positive)
		fused := referenceHelperRRF(fr, br)
		wantJoined, _ := InterleavePathBaselineFirst(base, fused)
		if !slices.Equal(got[0].Order, wantNormal) || got[0].HintCount != len(positive) || !slices.Equal(got[1].Order, wantJoined) || got[1].HintCount != len(fused) {
			t.Fatalf("frozen rule mismatch for %q", query)
		}
		for i, result := range got {
			if result.ID != PathHelperIDs()[i] || !result.Attempted || result.Fallback != "" {
				t.Fatal("result contract", result)
			}
			checkHelperOrder(t, base, result.Order)
		}
	}
}

func TestHelperRRFOneBasedPositiveOnlyAndStableTies(t *testing.T) {
	full := Ranking{Order: []int{2, 0, 3, 1, 4}, Scores: []float64{3, 0, 4, 2, 0}}
	names := Ranking{Order: []int{1, 0, 2, 3, 4}, Scores: []float64{3, 4, 2, 0, 0}}
	got := reciprocalHelperHints(full, names)
	if !slices.Equal(got, []int{2, 0, 1, 3}) || !slices.Equal(got, referenceHelperRRF(full, names)) {
		t.Fatal("one-based or zero-score RRF regression", got)
	}
	// 1/61+1/63 is slightly greater than 2/62; rounding ranks or using score
	// magnitude instead of rank would change this ordering.
	if !(1./61+1./63 > 2./62) {
		t.Fatal("invalid RRF fixture")
	}
	full = Ranking{Order: []int{0, 1, 2}, Scores: []float64{10, 1, 0}}
	names = Ranking{Order: []int{1, 0, 2}, Scores: []float64{1, 10, 0}}
	if got := reciprocalHelperHints(full, names); !slices.Equal(got, []int{0, 1}) {
		t.Fatal("catalog tie or zero-field support changed", got)
	}
}

func TestHelperPositivePrefixAndJoinedAliases(t *testing.T) {
	r := Ranking{Order: make([]int, 100), Scores: make([]float64, 100)}
	for i := range r.Order {
		r.Order[i] = 99 - i
		if i < 80 {
			r.Scores[99-i] = float64(80 - i)
		}
	}
	before := Ranking{slices.Clone(r.Order), slices.Clone(r.Scores)}
	if got := positiveHelperHints(r); len(got) != 64 || !slices.Equal(got, r.Order[:64]) || !reflect.DeepEqual(r, before) {
		t.Fatal("positive cap or ownership")
	}
	if got := reciprocalHelperHints(r, r); len(got) != 64 || !slices.Equal(got, referenceHelperRRF(r, r)) || !reflect.DeepEqual(r, before) {
		t.Fatal("fused cap or field ownership")
	}
	clear(r.Scores)
	r.Scores[r.Order[0]], r.Scores[r.Order[1]] = 2, 1
	if got := positiveHelperHints(r); !slices.Equal(got, r.Order[:2]) {
		t.Fatal("zero scores leaked into prefix", got)
	}
	h, err := NewPathHelpers([]string{"src/HTTPRequest.go", "src/HTTP_Response.go"})
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []*Index{h.full, h.names} {
		raw, _ := field.RankLongInto("httprequest", Ranking{})
		split, _ := field.RankLongInto("http request", Ranking{})
		if raw.Scores[0] <= 0 || raw.Scores[1] != 0 || split.Scores[0] <= 0 || split.Scores[1] <= 0 {
			t.Fatal("joined and split identifiers were not both retained")
		}
	}
}

func TestExplicitHelperExactCaseBoundaryAndEvidenceOrder(t *testing.T) {
	paths := []string{"a/src/Reader.go", "b/src/Reader.go", "src/NotReader.go", "Reader.go", "x/reader.go", "src/LongReader.go", "δ/Δ.go", "x/HTTP-reader_2.go"}
	query := "Reader.go src/Reader.go src/LongReader.go Δ.go Reader.go /src/Reader.go HTTP-reader_2.go"
	got, comparisons, fallback := explicitHelperHints(query, paths)
	if fallback != "" || comparisons != 6*len(paths) || !slices.Equal(got, []int{5, 0, 1, 7, 3, 6}) {
		t.Fatal("anchor evidence", got, comparisons, fallback)
	}
	for _, tc := range []struct {
		q    string
		want []int
	}{
		{"reader.go", []int{4}},
		{"NotReader.go", []int{2}},
		{"Reader", nil},
		{"src/reader.go", nil},
		{"HTTP-reader_2.go", []int{7}},
		{"src/NotReader.go", []int{2}},
	} {
		got, _, fallback := explicitHelperHints(tc.q, paths)
		if fallback != "" || !slices.Equal(got, tc.want) {
			t.Fatal(tc.q, got, fallback)
		}
	}
	// A shared basename can support many paths, but the fixed prefix is stable.
	many := make([]string, 90)
	for i := range many {
		many[i] = fmt.Sprintf("folder%03d/shared.go", i)
	}
	got, _, _ = explicitHelperHints("shared.go", many)
	if len(got) != PathHelperHintLimit || got[0] != 0 || got[63] != 63 {
		t.Fatal("explicit cap or catalog ties", got)
	}
}

func TestPathHelpersLifecycleOwnershipNoEvidenceAndLabelBlindness(t *testing.T) {
	paths := []string{"src/Reader.go", "src/Writer.go", "src/Other.go"}
	h, err := NewPathHelpers(paths)
	if err != nil {
		t.Fatal(err)
	}
	base := []int{2, 1, 0}
	first, err := h.Rank("Reader src/Reader.go", base)
	if err != nil {
		t.Fatal(err)
	}
	builds := h.BuildWork()
	if builds != ([3]HelperWork{{IndexBuildAttempts: 1}, {IndexBuildAttempts: 2}, {}}) {
		t.Fatal("constructor work", builds)
	}
	if first[0].Work.IndexSearchAttempts != 1 || first[1].Work.IndexSearchAttempts != 2 || first[2].Work.IndexSearchAttempts != 0 || first[2].Work.AnchorComparisons != 3 {
		t.Fatal("query work", first)
	}
	// Synthetic scoring labels stay outside the ranking API and cannot select
	// a variant or alter ordering. Their oracle ranks can differ independently.
	for _, targets := range [][]int{{0}, {1}, {2}, {0, 1}, nil} {
		for _, result := range first {
			for _, d := range targets {
				if slices.Index(result.Order, d) < 0 {
					t.Fatal("label projection lost a candidate")
				}
			}
		}
		replay, err := h.Rank("Reader src/Reader.go", base)
		if err != nil || !reflect.DeepEqual(first, replay) {
			t.Fatal("labels or replay changed rank", err)
		}
	}
	held := make([][]int, len(first))
	for i := range first {
		held[i] = slices.Clone(first[i].Order)
	}
	paths[0] = "src/Changed.go"
	base[0], base[1] = base[1], base[0]
	if _, err := h.Rank("Writer", base); err != nil {
		t.Fatal(err)
	}
	for i := range first {
		if !slices.Equal(first[i].Order, held[i]) {
			t.Fatal("later call or caller mutation changed retained order")
		}
	}
	if h.paths[0] != "src/Reader.go" || h.BuildWork() != builds {
		t.Fatal("stale identity or hidden rebuilding")
	}
	none, err := h.Rank("absentword", base)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range none {
		if !result.Attempted || result.HintCount != 0 || result.Fallback != "" || !slices.Equal(result.Order, base) {
			t.Fatal("no-evidence baseline changed", result)
		}
	}
	none[0].Order[0] = 0
	if none[1].Order[0] != base[0] || none[2].Order[0] != base[0] {
		t.Fatal("sibling result alias")
	}
	ids := PathHelperIDs()
	ids[0] = "changed"
	if PathHelperIDs()[0] != NormalizedPositive64 {
		t.Fatal("shared ID array")
	}
}

func TestPathHelpersLargeCatalogAndCompleteQueries(t *testing.T) {
	paths := pathFixture(6000)
	h, err := NewPathHelpers(paths)
	if err != nil {
		t.Fatal(err)
	}
	base := make([]int, len(paths))
	for i := range base {
		base[i] = i
	}
	query := strings.Repeat("prose ", 2000) + "needle src/needle.go"
	got, err := h.Rank(query, base)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range got {
		checkHelperOrder(t, base, result.Order)
		if result.Fallback != "" || slices.Index(result.Order, len(paths)-1) != 1 {
			t.Fatal("late evidence or query tail was lost", result.ID, result.Fallback)
		}
	}
	query = strings.Repeat("x", MaxLongQueryBytes-len(" src/needle.go")) + " src/needle.go"
	got, err = h.Rank(query, base)
	if err != nil || got[2].HintCount != 1 || got[2].Order[1] != len(paths)-1 || got[1].Fallback != "joined_query_error" {
		t.Fatal("complete-query boundary or joined expansion", err)
	}
}

func TestPathHelperFieldPreflightBounds(t *testing.T) {
	if idx, fallback := newHelperField([]string{strings.Repeat("aa ", MaxPathHelperTokens)}, false, false); fallback != "" || idx == nil {
		t.Fatal("inclusive token bound", fallback)
	}
	if idx, fallback := newHelperField([]string{strings.Repeat("aa ", MaxPathHelperTokens+1)}, false, false); fallback != "catalog_token_budget" || idx != nil {
		t.Fatal("token preflight", fallback)
	}
	var text strings.Builder
	for i := 0; i <= MaxPathHelperVocabulary; i++ {
		fmt.Fprintf(&text, "token%06d ", i)
	}
	tooMany := text.String()
	if idx, fallback := newHelperField([]string{tooMany}, false, false); fallback != "catalog_vocabulary_budget" || idx != nil {
		t.Fatal("vocabulary preflight", fallback)
	}
	atLimit := tooMany[:strings.LastIndex(strings.TrimSpace(tooMany), " ")]
	if idx, fallback := newHelperField([]string{atLimit}, false, false); fallback != "" || idx == nil || len(idx.terms) != MaxPathHelperVocabulary {
		t.Fatal("inclusive vocabulary bound", fallback)
	}
	if idx, fallback := newHelperField([]string{strings.Repeat("x", MaxPathCatalogBytes+1)}, false, false); fallback != "catalog_text_budget" || idx != nil {
		t.Fatal("text preflight", fallback)
	}
}

func TestPathHelpersIndividualConstructionAndQueryFallback(t *testing.T) {
	paths := make([]string, 50000)
	for i := range paths {
		paths[i] = fmt.Sprintf("folder/file%05d.go", i)
	}
	h, err := NewPathHelpers(paths)
	if err != nil || h.normalized == nil || h.full != nil || h.joinedFallback != "joined_full_catalog_token_budget" {
		t.Fatal("individual field construction", err)
	}
	base := helperBaseline(len(paths))
	got, err := h.Rank("folder/file00000.go", base)
	if err != nil || got[1].Fallback != "joined_full_catalog_token_budget" || !slices.Equal(got[1].Order, base) || got[1].Work.IndexSearchAttempts != 0 || got[2].HintCount != 1 {
		t.Fatal("individual fallback", err)
	}
	for _, result := range got {
		checkHelperOrder(t, base, result.Order)
	}
	if h.BuildWork()[1].IndexBuildAttempts != 1 {
		t.Fatal("unused basename index was constructed")
	}
	small, _ := NewPathHelpers([]string{"a.go", "b.go"})
	base = []int{1, 0}
	got, err = small.Rank(strings.Repeat("aB", 50000), base)
	if err != nil || got[0].Fallback != "normalized_query_error" || got[1].Fallback != "joined_query_error" || got[2].Fallback != "" {
		t.Fatal("query expansion fallback", err)
	}
	for _, result := range got {
		if !slices.Equal(result.Order, base) || result.Work.IndexSearchAttempts != 0 {
			t.Fatal("failed query searched or lost baseline")
		}
	}
}

func TestExplicitHelperComparisonBudgetAndAnchorDeduplication(t *testing.T) {
	paths := make([]string, 8000)
	for i := range paths {
		paths[i] = fmt.Sprintf("folder/file%05d.go", i)
	}
	var query strings.Builder
	for i := 0; i < 1001; i++ {
		fmt.Fprintf(&query, "absent%04d.go ", i)
	}
	h, err := NewPathHelpers(paths)
	if err != nil {
		t.Fatal(err)
	}
	base := helperBaseline(len(paths))
	got, err := h.Rank(query.String(), base)
	if err != nil || got[2].Fallback != "explicit_path_comparison_budget" || got[2].Work.AnchorComparisons != 0 || got[2].HintCount != 0 || !slices.Equal(got[2].Order, base) {
		t.Fatal("complete-scan preflight", err)
	}
	hints, comparisons, fallback := explicitHelperHints(strings.Repeat("file00000.go ", 10000), paths)
	if fallback != "" || comparisons != len(paths) || !slices.Equal(hints, []int{0}) {
		t.Fatal("anchor repetition counted as independent work", comparisons, fallback)
	}
}

func TestPathHelpersSharedInputErrorsAndLegacyIdentityContract(t *testing.T) {
	for _, paths := range [][]string{nil, make([]string, MaxPathDocuments+1), {"same.go", "same.go"}, {"../bad"}, {"/absolute"}, {"a/../b"}, {"."}, {"a\x00b"}, {string([]byte{0xff})}, {strings.Repeat("x", 4097)}} {
		_, legacyError := NewPathIndex(paths)
		_, helperError := NewPathHelpers(paths)
		if legacyError == nil || helperError == nil {
			t.Fatal("shared path validation accepted invalid input")
		}
	}
	paths := []string{"z/last.go", "a/first.go", "한글/파일.go"}
	h, err := NewPathHelpers(paths)
	if err != nil {
		t.Fatal("constructor changed ordered identity contract", err)
	}
	for _, q := range []string{"", " \t\n", string([]byte{255}), strings.Repeat("a", MaxLongQueryBytes+1)} {
		if result, err := h.Rank(q, []int{0, 1, 2}); err == nil || result != nil {
			t.Fatal("invalid query reached helper search")
		}
	}
	for _, base := range [][]int{nil, {0, 1}, {0, 1, 1}, {0, 1, 3}, {-1, 1, 2}} {
		if result, err := h.Rank("query", base); err == nil || result != nil {
			t.Fatal("invalid baseline reached helper search")
		}
	}
	var absent *PathHelpers
	if _, err := absent.Rank("query", []int{0}); err == nil {
		t.Fatal("nil helper accepted")
	}
	if absent.BuildWork() != ([3]HelperWork{}) {
		t.Fatal("nil build accounting")
	}
}

func TestPathHelpersConcurrentImmutableReads(t *testing.T) {
	paths := pathFixture(100)
	h, err := NewPathHelpers(paths)
	if err != nil {
		t.Fatal(err)
	}
	base := helperBaseline(len(paths))
	query := "needle src/needle.go"
	want, err := h.Rank(query, base)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 5 {
				got, err := h.Rank(query, base)
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Error("parallel read mismatch", err)
					return
				}
				got[0].Order[0] = 0
			}
		}()
	}
	wg.Wait()
	if math.IsNaN(h.normalized.norms[0]) || h.BuildWork()[0].IndexBuildAttempts != 1 {
		t.Fatal("immutable index changed")
	}
}
