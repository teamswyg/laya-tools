package statehintwide

import (
	"math"
	"math/rand/v2"
)

type Sample struct {
	Text  string `json:"text"`
	Label Intent `json:"label"`
}

// FitOptions configures mini-batch cross-entropy and AdamW. Zero values select
// documented defaults. Each Fit starts a fresh optimizer on existing weights;
// that is supervised warm-start fine-tuning, not optimizer-state resumption.
type FitOptions struct {
	Epochs       int
	BatchSize    int
	LearningRate float64
	WeightDecay  float64
	Seed         int64
}

type FitReport struct {
	Samples       int     `json:"samples"`
	Epochs        int     `json:"epochs"`
	Batches       int     `json:"batches"`
	TrainingSteps uint64  `json:"training_steps"`
	MeanLoss      float64 `json:"mean_loss"`
}

type adamWorkspace struct {
	gradient     [FeatureBins][IntentCount]float64
	first        [FeatureBins][IntentCount]float64
	second       [FeatureBins][IntentCount]float64
	biasGradient [IntentCount]float64
	biasFirst    [IntentCount]float64
	biasSecond   [IntentCount]float64
}

func (options FitOptions) normalized() (FitOptions, error) {
	if options.Epochs == 0 {
		options.Epochs = 20
	}
	if options.BatchSize == 0 {
		options.BatchSize = 32
	}
	if options.LearningRate == 0 {
		options.LearningRate = 0.02
	}
	if !finite(options.LearningRate) || options.LearningRate <= 0 || options.LearningRate > 1 ||
		!finite(options.WeightDecay) || options.WeightDecay < 0 || options.WeightDecay > 1 ||
		options.Epochs < 1 || options.Epochs > 10000 || options.BatchSize < 1 || options.BatchSize > 4096 {
		return FitOptions{}, ErrTraining
	}
	return options, nil
}

// Fit updates this exclusively owned model. Inputs are validated before the
// first update. Temperature is reset to one; calibration belongs to validation,
// independently of the training objective. Fit never selects a model using test
// data and makes no claims about the provenance of caller-provided examples.
func (m *Model) Fit(samples []Sample, options FitOptions) (FitReport, error) {
	if !m.valid() || len(samples) == 0 || len(samples) > 1_000_000 {
		return FitReport{}, ErrTraining
	}
	options, err := options.normalized()
	if err != nil {
		return FitReport{}, err
	}
	var features Workspace
	for _, sample := range samples {
		if _, ok := IntentIndex(sample.Label); !ok {
			return FitReport{}, ErrTraining
		}
		if err := extract(sample.Text, &features, m.mode); err != nil {
			return FitReport{}, ErrTraining
		}
	}
	batches := (len(samples) + options.BatchSize - 1) / options.BatchSize
	if uint64(batches)*uint64(options.Epochs) > ^uint64(0)-m.steps {
		return FitReport{}, ErrTraining
	}
	order := make([]int, len(samples))
	for index := range order {
		order[index] = index
	}
	random := rand.New(rand.NewPCG(uint64(options.Seed), uint64(options.Seed)^0x9e3779b97f4a7c15))
	optimizer := &adamWorkspace{}
	report := FitReport{Samples: len(samples), Epochs: options.Epochs}
	m.temperature = 1
	var lossSum float64
	var updates uint64
	for epoch := 0; epoch < options.Epochs; epoch++ {
		random.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for start := 0; start < len(order); start += options.BatchSize {
			end := min(start+options.BatchSize, len(order))
			optimizer.gradient = [FeatureBins][IntentCount]float64{}
			optimizer.biasGradient = [IntentCount]float64{}
			for _, sampleIndex := range order[start:end] {
				sample := samples[sampleIndex]
				// The complete input validation pass above fixes this contract.
				if err := extract(sample.Text, &features, m.mode); err != nil {
					return FitReport{}, ErrTraining
				}
				logits := m.logits(&features)
				probabilities, ok := softmax(logits, 1)
				if !ok {
					return FitReport{}, ErrTraining
				}
				target, _ := IntentIndex(sample.Label)
				// Stable log-sum-exp retains finite loss for very small posteriors.
				maxLogit := logits[0]
				for _, value := range logits {
					if value > maxLogit {
						maxLogit = value
					}
				}
				var total float64
				for _, value := range logits {
					total += math.Exp(value - maxLogit)
				}
				lossSum += maxLogit + math.Log(total) - logits[target]
				for class, probability := range probabilities {
					gradient := probability
					if class == target {
						gradient--
					}
					optimizer.biasGradient[class] += gradient
					for _, index := range features.indices[:features.count] {
						optimizer.gradient[index][class] += gradient * float64(features.values[index])
					}
				}
			}
			updates++
			if err := optimizer.update(m, options, end-start, updates); err != nil {
				return FitReport{}, err
			}
			m.steps++
			report.Batches++
		}
	}
	report.TrainingSteps = m.steps
	report.MeanLoss = lossSum / (float64(len(samples)) * float64(options.Epochs))
	return report, nil
}

func (a *adamWorkspace) update(m *Model, options FitOptions, batch int, step uint64) error {
	const beta1, beta2, epsilon = 0.9, 0.999, 1e-8
	correction1 := 1 - math.Pow(beta1, float64(step))
	correction2 := 1 - math.Pow(beta2, float64(step))
	scale := 1 / float64(batch)
	decay := 1 - options.LearningRate*options.WeightDecay
	for index := range m.weights {
		for class := range m.weights[index] {
			gradient := a.gradient[index][class] * scale
			a.first[index][class] = beta1*a.first[index][class] + (1-beta1)*gradient
			a.second[index][class] = beta2*a.second[index][class] + (1-beta2)*gradient*gradient
			update := (a.first[index][class] / correction1) / (math.Sqrt(a.second[index][class]/correction2) + epsilon)
			value := float64(m.weights[index][class])*decay - options.LearningRate*update
			if !finite(value) || math.Abs(value) > math.MaxFloat32 {
				return ErrTraining
			}
			m.weights[index][class] = float32(value)
		}
	}
	// Biases are not decayed, as in the usual AdamW classifier convention.
	for class := range m.bias {
		gradient := a.biasGradient[class] * scale
		a.biasFirst[class] = beta1*a.biasFirst[class] + (1-beta1)*gradient
		a.biasSecond[class] = beta2*a.biasSecond[class] + (1-beta2)*gradient*gradient
		update := (a.biasFirst[class] / correction1) / (math.Sqrt(a.biasSecond[class]/correction2) + epsilon)
		value := float64(m.bias[class]) - options.LearningRate*update
		if !finite(value) || math.Abs(value) > math.MaxFloat32 {
			return ErrTraining
		}
		m.bias[class] = float32(value)
	}
	return nil
}
