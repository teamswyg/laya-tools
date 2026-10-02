package pairlearn

import (
	"fmt"
	"math"
	"slices"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
)

const (
	RankingSchema = "riido-pairlearn-same-parent-ranking-v1"
	// This cap counts the new association, pair and diagnostic backing arrays.
	// It is not an OS RSS limit or a cap on the caller's sparse feature dataset.
	MaxRankingPayloadBytes = 64 << 20
	MaxRankingParentRows   = 8
)

// RankingConfig adds one fixed, opt-in objective. Lambda accepts only 0 or 1;
// nil disables it without inspecting association metadata. Parent IDs associate
// existing supervision rows, never features. The caller supplies all candidates
// of each known/no-answer parent and excludes unknown/calibration from datasets.
// This package cannot prove source clearance or caller-frozen truth/role bindings.
type RankingConfig struct {
	Schema                                 string
	Lambda                                 float64
	TrainingParentIDs, ValidationParentIDs []int
}

// PairAudit owns SoA columns for every positive/negative pair, including pairs
// with zero-weight endpoints when enabled. Weights are supervision weights, not model
// coefficients. Nonzero weights sum to one across eligible parents: average
// endpoint-weight products inside each parent, then average eligible parents.
type PairAudit struct {
	Rows, Parents, NoAnswerParents, EligibleParents, ZeroWeightPairs int
	ParentIDs, PositiveRows, NegativeRows                            []int
	Weights                                                          []float64
	PayloadBytes                                                     uint64
}

type RankingLoss struct {
	Defined bool
	Value   float64
}

type RankingEpoch struct {
	Epoch                int
	Training, Validation RankingLoss
}

// RankingTrace is separate from the historical BCE Trace. It neither chooses
// an epoch nor changes the earliest minimum validation-BCE selector. Undefined
// loss means there are no eligible pairs, not a measured loss of zero.
type RankingTrace struct {
	Schema        string
	Enabled       bool
	Lambda        float64
	Training      PairAudit
	Validation    PairAudit
	Epochs        []RankingEpoch
	SelectedEpoch int
	Complete      bool
	FailureEpoch  int
	FailureKind   string
}

type pairPlan struct {
	audit   PairAudit
	offsets []int // Pair range owned by each positive endpoint row.
}

type rankingObjective struct {
	lambda               float64
	training, validation pairPlan
}

// FitWithRanking preserves original BCE rows/weights and adds a same-parent
// RankNet-style term. For each shuffled minibatch, the positive endpoint owns
// its pair contributions. Multiplying by the full row count before the existing
// batch-length division gives an unbiased uniformly sampled minibatch gradient
// estimator at fixed coefficients. Existing random reshuffling/SGD still updates
// coefficients between batches; it is not an independent sampling guarantee.
// Scores for both endpoints use the same quantized coefficients as BCE, even
// when the negative endpoint belongs to a different minibatch.
func FitWithRanking(train, validation Dataset, c Config, cfg *RankingConfig) (Result, error) {
	if cfg == nil {
		return Fit(train, validation, c)
	}
	objective, _, err := prepareRanking(train, validation, c, cfg)
	if err != nil {
		return Result{}, err
	}
	return fitRanking(train, validation, c, nil, objective, nil)
}

