package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/teamswyg/laya-tools/internal/retrievalbench"
)

func run() error {
	input := flag.String("input", ".cache/retrieval-audit-06.jsonl", "pinned local corpus projection")
	archives := flag.String("archives", ".cache/retrieval-source-07", "pinned source ZIP directory")
	stage := flag.String("stage", "audit", "audit or evaluate")
	plan := flag.String("plan", "experiments/retrieval-baseline/plan-07.json", "fixed evaluation plan")
	flag.Parse()
	if *stage != "audit" && *stage != "evaluate" {
		return fmt.Errorf("invalid stage")
	}
	planBytes, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	planHash := sha256.Sum256(planBytes)
	if hex.EncodeToString(planHash[:]) != "1621f0bcf83d161afb63f9478c6a9063931565ec803746564bef4fe4ea330c32" {
		return fmt.Errorf("plan hash mismatch")
	}
	rows, e := retrievalbench.ReadRows(*input)
	if e != nil {
		return e
	}
	selected, audit, e := retrievalbench.Select(rows, *archives)
	if e != nil {
		return e
	}
	h := sha256.New()
	for _, r := range selected {
		fmt.Fprintln(h, retrievalbench.CandidateID(r))
	}
	membership := hex.EncodeToString(h.Sum(nil))
	if len(selected) != 3009 || membership != "3a6d161ac50e16f18b841645968f47af32055350e460c99a96d733ff217ff37d" {
		return fmt.Errorf("membership mismatch")
	}
	out := struct {
		Schema, Stage, MembershipSHA256 string
		Rows                            int
		Sources                         []retrievalbench.SourceAudit
		Results                         []retrievalbench.Result
		ProductionReady                 bool
	}{Schema: "riido-retrieval-proxy-v1", Stage: *stage, MembershipSHA256: membership, Rows: len(selected), Sources: audit}
	if *stage == "evaluate" {
		for _, mode := range []string{"raw", "identifiers"} {
			r, e := retrievalbench.Evaluate(selected, mode)
			if e != nil {
				return e
			}
			if r.All.Queries != 2948 {
				return fmt.Errorf("unexpected distinct question count")
			}
			out.Results = append(out.Results, r)
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
