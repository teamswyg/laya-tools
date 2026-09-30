package paireval

import (
	"math"
	"math/rand/v2"
	"slices"
	"sort"
)

func (l *Lexical) Score(q, doc string) float64 {
	qs, ds := Words(q), Words(doc)
	length := len(ds)
	sort.Strings(qs)
	qs = slices.Compact(qs)
	sort.Strings(ds)
	score := 0.
	for _, t := range qs {
		v, found := slices.BinarySearch(l.terms, t)
		if !found {
			continue
		}
		pos := sort.SearchStrings(ds, t)
		tf := 0.
		for pos < len(ds) && ds[pos] == t {
			tf++
			pos++
		}
		if tf == 0 {
			continue
		}
		df := float64(l.df[v])
		idf := math.Log(1 + (float64(l.documents)-df+.5)/(df+.5))
		score += idf * tf * 2.2 / (tf + 1.2*(.25+.75*float64(length)/l.average))
	}
	return score
}
func Overlap(q, doc string) float64 {
	a, b := Words(q), Words(doc)
	sort.Strings(a)
	a = slices.Compact(a)
	sort.Strings(b)
	b = slices.Compact(b)
	hits := 0.
	for _, t := range a {
		if _, ok := slices.BinarySearch(b, t); ok {
			hits++
		}
	}
	return hits / float64(max(1, len(a)))
}

type Prediction struct {
	Row, Group, Label  int
	Score              float64
	Positive, Eligible bool
}
type Metrics struct {
	Cases, PositiveCases, TP, TN, FP, FN, Eligible int
	Accuracy, BalancedAccuracy, Precision, Recall  float64
	EligibleAUC                                    *float64
}

func Measure(ps []Prediction) Metrics {
	m := Metrics{Cases: len(ps)}
	for _, p := range ps {
		if p.Eligible {
			m.Eligible++
		}
		if p.Label == 1 {
			m.PositiveCases++
			if p.Positive {
				m.TP++
			} else {
				m.FN++
			}
		} else if p.Positive {
			m.FP++
		} else {
			m.TN++
		}
	}
	if m.Cases > 0 {
		m.Accuracy = float64(m.TP+m.TN) / float64(m.Cases)
	}
	if m.TP+m.FN > 0 {
		m.Recall = float64(m.TP) / float64(m.TP+m.FN)
	}
	if m.TP+m.FP > 0 {
		m.Precision = float64(m.TP) / float64(m.TP+m.FP)
	}
	if m.TN+m.FP > 0 {
		m.BalancedAccuracy = (m.Recall + float64(m.TN)/float64(m.TN+m.FP)) / 2
	}
	m.EligibleAUC = AUC(ps)
	return m
}
func AUC(ps []Prediction) *float64 {
	xs := []Prediction{}
	for _, p := range ps {
		if p.Eligible {
			xs = append(xs, p)
		}
	}
	sort.Slice(xs, func(i, j int) bool { return xs[i].Score < xs[j].Score })
	pos, neg := 0., 0.
	sum := 0.
	for i := 0; i < len(xs); {
		j := i + 1
		for j < len(xs) && xs[j].Score == xs[i].Score {
			j++
		}
		rank := float64(i+1+j) / 2
		for k := i; k < j; k++ {
			if xs[k].Label == 1 {
				pos++
				sum += rank
			} else {
				neg++
			}
		}
		i = j
	}
	if pos == 0 || neg == 0 {
		return nil
	}
	v := (sum - pos*(pos+1)/2) / (pos * neg)
	return &v
}
func Threshold(ps []Prediction) float64 {
	xs := []Prediction{}
	for _, p := range ps {
		if p.Eligible {
			xs = append(xs, p)
		}
	}
	if len(xs) == 0 {
		return 0
	}
	sort.Slice(xs, func(i, j int) bool { return xs[i].Score > xs[j].Score })
	pos, neg := 0., 0.
	for _, p := range xs {
		if p.Label == 1 {
			pos++
		} else {
			neg++
		}
	}
	best := math.Nextafter(xs[0].Score, math.Inf(1))
	bestBA := -1.
	tp, fp := 0., 0.
	if pos > 0 && neg > 0 {
		bestBA = .5
	}
	for i := 0; i < len(xs); {
		j := i
		for j < len(xs) && xs[j].Score == xs[i].Score {
			if xs[j].Label == 1 {
				tp++
			} else {
				fp++
			}
			j++
		}
		if pos > 0 && neg > 0 {
			ba := (tp/pos + (neg-fp)/neg) / 2
			if ba > bestBA {
				bestBA = ba
				best = xs[i].Score
			}
		}
		i = j
	}
	return best
}

type Interval struct {
	Low, High          float64
	Replicates, Groups int
}

func DeltaInterval(a, b []Prediction) Interval {
	type counts struct{ ap, an, bp, bn, p, n float64 }
	maxGroup := 0
	for _, v := range a {
		maxGroup = max(maxGroup, v.Group)
	}
	cs := make([]counts, maxGroup+1)
	for i, v := range a {
		c := &cs[v.Group]
		if v.Label == 1 {
			c.p++
			if v.Positive {
				c.ap++
			}
			if b[i].Positive {
				c.bp++
			}
		} else {
			c.n++
			if !v.Positive {
				c.an++
			}
			if !b[i].Positive {
				c.bn++
			}
		}
	}
	groups := []counts{}
	for _, c := range cs {
		if c.p+c.n > 0 {
			groups = append(groups, c)
		}
	}
	rng := rand.New(rand.NewPCG(1729, 1730))
	values := []float64{}
	for i := 0; i < 500; i++ {
		sum := counts{}
		for range groups {
			c := groups[rng.IntN(len(groups))]
			sum.ap += c.ap
			sum.an += c.an
			sum.bp += c.bp
			sum.bn += c.bn
			sum.p += c.p
			sum.n += c.n
		}
		if sum.p > 0 && sum.n > 0 {
			values = append(values, ((sum.ap-sum.bp)/sum.p+(sum.an-sum.bn)/sum.n)/2)
		}
	}
	sort.Float64s(values)
	if len(values) == 0 {
		return Interval{}
	}
	return Interval{values[len(values)*25/1000], values[min(len(values)-1, len(values)*975/1000)], len(values), len(groups)}
}
