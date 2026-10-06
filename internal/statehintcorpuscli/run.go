// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintcorpuscli

import (
	"encoding/json"
	"errors"
	"flag"
	"io"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
)

var errOptions = errors.New("invalid corpus-check options; declare --partition and --rubric-sha256")

// Run reads one bounded partition from stdin and emits counts only. A
// structural pass is not semantic review, a fit authorization or a golden set.
func Run(args []string, in io.Reader, out, errOut io.Writer) error {
	flags := flag.NewFlagSet("corpus-check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	partition := flags.String("partition", "", "train, validation, calibration, or test")
	rubric := flags.String("rubric-sha256", "", "frozen rubric receipt digest")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err = io.WriteString(errOut, "corpus-check --partition train|validation|calibration|test --rubric-sha256 SHA < partition.jsonl\nCounts only; structure check, no semantic review, training, model or state change.\n")
			return err
		}
		return errOptions
	}
	if flags.NArg() != 0 || *partition == "" || *rubric == "" {
		return errOptions
	}
	_, summary, err := statehintcorpus.Read(in, statehintcorpus.Options{Partition: *partition, RubricSHA: *rubric})
	if err != nil {
		return statehintcorpus.ErrInput // Fixed error never echoes source or paths.
	}
	return json.NewEncoder(out).Encode(struct {
		Schema           string                  `json:"schema"`
		Status           string                  `json:"status"`
		Summary          statehintcorpus.Summary `json:"summary"`
		SemanticVerified bool                    `json:"semantic_verified"`
		TrainingAllowed  bool                    `json:"training_allowed"`
		ModelUsed        bool                    `json:"model_used"`
		MutationExecuted bool                    `json:"mutation_executed"`
	}{Schema: "riido-statehint-corpus-check-v1", Status: "structure_checked", Summary: summary})
}
