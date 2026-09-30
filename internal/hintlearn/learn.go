// Package hintlearn implements an experimental encoder-free sparse ranker.
// Its dataset loader and ranking trainer use original synthetic data only.
// Feature extraction and serialization are also shared by the external pair
// trainer; no trained weights are embedded in source.
package hintlearn

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"sort"
	"strings"
	"unicode"
)

const Dimension = 8192
const Origin = "original_public_synthetic_domain_transfer"

type Document struct{ ID, Family, Text string }
type Query struct{ ID, Family, Text, Target string }
type Dataset struct {
	Schema, Origin, Split string
	Documents             []Document
	Queries               []Query
}

func Hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func Load(path, split string) (Dataset, string, error) {
	var d Dataset
	f, e := os.Open(path)
	if e != nil {
		return d, "", e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (8<<20)+1))
	if e != nil {
		return d, "", e
	}
	if len(b) > 8<<20 {
		return d, "", fmt.Errorf("dataset too large")
	}
	if e = json.Unmarshal(b, &d); e != nil {
		return d, "", e
	}
	if d.Schema != "riido-hint-pairs-v1" || d.Origin != Origin || d.Split != split || len(d.Documents) < 2 || len(d.Documents) > 256 || len(d.Queries) == 0 || len(d.Queries) > 1024 {
		return d, "", fmt.Errorf("invalid dataset")
	}
	ids := []string{}
	for _, v := range d.Documents {
		if v.ID == "" || v.Family == "" || len(v.Text) > 4096 || len(words(v.Text)) > 64 || v.Text == "" || slices.Contains(ids, v.ID) {
			return d, "", fmt.Errorf("invalid document")
		}
		ids = append(ids, v.ID)
	}
	qids := []string{}
	for _, q := range d.Queries {
		i := slices.Index(ids, q.Target)
		if i < 0 || q.Family != d.Documents[i].Family || q.Text == "" || len(q.Text) > 4096 || len(words(q.Text)) > 64 || q.ID == "" || slices.Contains(qids, q.ID) {
			return d, "", fmt.Errorf("invalid query")
		}
		qids = append(qids, q.ID)
	}
	return d, Hash(b), nil
}
func Disjoint(a, b Dataset) error {
	for _, x := range a.Documents {
		for _, y := range b.Documents {
			if x.Family == y.Family || x.ID == y.ID || x.Text == y.Text {
				return fmt.Errorf("family/document leakage")
			}
		}
	}
	for _, x := range a.Queries {
		for _, y := range b.Queries {
			if x.ID == y.ID || x.Text == y.Text {
				return fmt.Errorf("query leakage")
			}
		}
	}
	return nil
}

type Feature struct {
	Index int
	Value float64
}

