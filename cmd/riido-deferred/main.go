package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/researchbundle"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
	"os"
	"path/filepath"
)

func run() error {
	input := flag.String("input", ".cache/retrieval-audit-06.jsonl", "pinned local corpus projection")
	archives := flag.String("archives", ".cache/retrieval-source-07", "pinned source archives")
	plan := flag.String("plan", "experiments/deferred-helper/plan-27.json", "fixed experiment plan")
	previous := flag.String("previous", ".cache/spread-claim-hf-verified-21", "verified experiment21 archive")
	current := flag.String("current", ".cache/coverage-claim-hf-verified-25", "verified experiment25 archive")
	out := flag.String("out", "", "new local output directory for aggregate deferred-helper experiment")
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
	if planHash != retrievalbench.DeferredPlanSHA256 {
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
	examples, extra, coverage, deferred, e := retrievalbench.PrepareDeferredClaims(selected)
	if e != nil {
		return e
	}
	old, e := loadPolicy(*previous, retrievalbench.ShallowControlManifestSHA256)
	if e != nil {
		return e
	}
	next, e := loadPolicy(*current, retrievalbench.Coverage25ManifestSHA256)
	if e != nil {
		return e
	}
	r, e := retrievalbench.ProbeDeferredHelpers(examples, extra, coverage, deferred, old, next, planHash)
	if e != nil {
		return e
	}
	if e = os.Mkdir(*out, 0700); e != nil {
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
		Questions, Policies int
	}{r.Questions, len(r.Aggregates)})
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

func loadPolicy(dir, wantHash string) (retrievalbench.FrozenClaimPolicy, error) {
	var p retrievalbench.FrozenClaimPolicy
	b, e := os.ReadFile(filepath.Join(dir, "MANIFEST.json"))
	if e != nil {
		return p, e
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != wantHash {
		return p, fmt.Errorf("archive manifest pin mismatch")
	}
	if _, e = researchbundle.Verify(dir); e != nil {
		return p, e
	}
	b, e = os.ReadFile(filepath.Join(dir, "results.json"))
	if e != nil {
		return p, e
	}
	if e = json.Unmarshal(b, &p.Report); e != nil {
		return p, e
	}
	for _, f := range p.Report.Folds {
		if f.Policy != "cost_selected" {
			continue
		}
		b, e = os.ReadFile(filepath.Join(dir, f.File))
		if e != nil {
			return p, e
		}
		var head retrievalbench.CostClaimHead
		if e = json.Unmarshal(b, &head); e != nil {
			return p, e
		}
		p.Heads = append(p.Heads, head)
	}
	return p, nil
}
