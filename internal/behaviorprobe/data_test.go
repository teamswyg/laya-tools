package behaviorprobe

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func fixture(t *testing.T) Dataset {
	t.Helper()
	d, err := Load("../../experiments/short-claim/probes-56.json")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestAuthoredDeclarationsAndLiteralControls(t *testing.T) {
	b, err := os.ReadFile("data.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != authoredSource {
		t.Fatal("embedded original source changed")
	}
	pins, err := SourcePins()
	if err != nil {
		t.Fatal(err)
	}
	if len(pins) != 36 || len(contracts) != 12 {
		t.Fatal("control catalog count")
	}
	var vectors, correct, wrong int
	for _, c := range contracts {
		vectors += len(c.vectors)
	}
	if vectors != 68 {
		t.Fatalf("independent literal vectors %d", vectors)
	}
	for i, s := range sourceCatalog {
		name := runtime.FuncForPC(reflect.ValueOf(s.run).Pointer()).Name()
		if !strings.HasSuffix(name, "."+s.function) {
			t.Fatal("code pin / executed function mismatch")
		}
		checked, failed, unknown := check(s.run, contracts[contractIndex(s.prototype)].vectors)
		if checked == 0 || unknown || s.correct && failed != 0 || !s.correct && failed == 0 {
			t.Fatalf("control %s: %d/%d unknown=%v", s.id, checked, failed, unknown)
		}
		if s.correct {
			correct++
		} else {
			wrong++
		}
		if pins[i].CodeSHA256 == "" || pins[i].NormalizedCodeSHA256 == "" {
			t.Fatal("missing source hash")
		}
	}
	if correct != 12 || wrong != 24 {
		t.Fatal("correct/wrong controls")
	}
}

func TestDatasetCountsUnknownAndNoAnswer(t *testing.T) {
	d := fixture(t)
	r, err := Evaluate(d)
	if err != nil {
		t.Fatal(err)
	}
	if r.Parents != 48 || r.Candidates != 144 || r.PrototypeCount != 12 || r.KnownParents != 36 || r.UnknownParents != 12 || r.NoAnswerParents != 12 {
		t.Fatalf("prepared counts %+v", r)
	}
	if r.ConnectedGroups != 11 || r.MeetsGroupMinimum || r.TrainingReady || r.Partitioned || r.FinalEligible || !r.NaturalLanguageReviewRequired {
		t.Fatal("insufficient groups must not create eligibility")
	}
	var answers, multiple int
	for _, o := range r.Outcomes {
		switch o.State {
		case "known":
			answers++
			if len(o.Acceptable) == 0 {
				t.Fatal("known answer missing")
			}
			if len(o.Acceptable) > 1 {
				multiple++
			}
		case "no_answer":
			if len(o.Acceptable) != 0 {
				t.Fatal("no-answer acceptable set")
			}
			for _, c := range o.Candidates {
				if c.State != "rejected" || c.VectorsChecked == 0 {
					t.Fatal("no-answer lacks independent rejection")
				}
			}
		case "unknown":
			if len(o.Acceptable) != 0 || o.Reason != "request_policy_ambiguous_no_literal_oracle" {
				t.Fatal("ambiguous request gained a label")
			}
			for _, c := range o.Candidates {
				if c.State != "unknown" || c.VectorsChecked != 0 {
					t.Fatal("unknown candidate label")
				}
			}
		default:
			t.Fatal("unknown status")
		}
	}
	if answers != 24 || multiple != 12 {
		t.Fatal("set-valued answers not retained")
	}
	for _, g := range r.Groups {
		if slices.Contains(g.Prototypes, "closed-window") && !slices.Contains(g.Prototypes, "clamp-window") {
			t.Fatal("shared closed-boundary core split")
		}
	}
}

func TestOnlyTextProjectionPreservesOrderAndOwnsSlices(t *testing.T) {
	d := fixture(t)
	a := FeatureInputs(d)
	if len(a) != len(d.Parents) {
		t.Fatal("projection count")
	}
	for i, p := range d.Parents {
		if a[i].Request != p.Request || len(a[i].Candidates) != len(p.Candidates) {
			t.Fatal("projection order")
		}
		for j, c := range p.Candidates {
			if a[i].Candidates[j] != c.Text {
				t.Fatal("candidate order")
			}
		}
	}
	b, _ := json.Marshal(a)
	for _, field := range []string{"contract_id", "source_id", "code_sha256", "prototype", "acceptable", "train", "correct_rank"} {
		if strings.Contains(string(b), field) {
			t.Fatal("bookkeeping entered features")
		}
	}
	a[0].Candidates[0] = "changed"
	if d.Parents[0].Candidates[0].Text == "changed" {
		t.Fatal("shared candidate projection memory")
	}
}

func TestBoundsPinsAndInjectedLabelFields(t *testing.T) {
	d := fixture(t)
	seed, _ := json.Marshal(d)
	tests := []struct {
		name   string
		change func(*Dataset)
	}{
		{"schema", func(x *Dataset) { x.Schema = "unknown" }},
		{"origin", func(x *Dataset) { x.Origin = "private" }},
		{"empty parents", func(x *Dataset) { x.Parents = nil }},
		{"too many parents", func(x *Dataset) {
			for len(x.Parents) <= MaxParents {
				x.Parents = append(x.Parents, x.Parents[0])
			}
		}},
		{"empty candidates", func(x *Dataset) { x.Parents[0].Candidates = nil }},
		{"too many candidates", func(x *Dataset) {
			for len(x.Parents[0].Candidates) < 9 {
				x.Parents[0].Candidates = append(x.Parents[0].Candidates, x.Parents[0].Candidates[0])
			}
		}},
		{"bytes", func(x *Dataset) { x.Parents[0].Request = strings.Repeat("x", 513) }},
		{"words", func(x *Dataset) { x.Parents[0].Candidates[0].Text = strings.Repeat("word ", 33) }},
		{"invalid utf8", func(x *Dataset) { x.Parents[0].Request = string([]byte{255}) }},
		{"source pin", func(x *Dataset) { x.Parents[0].Candidates[0].CodeSHA256 = strings.Repeat("0", 64) }},
		{"source identity", func(x *Dataset) { x.Parents[0].Candidates[0].SourceID = "unregistered" }},
		{"contract identity", func(x *Dataset) { x.Parents[0].ContractID = "unknown-new-rule" }},
		{"duplicate parent", func(x *Dataset) { x.Parents[1].ID = x.Parents[0].ID }},
		{"duplicate candidate", func(x *Dataset) { x.Parents[0].Candidates[1].ID = x.Parents[0].Candidates[0].ID }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var x Dataset
			_ = json.Unmarshal(seed, &x)
			tt.change(&x)
			if _, err := Evaluate(x); err == nil {
				t.Fatal("invalid data accepted")
			}
		})
	}
	for _, body := range []string{
		strings.Replace(string(seed), `"parents":`, `"labels":[1],"parents":`, 1),
		strings.Replace(string(seed), `"request":`, `"target_rank":1,"request":`, 1),
		string(seed) + "{}",
		string([]byte{255}) + string(seed),
	} {
		p := filepath.Join(t.TempDir(), "data.json")
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(p); err == nil {
			t.Fatal("unknown labels/trailing content/raw invalid UTF-8 accepted")
		}
	}
}