// FitWithRankingTrace additionally records the original BCE Trace and separate
// pair-loss observations. Lambda zero follows the identical legacy trainer and
// emits no pair-loss epochs. Invalid opt-in metadata errors never silently turn
// into a successful fallback. All caller inputs remain immutable during a call.
func FitWithRankingTrace(train, validation Dataset, c Config, cfg *RankingConfig) (Result, Trace, RankingTrace, error) {
	if cfg == nil {
		r, t, err := FitWithTrace(train, validation, c)
		return r, t, RankingTrace{Schema: RankingSchema, Complete: t.Complete, SelectedEpoch: r.Epoch, FailureEpoch: t.FailureEpoch, FailureKind: t.FailureKind}, err
	}
	objective, rt, err := prepareRanking(train, validation, c, cfg)
	if err != nil {
		return Result{}, Trace{Schema: TraceSchema, FailureKind: "input_validation"}, RankingTrace{Schema: RankingSchema, FailureKind: "input_validation"}, err
	}
	if objective == nil {
		r, t, err := FitWithTrace(train, validation, c)
		rt.Complete, rt.SelectedEpoch = t.Complete, r.Epoch
		if err != nil {
			rt.FailureEpoch, rt.FailureKind = t.FailureEpoch, t.FailureKind
		}
		return r, t, rt, err
	}
	t := Trace{Schema: TraceSchema}
	r, err := fitRanking(train, validation, c, &fitObserver{trace: &t}, objective, &rt)
	t.SelectedEpoch, rt.SelectedEpoch = r.Epoch, r.Epoch
	if err != nil {
		if t.FailureKind == "" {
			t.FailureKind = "input_validation"
		}
		if rt.FailureKind == "" {
			rt.FailureKind, rt.FailureEpoch = t.FailureKind, t.FailureEpoch
		}
		return r, t, rt, err
	}
	if t.FailureKind != "" || len(t.Epochs) != c.Epochs || len(rt.Epochs) != c.Epochs {
		rt.FailureKind, rt.FailureEpoch = t.FailureKind, t.FailureEpoch
		if rt.FailureKind == "" {
			rt.FailureKind = "incomplete_epoch_trace"
		}
		return r, t, rt, fmt.Errorf("incomplete ranking trace")
	}
	t.Complete, rt.Complete = true, true
	return r, t, rt, nil
}

type association struct {
	ids, order, starts, rowParent            []int
	pairTotals                               []float64
	parents, noAnswer, eligible, pairs, zero int
}

