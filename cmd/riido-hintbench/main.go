// riido-hintbench is an encoder-free, synthetic development experiment.
// Its oracle reads authored labels; it does not call an LLM or verify code.
package main

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"os"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"
)

type fixture struct{ Document, Direct, Paraphrase, Opposite string }

var fixtures = []fixture{
	{"payment idempotency prevents duplicate charges", "find payment idempotency", "stop charging a customer twice", "prevent duplicate charges not duplicate messages"},
	{"message deduplication prevents duplicate messages", "find message deduplication", "avoid processing the same event twice", "prevent duplicate messages not duplicate charges"},
	{"cache expiration removes stale entries", "find cache expiration", "discard out of date cached records", "remove stale entries not stale sessions"},
	{"session expiration removes stale sessions", "find session expiration", "end sign ins that have timed out", "remove stale sessions not stale entries"},
	{"certificate rotation renews service certificates", "find certificate rotation", "replace expiring transport credentials", "renew service certificates not service passwords"},
	{"password rotation renews service passwords", "find password rotation", "replace old account secrets", "renew service passwords not service certificates"},
	{"archive extraction rejects unsafe paths", "find archive extraction", "block zip entries escaping the destination", "reject unsafe paths not unsafe hosts"},
	{"outbound requests reject unsafe hosts", "find outbound requests", "block network access to private addresses", "reject unsafe hosts not unsafe paths"},
	{"pagination cursor advances result pages", "find pagination cursor", "fetch the next group of records", "advance result pages not log offsets"},
	{"consumer cursor advances log offsets", "find consumer cursor", "resume reading events after the saved position", "advance log offsets not result pages"},
	{"decimal rounding controls money precision", "find decimal rounding", "avoid fractional cent calculation errors", "control money precision not time precision"},
	{"timestamp rounding controls time precision", "find timestamp rounding", "reduce subsecond measurement detail", "control time precision not money precision"},
	{"file watcher detects local changes", "find file watcher", "notice when a document on disk is modified", "detect local changes not remote changes"},
	{"webhook listener detects remote changes", "find webhook listener", "receive notifications from an external service", "detect remote changes not local changes"},
	{"rate limiter restricts request frequency", "find rate limiter", "slow callers sending too many queries", "restrict request frequency not request size"},
	{"body limiter restricts request size", "find body limiter", "reject payloads that contain too many bytes", "restrict request size not request frequency"},
}

type query struct {
	text, kind string
	target     int
}

func queries() []query {
	var out []query
	for i, f := range fixtures {
		out = append(out, query{f.Direct, "direct", i}, query{f.Paraphrase, "paraphrase", i}, query{f.Opposite, "contrast", i})
	}
	for _, q := range []string{"schedule a dental appointment", "render a three dimensional scene", "train an image segmentation network", "translate a poem into French"} {
		out = append(out, query{q, "absent", -1})
	}
	return out
}
func tokens(s string) []string {
	ts := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	sort.Strings(ts)
	n := 0
	for _, t := range ts {
		if n == 0 || ts[n-1] != t {
			ts[n] = t
			n++
		}
	}
	return ts[:n]
}
func hash(s string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}
func overlap(a, b []string) int {
	i, j, n := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			n++
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return n
}

// Fixed-size arrays bound feature storage; there is no model, shared lock or map.
type bitset [64]uint64

func encode(ts []string, width int) (b bitset) {
	for _, t := range ts {
		h := hash(t) % uint64(width)
		b[h/64] |= 1 << (h % 64)
	}
	return
}
func intersect(a, b bitset, width int) int {
	n := 0
	for i := 0; i < width/64; i++ {
		n += bits.OnesCount64(a[i] & b[i])
	}
	return n
}

type index struct {
	terms        [][]string
	small, large []bitset
}

