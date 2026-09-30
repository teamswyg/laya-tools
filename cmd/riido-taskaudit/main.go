package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/laya-tools/internal/retrievalbench"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func run() error {
	full := flag.String("full", ".cache/real-task-full-28.jsonl", "pinned local full request projection")
	multi := flag.String("multilingual", ".cache/real-task-multilingual-28.jsonl", "pinned local multilingual request projection")
	prior := flag.String("prior", ".cache/retrieval-audit-06.jsonl", "pinned previous CodeSearchNet projection")
	plan := flag.String("plan", "experiments/real-task-audit/plan-28.json", "fixed audit plan")
	out := flag.String("out", "", "new local output directory")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require new output directory")
	}
	b, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != sweaudit.PlanSHA256 {
		return fmt.Errorf("plan hash mismatch")
	}
	a, e := sweaudit.Read(*full, sweaudit.FullProjectionSHA256)
	if e != nil {
		return e
	}
	m, e := sweaudit.Read(*multi, sweaudit.MultilingualProjectionSHA256)
	if e != nil {
		return e
	}
	if len(a) != 2294 || len(m) != 300 {
		return fmt.Errorf("pinned row count mismatch")
	}
	old, e := retrievalbench.ReadRows(*prior)
	if e != nil {
		return e
	}
	requests, repos := make([]string, len(old)), make([]string, len(old))
	for i, r := range old {
		requests[i] = r.Query
		repos[i] = r.Repository
	}
	report := sweaudit.Analyze(a, m, requests, repos)
	b, e = json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	return os.WriteFile(filepath.Join(*out, "results.json"), append(b, '\n'), 0600)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
