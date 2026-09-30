package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
	"os"
	"path/filepath"
)

func run() error {
	input := flag.String("input", ".cache/retrieval-audit-06.jsonl", "pinned local corpus projection")
	archives := flag.String("archives", ".cache/retrieval-source-07", "pinned source archives")
	plan := flag.String("plan", "experiments/page-claim/plan-14.json", "fixed experiment plan")
	out := flag.String("out", "", "new local output directory for numerical research heads")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require new output directory")
	}
	b, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	h := sha256.Sum256(b)
	planHash := hex.EncodeToString(h[:])
	if planHash != retrievalbench.PageClaimPlanSHA256 {
		return fmt.Errorf("plan hash mismatch")
	}
	rows, e := retrievalbench.ReadRows(*input)
	if e != nil {
		return e
	}
	selected, _, e := retrievalbench.Select(rows, *archives)
	if e != nil {
		return e
	}
	if len(selected) != 3009 {
		return fmt.Errorf("candidate count mismatch")
	}
	membership := sha256.New()
	for _, row := range selected {
		fmt.Fprintln(membership, retrievalbench.CandidateID(row))
	}
	if hex.EncodeToString(membership.Sum(nil)) != "3a6d161ac50e16f18b841645968f47af32055350e460c99a96d733ff217ff37d" {
		return fmt.Errorf("candidate membership mismatch")
	}
	examples, e := retrievalbench.PrepareBaselineFirstClaims(selected)
	if e != nil {
		return e
	}
	r, e := retrievalbench.TrainPageClaims(examples, *out, planHash)
	if e != nil {
		return e
	}
	b, e = json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	b = append(b, '\n')
	if e = os.WriteFile(filepath.Join(*out, "results.json"), b, 0600); e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Questions, Models            int
		PrimaryPassesExploratoryGate bool
	}{r.Questions, len(r.Models), r.PrimaryPassesExploratoryGate})
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