func words(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}
func terms(ws []string) []string {
	ts := append([]string(nil), ws...)
	for i := 1; i < len(ws); i++ {
		ts = append(ts, ws[i-1]+"_"+ws[i])
	}
	return ts
}
func hash(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

// Features includes ordered bigrams: plain token bags cannot distinguish the
// fixture's opposing statements. Cross features learn associations, not rules.
func Features(query, document string) []Feature {
	qw, dw := words(query), words(document)
	qt, dt := terms(qw), terms(dw)
	if len(qt) == 0 || len(dt) == 0 {
		return nil
	}
	fs := make([]Feature, 0, len(qt)*len(dt)+1)
	scale := 1 / math.Sqrt(float64(len(qt)*len(dt)))
	for _, q := range qt {
		for _, d := range dt {
			h := hash(q + "\x00" + d)
			v := scale
			if h>>63 != 0 {
				v = -v
			}
			fs = append(fs, Feature{1 + int(h%(Dimension-1)), v})
		}
	}
	sort.Slice(fs, func(i, j int) bool { return fs[i].Index < fs[j].Index })
	n := 0
	for _, f := range fs {
		if n > 0 && fs[n-1].Index == f.Index {
			fs[n-1].Value += f.Value
		} else {
			fs[n] = f
			n++
		}
	}
	fs = fs[:n]
	sort.Strings(qw)
	qw = slices.Compact(qw)
	sort.Strings(dw)
	dw = slices.Compact(dw)
	hits := 0
	for _, q := range qw {
		if slices.Contains(dw, q) {
			hits++
		}
	}
	fs = append(fs, Feature{0, float64(hits) / float64(max(1, len(qw)))})
	return fs
}
func Score(w []float64, fs []Feature) float64 {
	s := 0.
	for _, f := range fs {
		s += w[f.Index] * f.Value
	}
	return s
}
func Quantize(w []float64, mode string) []float64 {
	out := make([]float64, len(w))
	scale := 0.
	switch mode {
	case "fp32":
		for i, v := range w {
			out[i] = float64(float32(v))
		}
		return out
	case "int8":
		for _, v := range w {
			scale = math.Max(scale, math.Abs(v))
		}
		scale = float64(float32(scale / 127))
	case "ternary_ste", "ternary_ptq":
		for _, v := range w {
			scale += math.Abs(v)
		}
		scale = float64(float32(scale / float64(len(w))))
	default:
		panic("unknown quantization")
	}
	if scale == 0 {
		return out
	}
	for i, v := range w {
		if mode == "int8" {
			out[i] = math.Max(-127, math.Min(127, math.Round(v/scale))) * scale
		} else if math.Abs(v) >= .7*scale {
			out[i] = math.Copysign(scale, v)
		}
	}
	return out
}

type prepared struct {
	features [][][]Feature
	targets  []int
}

func prepare(d Dataset) prepared {
	p := prepared{}
	for _, q := range d.Queries {
		row := make([][]Feature, len(d.Documents))
		target := -1
		for i, doc := range d.Documents {
			row[i] = Features(q.Text, doc.Text)
			if doc.ID == q.Target {
				target = i
			}
		}
		p.features = append(p.features, row)
		p.targets = append(p.targets, target)
	}
	return p
}

type Metrics struct {
	Recall1, Recall4, MeanChecks, NLL float64
	Ranks                             []int
}

func evaluate(w []float64, p prepared) Metrics {
	m := Metrics{}
	for i, fs := range p.features {
		s := make([]float64, len(fs))
		for j, f := range fs {
			s[j] = Score(w, f)
		}
		target := p.targets[i]
		rank := 1
		for j, v := range s {
			if v > s[target] || (v == s[target] && j < target) {
				rank++
			}
		}
		m.Ranks = append(m.Ranks, rank)
		if rank == 1 {
			m.Recall1++
		}
		if rank <= 4 {
			m.Recall4++
		}
		m.MeanChecks += float64(rank)
		mx := slices.Max(s)
		sum := 0.
		for _, v := range s {
			sum += math.Exp(v - mx)
		}
		m.NLL += math.Log(sum) + mx - s[target]
	}
	n := float64(len(p.targets))
	m.Recall1 /= n
	m.Recall4 /= n
	m.MeanChecks /= n
	m.NLL /= n
	return m
}
func Evaluate(w []float64, d Dataset) Metrics { return evaluate(w, prepare(d)) }

type Config struct {
	Seed         uint64
	LearningRate float64
	Mode         string
	Epochs       int
}
type Trial struct {
	Config     Config
	Epoch      int
	Validation Metrics
	Train      Metrics
	Weights    []float64 `json:"-"`
}

func Fit(train, val Dataset, c Config) (Trial, error) {
	var best Trial
	if train.Split != "train" || val.Split != "validation" {
		return best, fmt.Errorf("wrong gradient/selection split")
	}
	if e := Disjoint(train, val); e != nil {
		return best, e
	}
	if c.Epochs < 1 || c.Epochs > 100 || c.LearningRate <= 0 || c.LearningRate > 1 || math.IsNaN(c.LearningRate) || (c.Mode != "fp32" && c.Mode != "ternary_ste") {
		return best, fmt.Errorf("invalid fit config")
	}
	tp, vp := prepare(train), prepare(val)
	rng := rand.New(rand.NewPCG(c.Seed, c.Seed+1))
	w := make([]float64, Dimension)
	for i := range w {
		w[i] = rng.NormFloat64() * .01
	}
	best.Validation.MeanChecks = math.Inf(1)
	order := make([]int, len(tp.targets))
	for i := range order {
		order[i] = i
	}
	grad := make([]float64, Dimension)
	for epoch := 1; epoch <= c.Epochs; epoch++ {
		rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for _, i := range order {
			forward := Quantize(w, c.Mode)
			fs := tp.features[i]
			s := make([]float64, len(fs))
			for j, f := range fs {
				s[j] = Score(forward, f)
			}
			mx := slices.Max(s)
			sum := 0.
			for j := range s {
				s[j] = math.Exp(s[j] - mx)
				sum += s[j]
			}
			clear(grad)
			for j, f := range fs {
				g := s[j] / sum
				if j == tp.targets[i] {
					g--
				}
				for _, v := range f {
					grad[v.Index] += g * v.Value
				}
			}
			for k := range w {
				w[k] -= c.LearningRate * (grad[k] + .0001*w[k])
			}
		}
		forward := Quantize(w, c.Mode)
		m := evaluate(forward, vp)
		if m.MeanChecks < best.Validation.MeanChecks || (m.MeanChecks == best.Validation.MeanChecks && m.NLL < best.Validation.NLL) {
			best = Trial{c, epoch, m, evaluate(forward, tp), forward}
		}
	}
	return best, nil
}

// Rank is the bounded reference runtime for the research scorer. It rejects
// long inputs rather than silently truncating; callers can retain BM25 fallback.
func Rank(w []float64, query string, documents []string) ([]int, []float64, error) {
	if len(w) != Dimension || len(documents) < 1 || len(documents) > 256 || len(query) > 4096 || len(words(query)) == 0 || len(words(query)) > 64 {
		return nil, nil, fmt.Errorf("model input out of scope")
	}
	for _, v := range w {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, nil, fmt.Errorf("nonfinite coefficient")
		}
	}
	for _, d := range documents {
		if len(d) > 4096 || len(words(d)) > 64 {
			return nil, nil, fmt.Errorf("model input out of scope")
		}
	}
	order := make([]int, len(documents))
	scores := make([]float64, len(documents))
	for i, d := range documents {
		order[i] = i
		scores[i] = Score(w, Features(query, d))
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := order[i], order[j]
		if scores[a] == scores[b] {
			return a < b
		}
		return scores[a] > scores[b]
	})
	return order, scores, nil
}

// RuleRanks is a deliberately task-specific, nonlearned control. Its two
// substitutions encode the shared grammar explicitly, exposing whether this
// synthetic probe requires learning at all. It is not a general semantic model.
func RuleRanks(d Dataset) []int {
	ranks := []int{}
	for _, q := range d.Queries {
		query := strings.NewReplacer("retain", "keeps", "discard", "removes").Replace(q.Text)
		qt := terms(words(query))
		scores := make([]int, len(d.Documents))
		target := 0
		for i, doc := range d.Documents {
			dt := terms(words(doc.Text))
			for _, t := range qt {
				if slices.Contains(dt, t) {
					scores[i]++
				}
			}
			if doc.ID == q.Target {
				target = i
			}
		}
		rank := 1
		for i, s := range scores {
			if s > scores[target] || (s == scores[target] && i < target) {
				rank++
			}
		}
		ranks = append(ranks, rank)
	}
	return ranks
}
