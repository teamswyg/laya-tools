// riido-longquery checks complete frozen requests against an original synthetic
// catalog. It measures input execution, not task retrieval or model quality.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

const planHash = "0e3262c9f780ae04995e33e90529c881e56380ea0ec0cac2caa5883d39ca44b0"

type input struct{ tag, query string }

func run() error {
	out := flag.String("out", "", "new private output directory")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require output directory")
	}
	b, e := os.ReadFile("experiments/long-query/plan-32.json")
	if e != nil {
		return e
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != planHash {
		return fmt.Errorf("plan mismatch")
	}
	full, e := sweaudit.Read(".cache/real-task-full-28.jsonl", sweaudit.FullProjectionSHA256)
	if e != nil {
		return e
	}
	multi, e := sweaudit.Read(".cache/real-task-multilingual-28.jsonl", sweaudit.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	groups, selected, e := sweaudit.GroupEvaluation(full, multi)
	if e != nil {
		return e
	}
	if groups.SelectedSHA256 != sweaudit.FrozenSelectionSHA256 || len(selected) != 2400 {
		return fmt.Errorf("membership mismatch")
	}
	var inputs []input
	for si, rows := range [][]sweaudit.Row{full, multi} {
		for _, r := range rows {
			inputs = append(inputs, input{[]string{"full", "multilingual"}[si] + "\x00" + r.ID, r.Request})
		}
	}
	slices.SortFunc(inputs, func(a, b input) int { return strings.Compare(a.tag, b.tag) })
	queries := make([]string, len(selected))
	for i, s := range selected {
		pos, ok := slices.BinarySearchFunc(inputs, s.Source+"\x00"+s.ID, func(a input, b string) int { return strings.Compare(a.tag, b) })
		if !ok {
			return fmt.Errorf("missing selected input")
		}
		queries[i] = inputs[pos].query
	}
	docs := make([]string, 4096)
	for i := range docs {
		docs[i] = fmt.Sprintf("document %04d source code function error test implementation branch repository", i)
	}
	idx, e := hintsearch.New(docs)
	if e != nil {
		return e
	}
	report := struct {
		Schema, PlanSHA256, SelectedSHA256                                                                                                                      string
		Requests, OriginalBytes, MaxRequestBytes, LongRequests, Executed, Failures, ShortParityChecks, ShortParityFailures, TruncatedRequests, CatalogDocuments int
		SyntheticCatalog, ActualTaskQualityEstablished, ProductionReady                                                                                         bool
	}{Schema: "riido-long-query-execution-v1", PlanSHA256: planHash, SelectedSHA256: groups.SelectedSHA256, Requests: len(queries), CatalogDocuments: len(docs), SyntheticCatalog: true}
	var dst hintsearch.Ranking
	// Warm only original text to allocate caller buffers; do not warm real tasks.
	dst, e = idx.RankLongInto("original fixture", dst)
	if e != nil {
		return e
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	start := time.Now()
	for _, q := range queries {
		report.OriginalBytes += len(q)
		report.MaxRequestBytes = max(report.MaxRequestBytes, len(q))
		if len(q) > hintsearch.MaxQueryBytes {
			report.LongRequests++
		}
		dst, e = idx.RankLongInto(q, dst)
		if e != nil {
			report.Failures++
			continue
		}
		report.Executed++
	}
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	// Parity runs are deliberately outside ranking timing/allocation statistics.
	for _, q := range queries {
		if len(q) > hintsearch.MaxQueryBytes {
			continue
		}
		a, e := idx.Rank(q)
		if e != nil {
			return e
		}
		dst, e = idx.RankLongInto(q, dst)
		if e != nil {
			return e
		}
		report.ShortParityChecks++
		if !reflect.DeepEqual(a, dst) {
			report.ShortParityFailures++
		}
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	b, e = json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "results.json"), append(b, '\n'), 0600); e != nil {
		return e
	}
	fmt.Fprintf(os.Stderr, "ranking_elapsed_ns=%d ranking_total_alloc_bytes=%d ranking_mallocs=%d\n", elapsed.Nanoseconds(), after.TotalAlloc-before.TotalAlloc, after.Mallocs-before.Mallocs)
	if report.Failures != 0 || report.ShortParityFailures != 0 {
		return fmt.Errorf("execution or parity failure; see aggregate")
	}
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
