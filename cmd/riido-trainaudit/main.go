// riido-trainaudit checks separate training identities without scoring tasks.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const planSHA = "ac31cd3b2dcdb90cae9435faab2c276d4bfc1c09830b2c30c3e35395c2411a0e"
const projectionSHA = "82419343bf7a06daffe79b2145800648e776f971a0f6f873fd11277602a87f51"

func writeJSON(path string, x any) error {
	b, e := json.MarshalIndent(x, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
func run() error {
	out := flag.String("out", "", "new private audit directory")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require output")
	}
	plan, e := os.ReadFile("experiments/training-source/plan-36.json")
	if e != nil {
		return e
	}
	h := sha256.Sum256(plan)
	if hex.EncodeToString(h[:]) != planSHA {
		return fmt.Errorf("plan mismatch")
	}
	train, e := sweaudit.ReadTraining(".cache/training-identity-36.jsonl", projectionSHA)
	if e != nil {
		return e
	}
	full, e := sweaudit.Read(".cache/real-task-full-28.jsonl", sweaudit.FullProjectionSHA256)
	if e != nil {
		return e
	}
	multi, e := sweaudit.Read(".cache/real-task-multilingual-28.jsonl", sweaudit.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	if len(train) != 19008 || len(full) != 2294 || len(multi) != 300 {
		return fmt.Errorf("source row count mismatch")
	}
	old, _, e := sweaudit.GroupEvaluation(full, multi)
	if e != nil || old.SelectedSHA256 != sweaudit.FrozenSelectionSHA256 {
		return fmt.Errorf("original evaluation selection changed")
	}
	report, selected, e := sweaudit.TrainingCandidates(train, append(slices.Clone(full), multi...), []string{"etcd-io/etcd", "kubernetes/kubernetes", "lxc/lxd"})
	if e != nil {
		return e
	}
	result := struct {
		PlanSHA256, SourceRevision, SourceSHA256, ProjectionSHA256, CardSHA256, PreviousEvaluationSHA256 string
		Audit                                                                                            sweaudit.TrainingReport
		DatasetCardLicenseDeclared, TrainingExecuted, ModelReleased                                      bool
	}{planSHA, "e48e2bd1e9fecd5bbd641e9414ac59da9f2e69f6", "0ba403a7060af657e1f0476937a703726c07da15b759836694b16e553705a10f", projectionSHA, "f0fb86ca5917025357564e0735f484bcfd340cc10445ca4eac4d866e9a164d52", old.SelectedSHA256, report, false, false, false}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if e = writeJSON(filepath.Join(*out, "results.json"), result); e != nil {
		return e
	}
	if e = writeJSON(filepath.Join(*out, "candidates.json"), selected); e != nil {
		return e
	}
	fmt.Fprintf(os.Stderr, "training_rows=%d eligible_components=%d excluded_rows=%d\n", len(train), len(selected), report.ExcludedTrainingRows)
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
