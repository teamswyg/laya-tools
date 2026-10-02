package pairlearn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"
)

func rankingToy() (Dataset, Dataset, RankingConfig) {
	train := Dataset{Split: "development", Offsets: []int{0, 2, 4, 6, 8, 10, 12, 14, 16},
		Indices: []uint16{0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
		Values:  []float64{1, .2, -.5, 1, 3, -2, -1, .4, .3, -.7, .5, 1, 2, -1, -2, .1},
		Labels:  []float64{1, 0, 0, 0, 0, 1, 1, 0}, Groups: []int{10, 10, 10, 11, 11, 12, 12, 12},
		SampleWeights: []float64{1, 1, 0, 1, 1, 0, 1, 1}}
	val := Dataset{Split: "validation", Offsets: []int{0, 2, 4, 6, 8}, Indices: []uint16{0, 1, 0, 1, 0, 1, 0, 1},
		Values: []float64{.5, .2, -1, .1, -.7, .4, -.2, .3}, Labels: []float64{1, 0, 0, 0}, Groups: []int{20, 20, 21, 21}, SampleWeights: []float64{1, 1, 1, 1}}
	cfg := RankingConfig{Schema: RankingSchema, Lambda: 1, TrainingParentIDs: []int{100, 100, 100, 101, 101, 102, 102, 102}, ValidationParentIDs: []int{200, 200, 201, 201}}
	return train, val, cfg
}

func rankingToyConfig() Config {
	return Config{Seed: 1729, Mode: "fp32", LearningRate: .1, L2: .0001, Epochs: 6, Batch: 3, Dimension: 2}
}

func TestRankingNilAndZeroExactLegacyParity(t *testing.T) {
	for _, mode := range []string{"fp32", "ternary_ste"} {
		for _, seed := range []uint64{1729, 2718} {
			for _, unit := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s_nilweights%v_seed%d", mode, unit, seed), func(t *testing.T) {
					tr, va, cfg := rankingToy()
					cfg.Lambda = 0
					if unit {
						tr.SampleWeights, va.SampleWeights = nil, nil
					}
					c := rankingToyConfig()
					c.Mode, c.Seed = mode, seed
					beforeTr, beforeVa := cloneDataset(tr), cloneDataset(va)
					plain, err := Fit(tr, va, c)
					if err != nil {
						t.Fatal(err)
					}
					old, oldTrace, err := FitWithTrace(tr, va, c)
					if err != nil || !reflect.DeepEqual(old, plain) {
						t.Fatal(err)
					}
					nilResult, err := FitWithRanking(tr, va, c, nil)
					if err != nil || !reflect.DeepEqual(nilResult, plain) {
						t.Fatal("nil changed Fit", err)
					}
					nilResult, nilTrace, disabled, err := FitWithRankingTrace(tr, va, c, nil)
					if err != nil || !reflect.DeepEqual(nilResult, plain) || !reflect.DeepEqual(nilTrace, oldTrace) || disabled.Enabled {
						t.Fatal("nil changed trace", err)
					}
					zero, err := FitWithRanking(tr, va, c, &cfg)
					if err != nil || !reflect.DeepEqual(zero, plain) {
						t.Fatal("lambda zero changed Fit", err)
					}
					zero, zeroTrace, rankTrace, err := FitWithRankingTrace(tr, va, c, &cfg)
					if err != nil || !reflect.DeepEqual(zero, plain) || !reflect.DeepEqual(zeroTrace, oldTrace) || rankTrace.Enabled || len(rankTrace.Epochs) != 0 || !rankTrace.Complete {
						t.Fatal("lambda zero changed trace", err)
					}
					if !reflect.DeepEqual(tr, beforeTr) || !reflect.DeepEqual(va, beforeVa) {
						t.Fatal("mutated caller dataset")
					}
				})
			}
		}
	}
}

