package retrievalbench

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

type Metrics struct {
	Queries                                   int
	Recall1, Recall5, Recall10, MRR, MeanRank float64
	P50Rank, P95Rank                          int
	QueryP50Nanos, QueryP95Nanos              int64
}
type Result struct {
	Mode                         string
	Documents, CatalogBytes      int
	BuildNanos                   int64
	All, NamePresent, NameAbsent Metrics
	ByRepository                 map[string]Metrics
}
type query struct {
	text       string
	targets    []int
	named      bool
	repository string
}
type observation struct {
	rank  int
	nanos int64
}

func namePresent(q, name string) bool {
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	for _, w := range strings.FieldsFunc(q, func(c rune) bool { return !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' }) {
		if name != "" && strings.EqualFold(w, name) {
			return true
		}
	}
	return false
}
func queries(rows []Row) []query {
	m := map[string]int{}
	var out []query
	for i, r := range rows {
		q := Normalize(r.Query)
		n, ok := m[q]
		if !ok {
			n = len(out)
			m[q] = n
			// Keep source case for camel-case splitting. The deduplication key
			// must not become the actual input to a case-sensitive preprocessor.
			out = append(out, query{text: strings.Join(strings.Fields(r.Query), " "), repository: r.Repository})
		}
		out[n].targets = append(out[n].targets, i)
		out[n].named = out[n].named || namePresent(q, r.Name)
		if out[n].repository != r.Repository {
			out[n].repository = "multiple"
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return digest([]byte(Normalize(out[i].text))) < digest([]byte(Normalize(out[j].text)))
	})
	return out
}
func summarize(obs []observation) Metrics {
	m := Metrics{Queries: len(obs)}
	if len(obs) == 0 {
		return m
	}
	ranks := make([]int, len(obs))
	times := make([]int64, len(obs))
	for i, o := range obs {
		ranks[i] = o.rank
		times[i] = o.nanos
		if o.rank <= 1 {
			m.Recall1++
		}
		if o.rank <= 5 {
			m.Recall5++
		}
		if o.rank <= 10 {
			m.Recall10++
		}
		m.MRR += 1 / float64(o.rank)
		m.MeanRank += float64(o.rank)
	}
	n := float64(len(obs))
	m.Recall1 /= n
	m.Recall5 /= n
	m.Recall10 /= n
	m.MRR /= n
	m.MeanRank /= n
	sort.Ints(ranks)
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	m.P50Rank = ranks[(len(obs)-1)/2]
	m.P95Rank = ranks[(len(obs)*95+99)/100-1]
	m.QueryP50Nanos = times[(len(obs)-1)/2]
	m.QueryP95Nanos = times[(len(obs)*95+99)/100-1]
	return m
}

// Evaluate uses known targets only; other documents remain unjudged.
// No learning, threshold selection, truncation or candidate dropping occurs.
func Evaluate(rows []Row, mode string) (Result, error) {
	r := Result{Mode: mode, Documents: len(rows), ByRepository: map[string]Metrics{}}
	if mode != "raw" && mode != "identifiers" {
		return r, fmt.Errorf("unknown mode")
	}
	start := time.Now()
	docs := make([]string, len(rows))
	for i, v := range rows {
		docs[i] = v.Code
		if mode == "identifiers" {
			docs[i] = lexicalhint.NormalizeText(v.Code)
		}
		r.CatalogBytes += len(docs[i])
	}
	idx, e := hintsearch.New(docs)
	r.BuildNanos = time.Since(start).Nanoseconds()
	if e != nil {
		return r, e
	}
	qs := queries(rows)
	var all, named, unnamed []observation
	byRepo := map[string][]observation{}
	for _, q := range qs {
		start = time.Now()
		text := q.text
		if mode == "identifiers" {
			text = lexicalhint.NormalizeText(text)
		}
		ranking, e := idx.Rank(text)
		elapsed := time.Since(start).Nanoseconds()
		if e != nil {
			return r, e
		}
		rank := len(rows) + 1
		for p, id := range ranking.Order {
			for _, target := range q.targets {
				if id == target && p+1 < rank {
					rank = p + 1
				}
			}
		}
		if rank > len(rows) {
			return r, fmt.Errorf("lost target")
		}
		o := observation{rank, elapsed}
		all = append(all, o)
		if q.named {
			named = append(named, o)
		} else {
			unnamed = append(unnamed, o)
		}
		byRepo[q.repository] = append(byRepo[q.repository], o)
	}
	r.All = summarize(all)
	r.NamePresent = summarize(named)
	r.NameAbsent = summarize(unnamed)
	for repo, obs := range byRepo {
		r.ByRepository[repo] = summarize(obs)
	}
	return r, nil
}
