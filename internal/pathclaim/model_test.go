package pathclaim

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

func sha256Hex(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// This is a hand-authored model-definition fixture, not a trained artifact.
func ownedModel(t *testing.T) Model {
	t.Helper()
	_, ready := ownedReadyRows(t)
	m := Model{Schema: ModelSchema, FeatureSchema: searchclaim.PathSchema, CoefficientPrecision: "float32", Readiness: ready, Lambda: 1, Seed: 1729, Epoch: 4, Threshold: Threshold{MinimumScore: .25, BudgetPercent: 90, ValidationQuestions: 2400, ValidationCalls: 1200, ValidationPages: 6000}}
	m.Coefficients[0], m.Coefficients[1] = .5, -.25
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestModelStorageSchemaAndStrictCoefficientCount(t *testing.T) {
	m := ownedModel(t)
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Model
	if err = json.Unmarshal(b, &decoded); err != nil || !reflect.DeepEqual(m, decoded) {
		t.Fatal("definition roundtrip", err)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(b, &obj); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{0, 15, 17} {
		wrong := make([]float32, n)
		obj["Coefficients"], _ = json.Marshal(wrong)
		bad, _ := json.Marshal(obj)
		before := decoded
		if err := json.Unmarshal(bad, &decoded); err == nil || !reflect.DeepEqual(decoded, before) {
			t.Fatal("coefficient shape accepted or receiver changed", n)
		}
	}
	obj["Coefficients"], _ = json.Marshal(m.Coefficients)
	obj["extra"] = json.RawMessage(`1`)
	bad, _ := json.Marshal(obj)
	if json.Unmarshal(bad, &decoded) == nil {
		t.Fatal("unknown model field accepted")
	}
	if json.Unmarshal(append(b, []byte(" {}")...), &decoded) == nil {
		t.Fatal("extra JSON accepted")
	}
	if _, err := DecodeModel([]byte(strings.Repeat(" ", 16<<10+1) + string(b))); err == nil {
		t.Fatal("model byte bound ignored")
	}
	delete(obj, "extra")
	var seal map[string]json.RawMessage
	if err := json.Unmarshal(obj["Readiness"], &seal); err != nil {
		t.Fatal(err)
	}
	var repos []RepositoryCount
	if err := json.Unmarshal(seal["ValidationRepositories"], &repos); err != nil {
		t.Fatal(err)
	}
	repos = append(repos, repos[0])
	seal["ValidationRepositories"], _ = json.Marshal(repos)
	obj["Readiness"], _ = json.Marshal(seal)
	bad, _ = json.Marshal(obj)
	if _, err := DecodeModel(bad); err == nil {
		t.Fatal("extra readiness repository silently discarded")
	}
}

func TestModelMetadataAndNonfiniteCoefficientRejections(t *testing.T) {
	for _, mode := range []string{"schema", "features", "precision", "plan", "source", "penalty", "seed", "epoch", "nan", "infinity", "cap", "calls", "disabled-calls", "questions", "pages", "threshold-nan"} {
		t.Run(mode, func(t *testing.T) {
			m := ownedModel(t)
			switch mode {
			case "schema":
				m.Schema = "old"
			case "features":
				m.FeatureSchema = "riido-search-claim-v1"
			case "precision":
				m.CoefficientPrecision = "ternary"
			case "plan":
				m.Readiness.PlanSHA256 = strings.Repeat("a", 64)
			case "source":
				m.Readiness.SourceEvidenceSHA256 = ""
			case "penalty":
				m.Lambda = .5
			case "seed":
				m.Seed = 123
			case "epoch":
				m.Epoch = 101
			case "nan":
				m.Coefficients[3] = float32(math.NaN())
			case "infinity":
				m.Coefficients[3] = float32(math.Inf(1))
			case "cap":
				m.Threshold.BudgetPercent = 100
			case "calls":
				m.Threshold.ValidationCalls = 2161
			case "disabled-calls":
				m.Threshold.Disabled = true
			case "questions":
				m.Threshold.ValidationQuestions = 2399
			case "pages":
				m.Threshold.ValidationPages = 2400*MaxPages + 1
			case "threshold-nan":
				m.Threshold.MinimumScore = math.NaN()
			}
			if err := m.Validate(); err == nil {
				t.Fatal("invalid model accepted")
			}
			if _, err := m.Selector(); err == nil {
				t.Fatal("invalid selector created")
			}
		})
	}
}

func TestSelectorCopiesCoefficientsAndRejectsBadFeatures(t *testing.T) {
	m := ownedModel(t)
	selector, err := m.Selector()
	if err != nil {
		t.Fatal(err)
	}
	x := [16]float64{1, .5}
	if score, err := m.Score(x); err != nil || score != .375 {
		t.Fatal("wrong float64 arithmetic", score, err)
	}
	use, err := selector(x)
	if err != nil || !use {
		t.Fatal("valid selection", err)
	}
	m.Coefficients[0], m.Threshold.MinimumScore = -100, 100
	if use, err := selector(x); err != nil || !use {
		t.Fatal("selector aliases mutable caller model", err)
	}
	for _, bad := range []float64{-.1, 1.1, math.NaN(), math.Inf(1)} {
		x[5] = bad
		if use, err := selector(x); err == nil || use {
			t.Fatal("invalid features accepted")
		}
	}
	disabled := ownedModel(t)
	disabled.Threshold.Disabled, disabled.Threshold.ValidationCalls = true, 0
	skip, err := disabled.Selector()
	if err != nil {
		t.Fatal(err)
	}
	if use, err := skip([16]float64{1}); err != nil || use {
		t.Fatal("disabled control requested helper", err)
	}
	if _, err := NewModel(Trial{}, disabled.Readiness, disabled.Threshold); err == nil {
		t.Fatal("unfitted trial became model")
	}
	forged := Trial{Fitted: true, Attempted: true, Lambda: 1, Seed: 1729, Epoch: 4}
	if _, err := NewModel(forged, disabled.Readiness, disabled.Threshold); err == nil {
		t.Fatal("manual fit flag bypassed input provenance")
	}
}