func TestRankingPairAuditRetainsMaskAndNoAnswer(t *testing.T) {
	tr, va, cfg := rankingToy()
	beforeTr, beforeVa := cloneDataset(tr), cloneDataset(va)
	beforeIDs := slices.Clone(cfg.TrainingParentIDs)
	r, bce, rt, err := FitWithRankingTrace(tr, va, rankingToyConfig(), &cfg)
	if err != nil || !rt.Enabled || !rt.Complete || !bce.Complete || len(rt.Epochs) != 6 || rt.SelectedEpoch != r.Epoch {
		t.Fatal(err, rt)
	}
	a := rt.Training
	if a.Rows != 8 || a.Parents != 3 || a.NoAnswerParents != 1 || a.EligibleParents != 2 || a.ZeroWeightPairs != 2 ||
		!slices.Equal(a.ParentIDs, []int{100, 100, 102, 102}) || !slices.Equal(a.PositiveRows, []int{0, 0, 5, 6}) || !slices.Equal(a.NegativeRows, []int{1, 2, 7, 7}) || !slices.Equal(a.Weights, []float64{.5, 0, 0, .5}) {
		t.Fatal("pair scope/normalization/mask lost", a)
	}
	if bce.Training.Rows != 8 || bce.Training.ZeroWeightRows != 2 || bce.Training.SampleWeightSum != 6 || rt.Validation.NoAnswerParents != 1 {
		t.Fatal("original BCE/no-answer coverage changed")
	}
	if !reflect.DeepEqual(tr, beforeTr) || !reflect.DeepEqual(va, beforeVa) || !slices.Equal(cfg.TrainingParentIDs, beforeIDs) {
		t.Fatal("input mutation")
	}
	for i, e := range rt.Epochs {
		if e.Epoch != i+1 || !e.Training.Defined || !e.Validation.Defined || !finite(e.Training.Value) || !finite(e.Validation.Value) {
			t.Fatal("invalid diagnostic", e)
		}
	}
	clear(cfg.TrainingParentIDs)
	clear(tr.Groups)
	clear(tr.Labels)
	if !slices.Equal(a.ParentIDs, []int{100, 100, 102, 102}) || bce.Training.Rows != 8 {
		t.Fatal("audit retained mutable caller metadata")
	}
	b, err := json.Marshal(struct {
		BCE     Trace
		Ranking RankingTrace
		Result  Result
	}{bce, rt, r})
	if err != nil {
		t.Fatal(err)
	}
	var payload struct{ Result json.RawMessage }
	if err := json.Unmarshal(b, &payload); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(payload.Result, []byte(`"Weights"`)) || bytes.Contains(b, []byte(`"Indices"`)) || bytes.Contains(b, []byte(`"Values"`)) {
		t.Fatal("unexpected model coefficient or sparse feature payload")
	}
}

// Independent scalar softplus expression and central differences test the
// continuous logistic surrogate at fixed scorer coefficients. This does not
// claim a finite-difference derivative through discontinuous FP32 rounding.
func independentPairLoss(d Dataset, p pairPlan, w []float64) float64 {
	score := func(row int) float64 {
		v := 0.
		for k := d.Offsets[row]; k < d.Offsets[row+1]; k++ {
			v += w[d.Indices[k]] * d.Values[k]
		}
		return v
	}
	v := 0.
	for i, weight := range p.audit.Weights {
		if weight != 0 {
			v += weight * math.Log1p(math.Exp(score(p.audit.NegativeRows[i])-score(p.audit.PositiveRows[i])))
		}
	}
	return v
}

func independentBCELoss(d Dataset, w []float64) float64 {
	v, total := 0., 0.
	for row, y := range d.Labels {
		weight := 1.
		if d.SampleWeights != nil {
			weight = d.SampleWeights[row]
		}
		if weight == 0 {
			continue
		}
		s := 0.
		for k := d.Offsets[row]; k < d.Offsets[row+1]; k++ {
			s += w[d.Indices[k]] * d.Values[k]
		}
		v += weight * (math.Log1p(math.Exp(s)) - y*s)
		total += weight
	}
	return v / total
}

