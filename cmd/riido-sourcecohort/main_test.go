package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func TestHelpDefaultsAndSanitizedArguments(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--default-recipe"}} {
		var out bytes.Buffer
		if run(args, &out) != 0 || out.Len() == 0 {
			t.Fatal("help/defaults unusable")
		}
	}
	var out bytes.Buffer
	if run([]string{"--private-secret-option"}, &out) != 2 || bytes.Contains(out.Bytes(), []byte("private-secret")) {
		t.Fatal("invalid argument echoed")
	}
}
func TestAllocationCLIWithSyntheticWhole400(t *testing.T) {
	root := t.TempDir()
	save := func(name string, v any) sourcecohort.File {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(root, name), b, 0600); e != nil {
			t.Fatal(e)
		}
		return sourcecohort.File{Path: name, SHA256: sourcecohort.Hash(b), Bytes: int64(len(b))}
	}
	recipe := sourcecohort.DefaultRecipe()
	reg := sourcecohort.Registry{Schema: "riido-sourcecohort-registry-v1", CohortID: "synthetic-cohort", Rows: []sourcecohort.Row{}}
	id := 0
	for a, stratum := range recipe.Strata {
		for _, split := range recipe.Splits {
			for j := 0; j < split.PerStratum; j++ {
				style := ""
				if split.Name == "dev" {
					style = "short"
					if a >= 6 {
						style = "general"
					}
				}
				reg.Rows = append(reg.Rows, sourcecohort.Row{ID: fmt.Sprintf("synthetic-%03d", id), Stratum: stratum, Split: split.Name, DEVStyle: style, Dependencies: []string{}})
				id++
			}
		}
	}
	plan := sourcecohort.Plan{Schema: "riido-sourcecohort-plan-v1", Registry: save("registry.json", reg), SplitConfig: save("recipe.json", recipe), InputMaxBytes: 1 << 20, SourceMaxBytes: 16384, FreezeMaxBytes: 1 << 20, OutputMaxBytes: 2 << 20}
	pin := save("plan.json", plan)
	var out bytes.Buffer
	exit := run([]string{"--root", root, "--plan", pin.Path, "--sha256", pin.SHA256, "--bytes", fmt.Sprint(pin.Bytes), "--out", filepath.Join(root, "out"), "--allocation-only"}, &out)
	if exit != 0 {
		t.Fatalf("allocation CLI %d %s", exit, out.String())
	}
	if bytes.Contains(out.Bytes(), []byte("synthetic-")) || bytes.Contains(out.Bytes(), []byte(root)) {
		t.Fatal("stdout exposed private identifiers/paths")
	}
	var report sourcecohort.Summary
	if json.Unmarshal(out.Bytes(), &report) != nil || report.State != "allocation_ready" || report.Expected != 400 || report.MeaningProven {
		t.Fatal("wrong stdout aggregate")
	}
}
