package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/retrievalbench"
	"os"
)

func run() error {
	input := flag.String("input", ".cache/retrieval-audit-06.jsonl", "pinned local projection")
	archives := flag.String("archives", ".cache/retrieval-source-07", "pinned ZIP directory")
	flag.Parse()
	rows, e := retrievalbench.ReadRows(*input)
	if e != nil {
		return e
	}
	rows, _, e = retrievalbench.Select(rows, *archives)
	if e != nil {
		return e
	}
	h := sha256.New()
	for _, r := range rows {
		fmt.Fprintln(h, retrievalbench.CandidateID(r))
	}
	if len(rows) != 3009 || hex.EncodeToString(h.Sum(nil)) != "3a6d161ac50e16f18b841645968f47af32055350e460c99a96d733ff217ff37d" {
		return fmt.Errorf("membership mismatch")
	}
	r, e := retrievalbench.ProbeInterleave(rows)
	if e != nil {
		return e
	}
	out := json.NewEncoder(os.Stdout)
	out.SetIndent("", "  ")
	return out.Encode(r)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
