package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func run() error {
	full := flag.String("full", ".cache/real-task-full-28.jsonl", "pinned local projection")
	multi := flag.String("multilingual", ".cache/real-task-multilingual-28.jsonl", "pinned local projection")
	plan := flag.String("plan", "experiments/evaluation-groups/plan-29.json", "fixed grouping plan")
	out := flag.String("out", "", "new private local output directory; selected.json must not be published")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("require new output directory")
	}
	b, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	h := sha256.Sum256(b)
	if hex.EncodeToString(h[:]) != sweaudit.GroupsPlanSHA256 {
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
		return fmt.Errorf("source row counts")
	}
	report, selected, e := sweaudit.GroupEvaluation(a, m)
	if e != nil {
		return e
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	for i, obj := range []any{report, selected} {
		b, e = json.MarshalIndent(obj, "", "  ")
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(*out, []string{"results.json", "selected.json"}[i]), append(b, '\n'), 0600); e != nil {
			return e
		}
	}
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