func TestRankingGradientFiniteDifferences(t *testing.T) {
	tr, va, cfg := rankingToy()
	objective, _, err := prepareRanking(tr, va, rankingToyConfig(), &cfg)
	if err != nil {
		t.Fatal(err)
	}
	w := []float64{.23, -.31}
	grad := make([]float64, 2)
	rows := []int{0, 1, 2, 3, 4, 5, 6, 7}
	if err := objective.training.addGradient(tr, w, rows, 1, grad); err != nil {
		t.Fatal(err)
	}
	for i := range w {
		a, b := slices.Clone(w), slices.Clone(w)
		a[i] += 1e-6
		b[i] -= 1e-6
		want := (independentPairLoss(tr, objective.training, a) - independentPairLoss(tr, objective.training, b)) / 2e-6
		if math.Abs(grad[i]/8-want) > 1e-9 {
			t.Fatalf("gradient %d got %.15g want %.15g", i, grad[i]/8, want)
		}
	}
	got, err := objective.training.nll(tr, w)
	if err != nil || !got.Defined || math.Abs(got.Value-independentPairLoss(tr, objective.training, w)) > 1e-15 {
		t.Fatal("pair loss mismatch", got, err)
	}
}

func TestRankingJointBCEGradientFiniteDifferences(t *testing.T) {
	tr, va, cfg := rankingToy()
	objective, _, err := prepareRanking(tr, va, rankingToyConfig(), &cfg)
	if err != nil {
		t.Fatal(err)
	}
	w := []float64{.23, -.31}
	grad := make([]float64, 2)
	for row, y := range tr.Labels {
		s := tr.score(w, row)
		delta := (1/(1+math.Exp(-s)) - y) * tr.SampleWeights[row] * 8 / 6
		for k := tr.Offsets[row]; k < tr.Offsets[row+1]; k++ {
			grad[tr.Indices[k]] += delta * tr.Values[k]
		}
	}
	if err := objective.training.addGradient(tr, w, []int{0, 1, 2, 3, 4, 5, 6, 7}, 1, grad); err != nil {
		t.Fatal(err)
	}
	for i := range w {
		a, b := slices.Clone(w), slices.Clone(w)
		a[i] += 1e-6
		b[i] -= 1e-6
		want := (independentBCELoss(tr, a) + independentPairLoss(tr, objective.training, a) - independentBCELoss(tr, b) - independentPairLoss(tr, objective.training, b)) / 2e-6
		if math.Abs(grad[i]/8-want) > 1e-9 {
			t.Fatal("joint objective derivative differs", grad[i]/8, want)
		}
	}
}

