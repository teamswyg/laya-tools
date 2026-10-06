package statehintpilotcli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const validCase = `{"id":"synthetic-case","family":"synthetic-family","locale":"en","text":"Where is the retry option?","expected_intent":"question"}`

func TestCaseDecoderRejectsAmbiguousOrCoercedInput(t *testing.T) {
	row, err := decodeCase([]byte(validCase))
	if err != nil || row.Expected != statehint.Question || row.Text != "Where is the retry option?" {
		t.Fatalf("valid case: %+v %v", row, err)
	}
	for _, input := range []string{
		strings.Replace(validCase, `"text":"Where is the retry option?"`, `"text":null`, 1),
		strings.Replace(validCase, `"locale":"en"`, `"locale":["en"]`, 1),
		strings.Replace(validCase, `"id":"synthetic-case"`, `"id":"synthetic-case","id":"duplicate"`, 1),
		strings.Replace(validCase, `"expected_intent":"question"`, `"Expected_intent":"question"`, 1),
		strings.TrimSuffix(validCase, "}") + `,"credential":"fixture-not-a-token"}`,
		validCase + validCase,
		strings.Replace(validCase, `"family":"synthetic-family",`, "", 1),
	} {
		if _, err := decodeCase([]byte(input)); !errors.Is(err, errCases) {
			t.Errorf("expected strict rejection, got %v", err)
		}
	}
	invalidUTF8 := []byte(validCase)
	invalidUTF8[10] = 0xff
	if _, err := decodeCase(invalidUTF8); !errors.Is(err, errCases) {
		t.Fatal("invalid UTF-8 accepted")
	}
}

func TestCaseFileBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cases.jsonl")
	if err := os.WriteFile(path, []byte(strings.Repeat(validCase+"\n", maxCases+1)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := readCases(path); !errors.Is(err, errCases) {
		t.Fatalf("case budget not enforced: %v", err)
	}
}

func saveToyModel(t *testing.T, trained bool) string {
	t.Helper()
	model := statehint.NewModel()
	if trained {
		_, err := model.Fit([]statehint.Sample{{Text: "Where is the retry option?", Label: statehint.Question}}, statehint.FitOptions{Epochs: 1, BatchSize: 1, Seed: 17})
		if err != nil {
			t.Fatal(err)
		}
	}
	var b bytes.Buffer
	if err := model.Save(&b); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "toy.rsh")
	if err := os.WriteFile(path, b.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNoFallbackFromUntrainedOrCorruptModel(t *testing.T) {
	if _, _, err := loadPredictor(saveToyModel(t, false)); !errors.Is(err, errArtifact) {
		t.Fatalf("untrained accepted: %v", err)
	}
	path := saveToyModel(t, true)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b[len(b)-1] ^= 1
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadPredictor(path); !errors.Is(err, errArtifact) {
		t.Fatalf("corrupt accepted: %v", err)
	}
}

func TestRunUsesActualWeightsAndDoesNotEchoText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cases.jsonl")
	if err := os.WriteFile(path, []byte(validCase+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	if err := Run([]string{"--cases", path, "--model", saveToyModel(t, true), "--compare-rules"}, &out, &stderr); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out.Bytes(), []byte("Where is the retry option?")) {
		t.Fatal("source text leaked into report")
	}
	var result Comparison
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Variants) != 3 || result.ReaderKind != "synthetic_fixture_only" || result.MutationExecuted || result.TrainingPerformed {
		t.Fatalf("wrong execution scope: %+v", result)
	}
	if result.Variants[0].Artifact == nil || result.Variants[0].Artifact.TrainingSteps != 1 || result.Variants[0].Report.Budgets.PredictionCallsUsed != 1 {
		t.Fatal("trained artifact not actually evaluated")
	}
}

func TestExistingOutputIsPreserved(t *testing.T) {
	root := t.TempDir()
	cases := filepath.Join(root, "cases.jsonl")
	output := filepath.Join(root, "report.json")
	if err := os.WriteFile(cases, []byte(validCase+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("retained-first-result"), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"--cases", cases, "--model", saveToyModel(t, true), "--out", output}, &stdout, &stderr); err == nil {
		t.Fatal("existing report overwritten")
	}
	b, err := os.ReadFile(output)
	if err != nil || string(b) != "retained-first-result" {
		t.Fatal("original report changed")
	}
}

func TestChecksumMismatchCreatesNoReport(t *testing.T) {
	root := t.TempDir()
	cases := filepath.Join(root, "cases.jsonl")
	output := filepath.Join(root, "report.json")
	if err := os.WriteFile(cases, []byte(validCase+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	err := Run([]string{"--cases", cases, "--model", saveToyModel(t, true), "--model-sha256", strings.Repeat("0", 64), "--out", output}, &stdout, &stderr)
	if err == nil {
		t.Fatal("wrong asset pin accepted")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("mismatch created output evidence")
	}
}

func TestHelpNeedsNoModelAndSucceeds(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Run([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "-cases") {
		t.Fatal("missing usage")
	}
}