func finiteRanking(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func reserveRanking(total *uint64, n, width uint64) error {
	if *total > MaxRankingPayloadBytes || width != 0 && n > (MaxRankingPayloadBytes-*total)/width {
		return fmt.Errorf("ranking payload exceeds 64 MiB")
	}
	*total += n * width
	return nil
}

// IDs may be sparse. Sorting row indices avoids a map or allocation indexed by
// arbitrary IDs. Owned ID copies and generated columns outlive no mutable caller
// metadata. At most eight complete candidate rows belong to any parent.
func bindAssociation(d Dataset, ids []int, unitWeights bool, total *uint64) (association, error) {
	var a association
	n := len(d.Labels)
	if len(ids) != n || n == 0 {
		return a, fmt.Errorf("ranking parent association length")
	}
	// Conservative preflight covers IDs/order/starts/rowParent/pairTotals with
	// eight-byte ints on both supported platforms, before any new allocation.
	if err := reserveRanking(total, uint64(n)+1, 40); err != nil {
		return a, err
	}
	a.ids, a.order = slices.Clone(ids), make([]int, n)
	a.starts, a.rowParent, a.pairTotals = make([]int, 0, n+1), make([]int, n), make([]float64, 0, n)
	for i, id := range a.ids {
		if id < 0 || d.Groups[i] < 0 || unitWeights && d.SampleWeights != nil && d.SampleWeights[i] != 0 && d.SampleWeights[i] != 1 {
			return association{}, fmt.Errorf("ranking requires nonnegative IDs/groups and unit or zero weights")
		}
		a.order[i] = i
	}
	slices.SortFunc(a.order, func(i, j int) int {
		if a.ids[i] < a.ids[j] {
			return -1
		}
		if a.ids[i] > a.ids[j] {
			return 1
		}
		if i < j {
			return -1
		}
		if i > j {
			return 1
		}
		return 0
	})
	for start := 0; start < n; {
		end := start + 1
		for end < n && a.ids[a.order[end]] == a.ids[a.order[start]] {
			end++
		}
		if end-start > MaxRankingParentRows {
			return association{}, fmt.Errorf("ranking parent exceeds eight rows")
		}
		a.starts = append(a.starts, start)
		positive, negative, positiveWeighted, negativeWeighted, pweight, nweight := 0, 0, 0, 0, 0., 0.
		for _, row := range a.order[start:end] {
			if d.Groups[row] != d.Groups[a.order[start]] {
				return association{}, fmt.Errorf("ranking parent crosses source groups")
			}
			a.rowParent[row] = a.parents
			weight := 1.
			if d.SampleWeights != nil {
				weight = d.SampleWeights[row]
			}
			if d.Labels[row] == 1 {
				positive++
				pweight += weight
				if weight > 0 {
					positiveWeighted++
				}
			} else {
				negative++
				nweight += weight
				if weight > 0 {
					negativeWeighted++
				}
			}
		}
		if positive == 0 {
			a.noAnswer++
		}
		pairTotal := pweight * nweight
		if pairTotal > 0 {
			a.eligible++
		}
		a.pairTotals = append(a.pairTotals, pairTotal)
		a.pairs += positive * negative
		a.zero += positive*negative - positiveWeighted*negativeWeighted
		a.parents++
		start = end
	}
	a.starts = append(a.starts, n)
	return a, nil
}

func planPairs(d Dataset, a association, total *uint64) (pairPlan, error) {
	var p pairPlan
	if err := reserveRanking(total, uint64(a.pairs), 32); err != nil {
		return p, err
	}
	if err := reserveRanking(total, uint64(len(d.Labels))+1, 8); err != nil {
		return p, err
	}
	p.audit = PairAudit{Rows: len(d.Labels), Parents: a.parents, NoAnswerParents: a.noAnswer, EligibleParents: a.eligible, ZeroWeightPairs: a.zero,
		ParentIDs: make([]int, a.pairs), PositiveRows: make([]int, a.pairs), NegativeRows: make([]int, a.pairs), Weights: make([]float64, a.pairs), PayloadBytes: uint64(a.pairs)*32 + uint64(len(d.Labels)+1)*8}
	p.offsets = make([]int, len(d.Labels)+1)
	index := 0
	for row, label := range d.Labels {
		p.offsets[row] = index
		if label != 1 {
			continue
		}
		parent := a.rowParent[row]
		for _, neg := range a.order[a.starts[parent]:a.starts[parent+1]] {
			if d.Labels[neg] != 0 {
				continue
			}
			weight := 1.
			if d.SampleWeights != nil {
				weight = d.SampleWeights[row] * d.SampleWeights[neg]
			}
			if weight != 0 {
				weight /= a.pairTotals[parent] * float64(a.eligible)
			}
			p.audit.ParentIDs[index], p.audit.PositiveRows[index], p.audit.NegativeRows[index], p.audit.Weights[index] = a.ids[row], row, neg, weight
			index++
		}
	}
	p.offsets[len(d.Labels)] = index
	return p, nil
}

func prepareRanking(train, validation Dataset, c Config, cfg *RankingConfig) (*rankingObjective, RankingTrace, error) {
	rt := RankingTrace{Schema: RankingSchema}
	if cfg.Schema != RankingSchema || cfg.Lambda != 0 && cfg.Lambda != 1 {
		return nil, rt, fmt.Errorf("invalid ranking schema or lambda")
	}
	dimension := c.Dimension
	if dimension == 0 {
		dimension = hintlearn.Dimension
	}
	if dimension < 1 || dimension > hintlearn.Dimension || train.Split != "development" || validation.Split != "validation" {
		return nil, rt, fmt.Errorf("invalid ranking dataset split or dimension")
	}
	for _, d := range []Dataset{train, validation} {
		if err := validate(d, dimension); err != nil {
			return nil, rt, err
		}
	}
	var total uint64
	tr, err := bindAssociation(train, cfg.TrainingParentIDs, cfg.Lambda != 0, &total)
	if err != nil {
		return nil, rt, err
	}
	va, err := bindAssociation(validation, cfg.ValidationParentIDs, cfg.Lambda != 0, &total)
	if err != nil {
		return nil, rt, err
	}
	for i, j := 0, 0; i < tr.parents && j < va.parents; {
		tid, vid := tr.ids[tr.order[tr.starts[i]]], va.ids[va.order[va.starts[j]]]
		if tid == vid {
			return nil, rt, fmt.Errorf("ranking parent crosses roles")
		}
		if tid < vid {
			i++
		} else {
			j++
		}
	}
	if cfg.Lambda == 0 {
		// No pair plan/normalization is evaluated in the disabled path. Valid
		// legacy non-unit BCE weights are therefore preserved exactly as well.
		rt.Training = PairAudit{Rows: len(train.Labels), Parents: tr.parents, NoAnswerParents: tr.noAnswer}
		rt.Validation = PairAudit{Rows: len(validation.Labels), Parents: va.parents, NoAnswerParents: va.noAnswer}
		return nil, rt, nil
	}
	tp, err := planPairs(train, tr, &total)
	if err != nil {
		return nil, rt, err
	}
	vp, err := planPairs(validation, va, &total)
	if err != nil {
		return nil, rt, err
	}
	if c.Epochs < 1 || c.Epochs > 100 {
		return nil, rt, fmt.Errorf("invalid ranking epoch count")
	}
	if err := reserveRanking(&total, uint64(c.Epochs), 48); err != nil {
		return nil, rt, err
	}
	rt.Training, rt.Validation, rt.Lambda = tp.audit, vp.audit, cfg.Lambda
	rt.Enabled = true
	rt.Epochs = make([]RankingEpoch, 0, c.Epochs)
	return &rankingObjective{lambda: cfg.Lambda, training: tp, validation: vp}, rt, nil
}

func (p pairPlan) addGradient(d Dataset, w []float64, batch []int, lambda float64, grad []float64) error {
	for _, row := range batch {
		for i := p.offsets[row]; i < p.offsets[row+1]; i++ {
			if p.audit.Weights[i] == 0 {
				continue
			}
			pos, neg := p.audit.PositiveRows[i], p.audit.NegativeRows[i]
			delta := d.score(w, pos) - d.score(w, neg)
			if !finiteRanking(delta) {
				return fmt.Errorf("nonfinite ranking margin")
			}
			factor := -lambda * float64(len(d.Labels)) * p.audit.Weights[i] * logistic(-delta)
			for k := d.Offsets[pos]; k < d.Offsets[pos+1]; k++ {
				grad[d.Indices[k]] += factor * d.Values[k]
			}
			for k := d.Offsets[neg]; k < d.Offsets[neg+1]; k++ {
				grad[d.Indices[k]] -= factor * d.Values[k]
			}
		}
	}
	for _, v := range grad {
		if !finiteRanking(v) {
			return fmt.Errorf("nonfinite ranking gradient")
		}
	}
	return nil
}

func (p pairPlan) nll(d Dataset, w []float64) (RankingLoss, error) {
	if p.audit.EligibleParents == 0 {
		return RankingLoss{}, nil
	}
	v := 0.
	for i, weight := range p.audit.Weights {
		if weight == 0 {
			continue
		}
		margin := d.score(w, p.audit.PositiveRows[i]) - d.score(w, p.audit.NegativeRows[i])
		if !finiteRanking(margin) {
			return RankingLoss{}, fmt.Errorf("nonfinite ranking margin")
		}
		v += weight * loss(margin, 1)
	}
	if !finiteRanking(v) {
		return RankingLoss{}, fmt.Errorf("nonfinite ranking nll")
	}
	return RankingLoss{Defined: true, Value: v}, nil
}

func (t *RankingTrace) fail(epoch int, kind string) {
	if t.FailureKind == "" {
		t.FailureEpoch, t.FailureKind = epoch, kind
	}
}

func (t *RankingTrace) record(epoch int, train, validation Dataset, w []float64, r *rankingObjective) error {
	tr, err := r.training.nll(train, w)
	if err != nil {
		t.fail(epoch, "nonfinite_ranking_nll")
		return err
	}
	va, err := r.validation.nll(validation, w)
	if err != nil {
		t.fail(epoch, "nonfinite_ranking_nll")
		return err
	}
	t.Epochs = append(t.Epochs, RankingEpoch{Epoch: epoch, Training: tr, Validation: va})
	return nil
}