func TestRankingAveragesPairsThenParents(t *testing.T) {
	tr := Dataset{Split: "development", Offsets: []int{0, 1, 2, 3, 4, 5, 6}, Indices: []uint16{0, 0, 0, 0, 0, 0}, Values: []float64{1, -1, -2, 2, 3, -3}, Labels: []float64{1, 0, 0, 1, 1, 0}, Groups: []int{1, 1, 1, 2, 2, 2}, SampleWeights: []float64{1, 1, 1, 1, 0, 1}}
	va := Dataset{Split: "validation", Offsets: []int{0, 1, 2}, Indices: []uint16{0, 0}, Values: []float64{1, -1}, Labels: []float64{1, 0}, Groups: []int{3, 3}}
	cfg := RankingConfig{Schema: RankingSchema, Lambda: 1, TrainingParentIDs: []int{0, 0, 0, 1, 1, 1}, ValidationParentIDs: []int{2, 2}}
	c := rankingToyConfig()
	c.Dimension = 1
	p, _, err := prepareRanking(tr, va, c, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(p.training.audit.PositiveRows, []int{0, 0, 3, 4}) || !slices.Equal(p.training.audit.NegativeRows, []int{1, 2, 5, 5}) || !slices.Equal(p.training.audit.Weights, []float64{.25, .25, .5, 0}) {
		t.Fatal("pair/parent average changed", p.training.audit)
	}
	got, err := p.training.nll(tr, []float64{.2})
	want := .25*math.Log1p(math.Exp(-.4)) + .25*math.Log1p(math.Exp(-.6)) + .5*math.Log1p(math.Exp(-1))
	if err != nil || !got.Defined || math.Abs(got.Value-want) > 1e-15 {
		t.Fatal("literal weighted loss differs", got, want, err)
	}
}

func TestRankingPositiveOwnerUnbiasedMiniBatches(t *testing.T) {
	tr, va, cfg := rankingToy()
	objective, _, err := prepareRanking(tr, va, rankingToyConfig(), &cfg)
	if err != nil {
		t.Fatal(err)
	}
	w := []float64{.4, -.2}
	full := make([]float64, 2)
	if err := objective.training.addGradient(tr, w, []int{0, 1, 2, 3, 4, 5, 6, 7}, 1, full); err != nil {
		t.Fatal(err)
	}
	// Exhaust every size-two subset. The negative endpoint is often absent from
	// the batch; it must still be scored and receive its opposite contribution.
	average := make([]float64, 2)
	batches := 0
	for i := 0; i < 8; i++ {
		for j := i + 1; j < 8; j++ {
			g := make([]float64, 2)
			if err := objective.training.addGradient(tr, w, []int{i, j}, 1, g); err != nil {
				t.Fatal(err)
			}
			for k := range g {
				average[k] += g[k] / 2
			}
			batches++
		}
	}
	for k := range full {
		if math.Abs(average[k]/float64(batches)-full[k]/8) > 1e-14 {
			t.Fatal("biased minibatch gradient", average, full)
		}
	}
}

func TestRankingOneEpochUsesSameForwardAndUnevenBatchScale(t *testing.T) {
	tr, va, cfg := rankingToy()
	c := rankingToyConfig()
	c.Epochs = 1
	objective, _, err := prepareRanking(tr, va, c, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewPCG(c.Seed, c.Seed+1))
	shadow := []float64{rng.NormFloat64() * .01, rng.NormFloat64() * .01}
	order := []int{0, 1, 2, 3, 4, 5, 6, 7}
	rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	for start := 0; start < 8; start += c.Batch {
		batch := order[start:min(start+c.Batch, 8)]
		qw := []float64{float64(float32(shadow[0])), float64(float32(shadow[1]))}
		g := make([]float64, 2)
		score := func(row int) float64 { return qw[0]*tr.Values[2*row] + qw[1]*tr.Values[2*row+1] }
		for _, row := range batch {
			delta := (1/(1+math.Exp(-score(row))) - tr.Labels[row]) * tr.SampleWeights[row] * 8 / 6
			for k := 0; k < 2; k++ {
				g[k] += delta * tr.Values[2*row+k]
			}
			for j, pos := range objective.training.audit.PositiveRows {
				if pos != row || objective.training.audit.Weights[j] == 0 {
					continue
				}
				neg := objective.training.audit.NegativeRows[j]
				margin := score(pos) - score(neg)
				factor := -8 * objective.training.audit.Weights[j] / (1 + math.Exp(margin))
				for k := 0; k < 2; k++ {
					g[k] += factor * (tr.Values[2*pos+k] - tr.Values[2*neg+k])
				}
			}
		}
		for k := range shadow {
			shadow[k] -= c.LearningRate * (g[k]/float64(len(batch)) + c.L2*shadow[k])
		}
	}
	want := []float64{float64(float32(shadow[0])), float64(float32(shadow[1]))}
	got, _, _, err := FitWithRankingTrace(tr, va, c, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	for k := range want {
		if math.Abs(got.Weights[k]-want[k]) > 1e-9 {
			t.Fatal("forward/batch scaling differs", got.Weights, want)
		}
	}
}

func TestRankingMaskedEndpointHasNoGradientEffect(t *testing.T) {
	tr, va, cfg := rankingToy()
	c := rankingToyConfig()
	first, err := FitWithRanking(tr, va, c, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Both a masked negative and a masked positive remain physical rows. Change
	// their finite feature values drastically; do not delete rows or their pairs.
	tr.Values[4], tr.Values[5], tr.Values[10], tr.Values[11] = 1e150, -1e150, -1e150, 1e150
	second, err := FitWithRanking(tr, va, c, &cfg)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatal("masked endpoint influenced fit", err, first, second)
	}
}

func TestRankingNoAnswerRowsKeepBCEAndUndefinedPairs(t *testing.T) {
	tr, va, cfg := rankingToy()
	clear(tr.Labels)
	clear(va.Labels)
	base, err := Fit(tr, va, rankingToyConfig())
	if err != nil {
		t.Fatal(err)
	}
	got, bce, rt, err := FitWithRankingTrace(tr, va, rankingToyConfig(), &cfg)
	if err != nil || !reflect.DeepEqual(base, got) || rt.Training.NoAnswerParents != 3 || rt.Validation.NoAnswerParents != 2 || len(rt.Training.Weights) != 0 || bce.Training.SampleWeightSum != 6 {
		t.Fatal("no-answer BCE changed", err)
	}
	for _, e := range rt.Epochs {
		if e.Training.Defined || e.Validation.Defined {
			t.Fatal("fabricated no-answer ordering loss", e)
		}
	}
}

func TestRankingKeepsEarliestValidationBCEMinimum(t *testing.T) {
	tr, va, cfg := rankingToy()
	clear(tr.Values)
	clear(va.Values)
	r, bce, rt, err := FitWithRankingTrace(tr, va, rankingToyConfig(), &cfg)
	if err != nil || r.Epoch != 1 || rt.SelectedEpoch != 1 || bce.SelectedEpoch != 1 {
		t.Fatal("selector changed", err)
	}
	for _, e := range rt.Epochs {
		if math.Abs(e.Training.Value-math.Log(2)) > 1e-15 || math.Abs(e.Validation.Value-math.Log(2)) > 1e-15 {
			t.Fatal("wrong zero-margin loss", e)
		}
	}
}

func TestRankingRejectsInvalidOptInWithoutFallback(t *testing.T) {
	for _, kind := range []string{"schema", "nan_lambda", "fraction_lambda", "ids_length", "negative_id", "cross_group", "cross_parent_role", "calibration", "unknown_label", "fraction_weight", "row_bound", "offset", "feature"} {
		t.Run(kind, func(t *testing.T) {
			tr, va, cfg := rankingToy()
			c := rankingToyConfig()
			switch kind {
			case "schema":
				cfg.Schema = "wrong"
			case "nan_lambda":
				cfg.Lambda = math.NaN()
			case "fraction_lambda":
				cfg.Lambda = .5
			case "ids_length":
				cfg.TrainingParentIDs = cfg.TrainingParentIDs[:7]
			case "negative_id":
				cfg.TrainingParentIDs[0] = -1
			case "cross_group":
				tr.Groups[1] = 999
			case "cross_parent_role":
				cfg.ValidationParentIDs[0], cfg.ValidationParentIDs[1] = 100, 100
			case "calibration":
				va.Split = "calibration"
			case "unknown_label":
				tr.Labels[0] = math.NaN()
			case "fraction_weight":
				tr.SampleWeights[0] = .5
			case "row_bound":
				tr = Dataset{Split: "development", Offsets: make([]int, 10), Labels: make([]float64, 9), Groups: make([]int, 9)}
				cfg.TrainingParentIDs = make([]int, 9)
			case "offset":
				tr.Offsets[2] = 100
			case "feature":
				tr.Values[0] = math.Inf(1)
			}
			r, bce, rt, err := FitWithRankingTrace(tr, va, c, &cfg)
			if err == nil || len(r.Weights) != 0 || bce.Complete || rt.Complete || len(bce.Epochs) != 0 || len(rt.Epochs) != 0 {
				t.Fatal("invalid opt-in entered trainer/fallback", err)
			}
		})
	}
	tr, va, cfg := rankingToy()
	va.Groups[0], va.Groups[1] = 10, 10
	if _, err := FitWithRanking(tr, va, rankingToyConfig(), &cfg); err == nil {
		t.Fatal("source-group role leakage accepted")
	}
}

func TestRankingResourceAndNumericalBounds(t *testing.T) {
	var total uint64
	if err := reserveRanking(&total, MaxRankingPayloadBytes, 1); err != nil || total != MaxRankingPayloadBytes {
		t.Fatal(err)
	}
	if err := reserveRanking(&total, 1, 1); err == nil {
		t.Fatal("cap bypass")
	}
	total = 0
	if err := reserveRanking(&total, ^uint64(0), 8); err == nil || total != 0 {
		t.Fatal("overflow bypass")
	}
	total = MaxRankingPayloadBytes - 39
	if _, err := bindAssociation(Dataset{Labels: make([]float64, 2)}, []int{0, 0}, true, &total); err == nil {
		t.Fatal("association allocated beyond preflight")
	}
	tr, va, cfg := rankingToy()
	objective, _, err := prepareRanking(tr, va, rankingToyConfig(), &cfg)
	if err != nil {
		t.Fatal(err)
	}
	tr.Values[0], tr.Values[2] = math.MaxFloat64, -math.MaxFloat64
	if _, err := objective.training.nll(tr, []float64{1, 0}); err == nil {
		t.Fatal("nonfinite margin accepted")
	}
	if err := objective.training.addGradient(tr, []float64{1, 0}, []int{0}, 1, make([]float64, 2)); err == nil {
		t.Fatal("nonfinite gradient accepted")
	}
}

func TestRankingReproducibilityAndAuditOwnership(t *testing.T) {
	tr, va, cfg := rankingToy()
	c := rankingToyConfig()
	a, at, ar, err := FitWithRankingTrace(tr, va, c, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, bt, br, err := FitWithRankingTrace(tr, va, c, &cfg)
	if err != nil || !reflect.DeepEqual(a, b) || !reflect.DeepEqual(at, bt) || !reflect.DeepEqual(ar, br) {
		t.Fatal("not reproducible", err)
	}
	ar.Training.Weights[0] = 999
	ar.Epochs[0].Training.Value = 999
	if br.Training.Weights[0] == 999 || br.Epochs[0].Training.Value == 999 || a.ValidationNLL == 999 {
		t.Fatal("shared mutable output backing")
	}
}

func TestRankingParentNamesAreAssociationOnly(t *testing.T) {
	tr, va, cfg := rankingToy()
	first, err := FitWithRanking(tr, va, rankingToyConfig(), &cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Sparse names and reversed numeric ordering cannot change features, row
	// shuffling, pair enumeration within a parent, or model coefficients.
	for i, id := range cfg.TrainingParentIDs {
		switch id {
		case 100:
			cfg.TrainingParentIDs[i] = int(^uint(0) >> 1)
		case 101:
			cfg.TrainingParentIDs[i] = 2
		case 102:
			cfg.TrainingParentIDs[i] = 1
		}
	}
	second, err := FitWithRanking(tr, va, rankingToyConfig(), &cfg)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatal("association ID entered model", err)
	}
}

func TestRankingZeroPreservesArbitraryLegacyBCEWeights(t *testing.T) {
	tr, va, cfg := rankingToy()
	cfg.Lambda = 0
	tr.SampleWeights = []float64{2, .25, 0, 3, 1, 0, .5, 1}
	va.SampleWeights = []float64{.25, 2, 0, 1}
	base, baseTrace, err := FitWithTrace(tr, va, rankingToyConfig())
	if err != nil {
		t.Fatal(err)
	}
	got, trace, ranking, err := FitWithRankingTrace(tr, va, rankingToyConfig(), &cfg)
	if err != nil || !reflect.DeepEqual(got, base) || !reflect.DeepEqual(trace, baseTrace) || ranking.Enabled || len(ranking.Training.Weights) != 0 || len(ranking.Epochs) != 0 {
		t.Fatal("lambda zero changed legacy weights", err)
	}
}
