package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/laya-tools/internal/researchbundle"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
)

func run() error {
	input := flag.String("input", ".cache/retrieval-audit-06.jsonl", "pinned source projection")
	archives := flag.String("archives", ".cache/retrieval-source-07", "pinned source ZIPs")
	models := flag.String("models", ".cache/page-claim-hf-verified-14", "verified frozen experiment-14 package")
	plan := flag.String("plan", "experiments/page-budget/plan-15.json", "fixed calibration plan")
	flag.Parse()
	for _, p := range []struct{ path, hash string }{{*plan, retrievalbench.PageBudgetPlanSHA256}, {filepath.Join(*models, "MANIFEST.json"), retrievalbench.PageBudgetManifestSHA256}} {
		b, err := os.ReadFile(p.path)
		if err != nil {
			return err
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != p.hash {
			return fmt.Errorf("pinned input hash mismatch")
		}
	}
	if _, err := researchbundle.Verify(*models); err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Join(*models, "results.json"))
	if err != nil {
		return err
	}
	var trained retrievalbench.PageClaimReport
	if err := json.Unmarshal(b, &trained); err != nil {
		return err
	}
	var heads []retrievalbench.BudgetHead
	for _, m := range trained.Models {
		b, err := os.ReadFile(filepath.Join(*models, m.File))
		if err != nil {
			return err
		}
		var h retrievalbench.ClaimHead
		if err := json.Unmarshal(b, &h); err != nil {
			return err
		}
		heads = append(heads, retrievalbench.BudgetHead{File: m.File, SHA256: m.SHA256, Head: h})
	}
	rows, err := retrievalbench.ReadRows(*input)
	if err != nil {
		return err
	}
	selected, _, err := retrievalbench.Select(rows, *archives)
	if err != nil {
		return err
	}
	h := sha256.New()
	for _, row := range selected {
		fmt.Fprintln(h, retrievalbench.CandidateID(row))
	}
	if len(selected) != 3009 || hex.EncodeToString(h.Sum(nil)) != "3a6d161ac50e16f18b841645968f47af32055350e460c99a96d733ff217ff37d" {
		return fmt.Errorf("candidate membership mismatch")
	}
	examples, err := retrievalbench.PrepareBaselineFirstClaims(selected)
	if err != nil {
		return err
	}
	report, err := retrievalbench.ProbePageBudgets(examples, heads)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
