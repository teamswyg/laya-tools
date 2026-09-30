package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoldenFixture(t *testing.T) {
	tasks, hash, err := load("../../benchmarks/routing-golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 64 || len(tasks) != 36 {
		t.Fatal("unexpected fixture")
	}
	var stages [3]int
	for _, task := range tasks {
		stages[task.Stage-1]++
	}
	if stages != [3]int{12, 12, 12} {
		t.Fatal(stages)
	}
}
func TestInvalidLabelsRejected(t *testing.T) {
	for _, data := range []string{`[]`, `[{"id":"x","stage":1,"language":"en","prompt":"task","expected":"easy","rationale":"label"}]`, `[{"id":"x","stage":4,"language":"en","prompt":"task","expected":"fast","rationale":"label"}]`} {
		path := filepath.Join(t.TempDir(), "tasks.json")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := load(path); err == nil {
			t.Fatal("invalid fixture accepted")
		}
	}
}
