package statehint

import (
	"errors"
	"math"
	"sync"
	"testing"
)

// Tiny original sentences exercise the training implementation only. These
// tests are not native-model evaluation, deployment evidence or accuracy claims.
func toySamples() []Sample {
	return []Sample{
		{Text: "how can I select a folder", Label: Question},
		{Text: "broken upload error is blocking delivery", Label: Blocker},
		{Text: "reference manual attached for the reader", Label: Reference},
		{Text: "working on the parser implementation", Label: Progress},
		{Text: "finished the reviewed checklist successfully", Label: CompletionReport},
		{Text: "cancel this obsolete task please", Label: CancelRequest},
		{Text: "planned for next week after the meeting", Label: Planned},
		{Text: "miscellaneous vague words without context", Label: Unclear},
	}
}

func TestSupervisedFitCloneWarmStartAndTemperature(t *testing.T) {
	model := NewModel()
	samples := toySamples()
	before, err := model.Predict(samples[0].Text, nil)
	if err != nil || before.Source != Untrained || before.Intent != Unclear {
		t.Fatal(before, err)
	}
	options := FitOptions{Epochs: 80, BatchSize: 8, LearningRate: .03, WeightDecay: .01, Seed: 37}
	report, err := model.Fit(samples, options)
	if err != nil || report.TrainingSteps != 80 || report.Batches != 80 || report.MeanLoss >= 1 {
		t.Fatal(report, err)
	}
	for _, sample := range samples {
		p, err := model.Predict(sample.Text, nil)
		if err != nil || p.Intent != sample.Label || p.Confidence < .8 || p.Source != Learned {
			t.Fatalf("toy fitting did not converge: %s %v %v", sample.Label, p, err)
		}
	}
	copy := model.Clone()
	baseSteps := model.TrainingSteps()
	if _, err := copy.Fit(samples, FitOptions{Epochs: 1, BatchSize: 4, Seed: 21}); err != nil {
		t.Fatal(err)
	}
	if model.TrainingSteps() != baseSteps || copy.TrainingSteps() != baseSteps+2 || copy.weights == model.weights {
		t.Fatal("warm-start clone aliased or steps did not continue")
	}
	old, _ := model.Predict(samples[0].Text, nil)
	if err := model.SetTemperature(2); err != nil {
		t.Fatal(err)
	}
	calibrated, _ := model.Predict(samples[0].Text, nil)
	if calibrated.Intent != old.Intent || calibrated.Confidence >= old.Confidence {
		t.Fatal("temperature did not affect posterior")
	}
	for _, bad := range []float64{0, .049, 20.01, math.NaN(), math.Inf(1)} {
		if !errors.Is(model.SetTemperature(bad), ErrModel) {
			t.Fatal("invalid temperature accepted")
		}
	}
	if model.Temperature() != 2 {
		t.Fatal("invalid calibration changed the model")
	}
}

func TestFitDeterministicAndInvalidInputDoesNotUpdate(t *testing.T) {
	first, second := NewModel(), NewModel()
	options := FitOptions{Epochs: 3, BatchSize: 3, Seed: 123}
	if _, err := first.Fit(toySamples(), options); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Fit(toySamples(), options); err != nil {
		t.Fatal(err)
	}
	if *first != *second {
		t.Fatal("fixed seed fitting is nondeterministic")
	}
	snapshot := *first
	samples := append(toySamples(), Sample{Text: "bad label", Label: "other"})
	if _, err := first.Fit(samples, options); !errors.Is(err, ErrTraining) || *first != snapshot {
		t.Fatal("invalid sample partially updated model")
	}
	for _, bad := range []FitOptions{{Epochs: -1}, {BatchSize: -1}, {LearningRate: math.NaN()}, {WeightDecay: -1}} {
		if _, err := first.Fit(toySamples(), bad); !errors.Is(err, ErrTraining) {
			t.Fatal("invalid training settings accepted")
		}
	}
}

func TestNoWordContentGuardAndFiniteLogits(t *testing.T) {
	model := NewModel()
	model.steps = 1 // deliberately synthetic test scores, not a learned fixture
	model.bias[4] = 100
	for _, text := range []string{"", " \t\n", "✅", "📎❓"} {
		p, err := model.Predict(text, nil)
		if err != nil || p.Intent != Unclear || p.Confidence != .125 || p.GuardReason != "no_word_content" {
			t.Fatal(p, err)
		}
	}
	model.bias[0] = float32(math.NaN())
	if _, err := model.Predict("test words", nil); !errors.Is(err, ErrModel) {
		t.Fatal("nonfinite score accepted")
	}
}

func TestRuleBaselineIsExplicitlyUnlearnedAndAbstains(t *testing.T) {
	for _, sample := range []Sample{{"how do I choose", Question}, {"blocked by an error", Blocker}, {"reference attached", Reference}, {"working on implementation", Progress}, {"finished the checklist", CompletionReport}, {"cancel this task", CancelRequest}, {"planned tomorrow", Planned}, {"not completed yet", Unclear}, {"working tomorrow", Unclear}, {"background information", Unclear}} {
		p, err := RuleBaseline(sample.Text)
		if err != nil || p.Intent != sample.Label || p.Source != RuleSource || p.TrainingSteps != 0 {
			t.Fatal(sample, p, err)
		}
	}
}

func TestReadOnlyPredictionUsesIndependentCallerWorkspaces(t *testing.T) {
	model := NewModel()
	if _, err := model.Fit(toySamples(), FitOptions{Epochs: 2, BatchSize: 8, Seed: 7}); err != nil {
		t.Fatal(err)
	}
	want, err := model.Predict("original text for concurrent reads", nil)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			var workspace Workspace
			for range 20 {
				got, err := model.Predict("original text for concurrent reads", &workspace)
				if err != nil || got != want {
					t.Errorf("concurrent read changed prediction: %v", err)
					return
				}
			}
		})
	}
	group.Wait()
}