func newIndex() index {
	var x index
	for _, f := range fixtures {
		t := tokens(f.Document)
		x.terms = append(x.terms, t)
		x.small = append(x.small, encode(t, 256))
		x.large = append(x.large, encode(t, 4096))
	}
	return x
}
func (x index) rank(text, method string) []int {
	n := len(x.terms)
	order := make([]int, n)
	scores := make([]int, n)
	var ts []string
	var b bitset
	if method == "tokens" || method == "hash256" || method == "hash4096" {
		ts = tokens(text)
	}
	if method == "hash256" {
		b = encode(ts, 256)
	}
	if method == "hash4096" {
		b = encode(ts, 4096)
	}
	for i := range order {
		order[i] = i
		switch method {
		case "tokens":
			scores[i] = overlap(ts, x.terms[i])
		case "hash256":
			scores[i] = intersect(b, x.small[i], 256)
		case "hash4096":
			scores[i] = intersect(b, x.large[i], 4096)
		case "fixed_shuffle":
			scores[i] = int(hash(fmt.Sprint("seed1729-", i)) & 0x7fffffff)
		}
	}
	sort.Slice(order, func(i, j int) bool {
		if scores[order[i]] == scores[order[j]] {
			return order[i] < order[j]
		}
		return scores[order[i]] > scores[order[j]]
	})
	return order
}

type caseResult struct {
	Kind   string `json:"kind"`
	Target int    `json:"target"`
	Rank   int    `json:"rank"`
	Checks int    `json:"oracle_checks"`
}
type result struct {
	Method      string       `json:"method"`
	Recall1     float64      `json:"recall_at_1_present"`
	Recall4     float64      `json:"recall_at_4_present"`
	MeanChecks  float64      `json:"mean_oracle_checks_all"`
	WorstChecks int          `json:"worst_oracle_checks"`
	P50NS       int64        `json:"warm_p50_ns"`
	P95NS       int64        `json:"warm_p95_ns"`
	Allocs      float64      `json:"allocations_per_query"`
	Cases       []caseResult `json:"cases"`
}

func measure(x index, qs []query, method string) result {
	r := result{Method: method}
	hits1, hits4, present, total := 0, 0, 0, 0
	for _, q := range qs {
		order := x.rank(q.text, method)
		rank := 0
		checks := len(order)
		for i, v := range order {
			if v == q.target {
				rank = i + 1
				checks = rank
				break
			}
		}
		if q.target >= 0 {
			present++
			if rank == 1 {
				hits1++
			}
			if rank > 0 && rank <= 4 {
				hits4++
			}
		}
		total += checks
		r.WorstChecks = max(r.WorstChecks, checks)
		r.Cases = append(r.Cases, caseResult{q.kind, q.target, rank, checks})
	}
	r.Recall1 = float64(hits1) / float64(present)
	r.Recall4 = float64(hits4) / float64(present)
	r.MeanChecks = float64(total) / float64(len(qs))
	for i := 0; i < 1000; i++ {
		x.rank(qs[i%len(qs)].text, method)
	}
	times := make([]int64, 10000)
	for i := range times {
		start := time.Now()
		x.rank(qs[i%len(qs)].text, method)
		times[i] = time.Since(start).Nanoseconds()
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	r.P50NS = times[len(times)/2]
	r.P95NS = times[len(times)*95/100]
	i := 0
	r.Allocs = testing.AllocsPerRun(1000, func() { x.rank(qs[i%len(qs)].text, method); i++ })
	return r
}
func main() {
	x := newIndex()
	qs := queries()
	var rs []result
	for _, m := range []string{"catalog_order", "fixed_shuffle", "tokens", "hash256", "hash4096"} {
		rs = append(rs, measure(x, qs, m))
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	if err := e.Encode(struct {
		Schema, Scope, Go, Platform string
		Candidates, Queries         int
		Results                     []result
	}{"semantic-hints-baseline-v1", "authored synthetic development; no trained model; oracle checks are not LLM calls", runtime.Version(), runtime.GOOS + "/" + runtime.GOARCH, len(fixtures), len(qs), rs}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
