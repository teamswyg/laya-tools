package inference

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTokenizerParity(t *testing.T) {
	dir := os.Getenv("LAYA_MODEL_DIR")
	if dir == "" {
		t.Skip("set LAYA_MODEL_DIR for HF reference parity")
	}
	enc, err := LoadEncoder(filepath.Join(dir, "tokenizer.json"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile("../../testdata/tokenizer-parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		State, Kind, Instruction string
		Options                  []string
		IDs, Markers             []int64
	}
	if err = json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		seq, err := enc.Build(c.State, c.Kind, c.Instruction, c.Options, 512)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(seq.IDs, c.IDs) || !reflect.DeepEqual(seq.Markers, c.Markers) {
			t.Errorf("parity failed for %q\ngot %v\nwant %v", c.State, seq.IDs, c.IDs)
		}
	}
}
func TestNativeInference(t *testing.T) {
	dir := os.Getenv("LAYA_MODEL_DIR")
	lib := os.Getenv("LAYA_RUNTIME")
	if dir == "" || lib == "" {
		t.Skip("set model and runtime for native integration")
	}
	e, err := New(Options{ModelDir: dir, Runtime: lib, Provider: "cpu", Threads: 4, MaxTokens: 512})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	for _, state := range []string{"The deployment succeeded. All checks passed.", "The deployment failed. Every check failed."} {
		p, err := e.Predict(state, "noul", "Did the deployment succeed?", []string{"false: no, the statement does not hold", "true: yes, the statement holds"})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s %+v", state, p)
		if len(p.Probabilities) != 2 {
			t.Fatal("wrong shape")
		}
		if state[15] == 's' && p.Winner != 1 {
			t.Fatal("positive reference failed")
		}
		if state[15] == 'f' && p.Winner != 0 {
			t.Fatal("negative reference failed")
		}
	}
}
