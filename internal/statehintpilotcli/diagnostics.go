package statehintpilotcli

import (
	"math"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintpilot"
)

type ConfidenceBin struct {
	Cases          int     `json:"cases"`
	Correct        int     `json:"correct"`
	MeanConfidence float64 `json:"mean_confidence"`
	Accuracy       float64 `json:"accuracy"`
}

// These are diagnostic losses on authored development labels, never evidence
// of real completion or calibration at product base rates. Display loss groups
// the five excluded meanings as none without altering prediction or eligibility.
type ProbabilityDiagnostics struct {
	Rows               int               `json:"rows"`
	Classified         int               `json:"classified"`
	Unavailable        int               `json:"unavailable"`
	Guarded            int               `json:"guarded"`
	NLL8               float64           `json:"nll_8_intents"`
	Brier8             float64           `json:"brier_8_intents_sum"`
	DisplayNLL4        float64           `json:"nll_3_displays_plus_none"`
	DisplayBrier4      float64           `json:"brier_3_displays_plus_none_sum"`
	LogClamp           float64           `json:"log_clamp"`
	NLL8Clamped        int               `json:"nll_8_clamped"`
	DisplayNLL4Clamped int               `json:"display_nll_4_clamped"`
	ECE10              float64           `json:"expected_calibration_error_10_bins"`
	ConfidenceBins     [10]ConfidenceBin `json:"confidence_bins"`
}

func displayIndex(intent statehint.Intent) int {
	switch intent {
	case statehint.Progress:
		return 0
	case statehint.CompletionReport:
		return 1
	case statehint.Question:
		return 2
	default:
		return 3
	}
}

func probabilityDiagnostics(report statehintpilot.Report) ProbabilityDiagnostics {
	d := ProbabilityDiagnostics{Rows: len(report.Observations), LogClamp: 1e-15}
	for _, o := range report.Observations {
		if o.Observed == "" {
			d.Unavailable++
			continue
		}
		d.Classified++
		if o.Guard != "" {
			d.Guarded++
		}
		expected, _ := statehint.IntentIndex(o.Expected)
		p := o.Probability[expected]
		if p < d.LogClamp {
			p = d.LogClamp
			d.NLL8Clamped++
		}
		d.NLL8 -= math.Log(p)
		var grouped [4]float64
		for index, value := range o.Probability {
			y := 0.
			if index == expected {
				y = 1
			}
			d.Brier8 += (value - y) * (value - y)
			grouped[displayIndex(report.IntentOrder[index])] += value
		}
		groupExpected := displayIndex(o.Expected)
		p = grouped[groupExpected]
		if p < d.LogClamp {
			p = d.LogClamp
			d.DisplayNLL4Clamped++
		}
		d.DisplayNLL4 -= math.Log(p)
		for group, value := range grouped {
			y := 0.
			if group == groupExpected {
				y = 1
			}
			d.DisplayBrier4 += (value - y) * (value - y)
		}
		bin := min(int(o.Confidence*10), 9)
		b := &d.ConfidenceBins[bin]
		b.Cases++
		b.MeanConfidence += o.Confidence
		if o.Observed == o.Expected {
			b.Correct++
		}
	}
	if d.Classified > 0 {
		n := float64(d.Classified)
		d.NLL8 /= n
		d.Brier8 /= n
		d.DisplayNLL4 /= n
		d.DisplayBrier4 /= n
		for i := range d.ConfidenceBins {
			b := &d.ConfidenceBins[i]
			if b.Cases == 0 {
				continue
			}
			b.MeanConfidence /= float64(b.Cases)
			b.Accuracy = float64(b.Correct) / float64(b.Cases)
			d.ECE10 += float64(b.Cases) / n * math.Abs(b.MeanConfidence-b.Accuracy)
		}
	}
	return d
}