func TestTransitiveNegativeAndNormalizedSourceGrouping(t *testing.T) {
	d := fixture(t)
	pins, err := SourcePins()
	if err != nil {
		t.Fatal(err)
	}
	_, before := groupAudit(d, pins)
	if len(before) != 11 {
		t.Fatal("initial groups")
	}
	// The shared wrong source is a bridge between previously separate families.
	// Its status does not permit cross-role source reuse; joining is transitive.
	d.Parents[8].Candidates = append(d.Parents[8].Candidates, d.Parents[0].Candidates[0])
	_, after := groupAudit(d, pins)
	if len(after) != 10 {
		t.Fatalf("negative source bridge groups %d", len(after))
	}
	a := []byte("func alpha(x []int) []int { return []int{x[0]+1} }")
	b := []byte("func beta(y []int) []int { return []int{y[0]+1} }")
	c := []byte("func beta(y []int) []int { return []int{y[0]+2} }")
	if normalizedCodeSHA(a) != normalizedCodeSHA(b) || normalizedCodeSHA(a) == normalizedCodeSHA(c) {
		t.Fatal("normalized behavior signature")
	}
}

func TestCheckFailureRemainsUnknownAndInputsAreOwned(t *testing.T) {
	vectors := []vector{{[]int{1, 2}, []int{1}}}
	_, _, unknown := check(func(v []int) []int { panic("check unavailable") }, vectors)
	if !unknown {
		t.Fatal("panic became a false behavior label")
	}
	_, _, unknown = check(func(v []int) []int { v[0] = 9; return []int{1} }, vectors)
	if !unknown || vectors[0].Input[0] != 1 {
		t.Fatal("candidate mutation escaped or became a label")
	}
}
