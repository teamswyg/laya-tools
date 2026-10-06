package statehint

import "math"

type Source string

const (
	Learned    Source = "linear_softmax_trained"
	Untrained  Source = "linear_softmax_untrained"
	RuleSource Source = "rule_baseline_unlearned"
)

type Prediction struct {
	Intent        Intent               `json:"intent"`
	Probabilities [IntentCount]float64 `json:"probabilities"`
	Confidence    float64              `json:"confidence"`
	Margin        float64              `json:"margin"`
	Source        Source               `json:"source"`
	TrainingSteps uint64               `json:"training_steps"`
	GuardReason   string               `json:"guard_reason,omitempty"`
}

// Model uses fixed arrays of numeric feature rows and eight class columns,
// with no locks or process-global cache.
// Training and SetTemperature require exclusive ownership of the model.
type Model struct {
	weights     [FeatureBins][IntentCount]float32
	bias        [IntentCount]float32
	temperature float64
	steps       uint64
}

func NewModel() *Model { return &Model{temperature: 1} }

func (m *Model) Clone() *Model {
	if m == nil {
		return nil
	}
	copy := *m
	return &copy
}

func (m *Model) Temperature() float64 {
	if m == nil {
		return 0
	}
	return m.temperature
}
func (m *Model) TrainingSteps() uint64 {
	if m == nil {
		return 0
	}
	return m.steps
}

func (m *Model) SetTemperature(value float64) error {
	if m == nil || !finite(value) || value < 0.05 || value > 20 {
		return ErrModel
	}
	m.temperature = value
	return nil
}

func (m *Model) valid() bool {
	if m == nil || !finite(m.temperature) || m.temperature < 0.05 || m.temperature > 20 {
		return false
	}
	for _, row := range m.weights {
		for _, value := range row {
			if !finite(float64(value)) {
				return false
			}
		}
	}
	for _, value := range m.bias {
		if !finite(float64(value)) {
			return false
		}
	}
	return true
}

func (m *Model) logits(w *Workspace) [IntentCount]float64 {
	var result [IntentCount]float64
	for class, bias := range m.bias {
		result[class] = float64(bias)
	}
	for _, index := range w.indices[:w.count] {
		value := float64(w.values[index])
		for class, weight := range m.weights[index] {
			result[class] += float64(weight) * value
		}
	}
	return result
}

func softmax(logits [IntentCount]float64, temperature float64) ([IntentCount]float64, bool) {
	var result [IntentCount]float64
	maxValue := logits[0]
	for _, value := range logits {
		if !finite(value) {
			return result, false
		}
		if value > maxValue {
			maxValue = value
		}
	}
	var total float64
	for class, value := range logits {
		result[class] = math.Exp((value - maxValue) / temperature)
		total += result[class]
	}
	if !finite(total) || total <= 0 {
		return result, false
	}
	for class := range result {
		result[class] /= total
	}
	return result, true
}

func prediction(probabilities [IntentCount]float64, source Source, steps uint64) Prediction {
	winner, runner := IntentCount-1, -1 // deterministic uncertainty tie -> unclear
	for class, value := range probabilities {
		if value > probabilities[winner] {
			winner = class
		}
	}
	for class := range probabilities {
		if class != winner && (runner < 0 || probabilities[class] > probabilities[runner]) {
			runner = class
		}
	}
	return Prediction{Intent: Intents()[winner], Probabilities: probabilities, Confidence: probabilities[winner], Margin: probabilities[winner] - probabilities[runner], Source: source, TrainingSteps: steps}
}

func (m *Model) Predict(text string, workspace *Workspace) (Prediction, error) {
	if m == nil || !finite(m.temperature) || m.temperature < 0.05 || m.temperature > 20 {
		return Prediction{}, ErrModel
	}
	if workspace == nil {
		workspace = &Workspace{}
	}
	if err := extract(text, workspace); err != nil {
		return Prediction{}, err
	}
	source := Learned
	if m.steps == 0 {
		source = Untrained
	}
	if workspace.wordCount == 0 {
		var uniform [IntentCount]float64
		for index := range uniform {
			uniform[index] = 1.0 / IntentCount
		}
		p := prediction(uniform, source, m.steps)
		p.GuardReason = "no_word_content"
		return p, nil
	}
	probabilities, ok := softmax(m.logits(workspace), m.temperature)
	if !ok {
		return Prediction{}, ErrModel
	}
	return prediction(probabilities, source, m.steps), nil
}
