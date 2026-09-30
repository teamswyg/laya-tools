// riido-staticprobe is a standalone experimental static embedding runner.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/staticembed"
	"io"
	"os"
	"runtime"
	"slices"
	"time"
	"unicode/utf8"
)

const planSHA = "0e48f9da91a345d850e7f61bf5f0b224c175c656179f9b3c3e0554974483ebcc"

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	stage := flag.String("stage", "embed", "setup, embed, validation or benchmark")
	dir := flag.String("model-dir", ".cache/potion-base-2M", "verified static model directory")
	mode := flag.String("mode", "fp32", "fp32, int8 or ternary")
	input := flag.String("input", ".cache/semantic-scale/cosqa-all.json", "pinned pair source, validation only")
	plan := flag.String("plan", "experiments/static-embedding/plan-04.json", "fixed validation plan")
	out := flag.String("out", "", "new aggregate report file, required for validation")
	flag.Parse()
	switch *stage {
	case "setup":
		if e := staticembed.Setup(*dir); e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(struct{ Status, Revision string }{"verified", staticembed.Revision})
	case "embed", "validation", "benchmark":
	default:
		return fmt.Errorf("invalid stage")
	}
	start := time.Now()
	m, e := staticembed.Load(*dir, *mode)
	if e != nil {
		return e
	}
	loadTime := time.Since(start).Seconds()
	if *stage == "embed" {
		text, e := readText(os.Stdin)
		if e != nil {
			return e
		}
		v, ids, e := m.Encode(text)
		if e != nil {
			return e
		}
		return json.NewEncoder(os.Stdout).Encode(struct {
			Schema, Mode string
			Vector       [64]float32
			TokenIDs     []int
			TableBytes   int
		}{"riido-static-vector-v1", *mode, v, ids, m.TableBytes()})
	}
	if *stage == "benchmark" {
		const text = "parse HTTP response headers and return the status code"
		const n = 2000
		for range 20 {
			if _, _, e = m.Encode(text); e != nil {
				return e
			}
		}
		times := make([]int64, n)
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		for i := range times {
			s := time.Now()
			if _, _, e = m.Encode(text); e != nil {
				return e
			}
			times[i] = time.Since(s).Nanoseconds()
		}
		runtime.ReadMemStats(&after)
		slices.Sort(times)
		return json.NewEncoder(os.Stdout).Encode(struct {
			Schema, Mode                          string
			Iterations, TableBytes                int
			LoadSeconds                           float64
			MedianNS, P95NS                       int64
			AllocatedBytesPerOp, AllocationsPerOp float64
			LiveHeapBytes                         uint64
		}{"riido-static-benchmark-v1", *mode, n, m.TableBytes(), loadTime, times[n/2], times[n*95/100], float64(after.TotalAlloc-before.TotalAlloc) / n, float64(after.Mallocs-before.Mallocs) / n, before.HeapAlloc})
	}
	if *out == "" {
		return fmt.Errorf("--out required for validation")
	}
	pb, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	if paireval.Hash(pb) != planSHA {
		return fmt.Errorf("plan hash mismatch")
	}
	rows, e := paireval.Load(*input)
	if e != nil {
		return e
	}
	split := paireval.Partition(rows)
	lex := paireval.NewLexical(rows, split)
	ps, bp, op := []paireval.Prediction{}, []paireval.Prediction{}, []paireval.Prediction{}
	excluded, zero := 0, 0
	for i, r := range rows {
		if split.Rows[i].Split != "validation" {
			continue
		}
		if !paireval.InScope(r) {
			excluded++
			continue
		}
		q, qids, e := m.Encode(r.Query)
		if e != nil {
			return e
		}
		c, cids, e := m.Encode(r.Code)
		if e != nil {
			return e
		}
		if len(qids) == 0 || len(cids) == 0 {
			zero++
		}
		p := paireval.Prediction{Row: i, Group: split.Rows[i].Group, Label: *r.Label, Eligible: true, Score: staticembed.Cosine(q, c)}
		ps = append(ps, p)
		p.Score = lex.Score(r.Query, r.Code)
		bp = append(bp, p)
		p.Score = paireval.Overlap(r.Query, r.Code)
		op = append(op, p)
	}
	auc, bauc := paireval.AUC(ps), paireval.AUC(bp)
	ready := *mode == "fp32" && auc != nil && bauc != nil && *auc >= .60 && *auc-*bauc >= .03
	r := struct {
		Schema, Mode, ModelRevision, ModelSHA256, TokenizerSHA256, PlanSHA256, PairSourceSHA256, MembershipSHA256 string
		ValidationEligible, ValidationExcluded, ZeroVectorPairs, TableBytes                                       int
		ValidationAUC, BM25ValidationAUC, OverlapValidationAUC                                                    *float64
		PrimaryReadyForFinal, Promotion                                                                           bool
		LoadSeconds, TotalSeconds                                                                                 float64
	}{"riido-static-validation-v1", *mode, staticembed.Revision, staticembed.ModelSHA, staticembed.VocabularySHA256, planSHA, paireval.SourceSHA, split.MembershipSHA256, len(ps), excluded, zero, m.TableBytes(), auc, bauc, paireval.AUC(op), ready, false, loadTime, time.Since(start).Seconds()}
	b, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(append(b, '\n'))
	ce := f.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	return json.NewEncoder(os.Stdout).Encode(r)
}

func readText(reader io.Reader) (string, error) {
	raw, e := io.ReadAll(io.LimitReader(reader, 65537))
	if e != nil {
		return "", e
	}
	if len(raw) > 65536 || !utf8.Valid(raw) {
		return "", fmt.Errorf("JSON input exceeds 64KiB or is invalid UTF-8")
	}
	var q struct {
		Text *string `json:"text"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e = d.Decode(&q); e != nil {
		return "", e
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return "", fmt.Errorf("one JSON object required")
	}
	if q.Text == nil {
		return "", fmt.Errorf("text field required")
	}
	return *q.Text, nil
}
