// Package ternarytrain trains a small ternary decision head, never the encoder.
package ternarytrain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"io"
	"math"
	"os"
)

type Row struct {
	ID        string     `json:"id"`
	Split     string     `json:"split"`
	Label     int        `json:"label"`
	Feature   []float64  `json:"feature"`
	Reference [3]float64 `json:"reference"`
}
type Dataset struct {
	Schema     string          `json:"schema"`
	Origin     string          `json:"origin"`
	Source     tinyhead.Source `json:"source"`
	Rows       []Row           `json:"rows"`
	Provenance json.RawMessage `json:"provenance"`
}

func Hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func Load(path string) (Dataset, string, error) {
	var d Dataset
	f, e := os.Open(path)
	if e != nil {
		return d, "", e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (64<<20)+1))
	if e != nil {
		return d, "", e
	}
	if len(b) > 64<<20 {
		return d, "", fmt.Errorf("dataset exceeds 64 MiB")
	}
	if e = json.Unmarshal(b, &d); e != nil {
		return d, "", e
	}
	if e = Validate(d); e != nil {
		return d, "", e
	}
	return d, Hash(b), nil
}
func Validate(d Dataset) error {
	if len(d.Rows) == 0 || len(d.Rows) > 10000 {
		return fmt.Errorf("invalid row count")
	}
	isFinal := d.Schema == "riido-qat-final-v1" && d.Origin == "original_synthetic_ternary_qat_final"
	isDev := d.Schema == "riido-tiny-features-v1" && d.Origin == "original_synthetic_pdca06_development"
	if !isFinal && !isDev {
		return fmt.Errorf("unknown dataset provenance")
	}
	seen := map[string]bool{}
	counts := map[string]int{}
	for _, r := range d.Rows {
		if r.ID == "" || seen[r.ID] || r.Label < 0 || r.Label > 2 || len(r.Feature) != 1024 {
			return fmt.Errorf("invalid row identity, label or width")
		}
		seen[r.ID] = true
		counts[r.Split]++
		if isFinal && r.Split != "final" {
			return fmt.Errorf("invalid final split")
		}
		if isDev && r.Split != "train" && r.Split != "validation" && r.Split != "calibration" && r.Split != "test" {
			return fmt.Errorf("final/unknown split in development file")
		}
		for _, v := range r.Feature {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("non-finite feature")
			}
		}
		sum := 0.
		for _, v := range r.Reference {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return fmt.Errorf("invalid reference")
			}
			sum += v
		}
		if math.Abs(sum-1) > 1e-5 {
			return fmt.Errorf("reference not normalized")
		}
	}
	if isDev {
		for _, s := range []string{"train", "validation", "calibration"} {
			if counts[s] == 0 {
				return fmt.Errorf("missing %s", s)
			}
		}
		if len(d.Source.Gamma) != 1024 {
			return fmt.Errorf("invalid source width")
		}
		if _, e := tinyhead.Build(d.Source, tinyhead.Float32, 0); e != nil {
			return e
		}
	}
	return nil
}
func Split(rows []Row, s string) []Row {
	out := []Row{}
	for _, r := range rows {
		if r.Split == s {
			out = append(out, r)
		}
	}
	return out
}
func Winner(p [3]float64) int {
	k := 0
	for i := 1; i < 3; i++ {
		if p[i] > p[k] {
			k = i
		}
	}
	return k
}

type Metrics struct {
	Cases           int     `json:"cases"`
	Correct         int     `json:"correct"`
	Accepted        int     `json:"accepted"`
	AcceptedCorrect int     `json:"accepted_correct"`
	Strong          int     `json:"strong_cases"`
	StrongCorrect   int     `json:"strong_correct"`
	StrongToFast    int     `json:"strong_to_fast"`
	Accuracy        float64 `json:"accuracy"`
	Precision       float64 `json:"accepted_precision"`
	Coverage        float64 `json:"coverage"`
	Recall          float64 `json:"strong_recall"`
	NLL             float64 `json:"nll"`
	Brier           float64 `json:"brier"`
}
type Prediction struct {
	ID          string     `json:"id"`
	Label       int        `json:"label"`
	Winner      int        `json:"winner"`
	Probability [3]float64 `json:"probabilities"`
}

func Evaluate(m *tinyhead.Model, rows []Row) (Metrics, []Prediction, error) {
	var v Metrics
	pred := make([]Prediction, 0, len(rows))
	scratch := make([]float64, m.Width())
	for _, r := range rows {
		if r.Label < 0 || r.Label > 2 {
			return v, nil, fmt.Errorf("invalid label")
		}
		p, e := m.Predict(r.Feature, scratch)
		if e != nil {
			return v, nil, e
		}
		k := Winner(p)
		v.Cases++
		if k == r.Label {
			v.Correct++
		}
		if p[k] >= .9 {
			v.Accepted++
			if k == r.Label {
				v.AcceptedCorrect++
			}
		}
		if r.Label == 2 {
			v.Strong++
			if k == 2 {
				v.StrongCorrect++
			}
			if k == 0 {
				v.StrongToFast++
			}
		}
		v.NLL -= math.Log(math.Max(p[r.Label], 1e-300))
		for c := 0; c < 3; c++ {
			y := 0.
			if c == r.Label {
				y = 1
			}
			v.Brier += (p[c] - y) * (p[c] - y)
		}
		pred = append(pred, Prediction{r.ID, r.Label, k, p})
	}
	if v.Cases == 0 {
		return v, pred, fmt.Errorf("empty evaluation")
	}
	v.Accuracy = float64(v.Correct) / float64(v.Cases)
	v.Coverage = float64(v.Accepted) / float64(v.Cases)
	v.NLL /= float64(v.Cases)
	v.Brier /= float64(v.Cases)
	if v.Accepted > 0 {
		v.Precision = float64(v.AcceptedCorrect) / float64(v.Accepted)
	}
	if v.Strong > 0 {
		v.Recall = float64(v.StrongCorrect) / float64(v.Strong)
	}
	return v, pred, nil
}
func Pass(v, parent Metrics, bytes int) bool {
	return bytes <= 600 && v.Accuracy >= .8 && v.Accuracy+.03 >= parent.Accuracy && v.Recall >= .875 && v.StrongToFast == 0 && v.Precision >= .9 && v.Coverage >= .25
}
