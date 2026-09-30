package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/conalaaudit"
	"os"
	"path/filepath"
)

func run() error {
	root := flag.String("source", ".cache/conala-16/data", "pinned curated JSONL directory")
	plan := flag.String("plan", "experiments/conala-audit/plan-16.json", "fixed audit plan")
	flag.Parse()
	b, err := os.ReadFile(*plan)
	if err != nil {
		return err
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != conalaaudit.PlanSHA256 {
		return fmt.Errorf("plan mismatch")
	}
	train, err := conalaaudit.Read(filepath.Join(*root, "conala-paired-train.json"), conalaaudit.TrainSHA256)
	if err != nil {
		return err
	}
	test, err := conalaaudit.Read(filepath.Join(*root, "conala-paired-test.json"), conalaaudit.TestSHA256)
	if err != nil {
		return err
	}
	if len(train) != 2379 || len(test) != 500 {
		return fmt.Errorf("source row count mismatch")
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(conalaaudit.Audit(train, test))
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
