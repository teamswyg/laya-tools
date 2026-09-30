// riido-hints ranks an explicitly supplied, permission-filtered catalog.
// It provides unverified hints only and never reads repository files itself.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

type document struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
type hints struct {
	Snapshot string   `json:"snapshot_id"`
	Query    string   `json:"query"`
	IDs      []string `json:"candidate_ids"`
}
type request struct {
	Snapshot  string     `json:"snapshot_id"`
	Query     string     `json:"query"`
	Documents []document `json:"documents"`
	Hints     *hints     `json:"hints,omitempty"`
}
type candidate struct {
	ID   string  `json:"candidate_id"`
	Rank int     `json:"inspection_rank"`
	BM25 float64 `json:"bm25_score"`
}
type response struct {
	Schema     string      `json:"schema"`
	Snapshot   string      `json:"snapshot_id"`
	Status     string      `json:"status"`
	Policy     string      `json:"policy"`
	ModelHash  string      `json:"model_sha256,omitempty"`
	Candidates []candidate `json:"candidates"`
	Page       *pageInfo   `json:"page,omitempty"`
}

func run(in io.Reader, out io.Writer) error { return runModel(in, out, nil, "") }
func runModel(in io.Reader, out io.Writer, weights []float64, modelHash string) error {
	return runPolicy(in, out, weights, modelHash, false)
}

func runPolicy(in io.Reader, out io.Writer, weights []float64, modelHash string, identifierHints bool) error {
	return runPage(in, out, weights, modelHash, identifierHints, pageOptions{})
}

func runPage(in io.Reader, out io.Writer, weights []float64, modelHash string, identifierHints bool, paging pageOptions) error {
	if err := paging.validate(); err != nil {
		return err
	}
	if identifierHints && weights != nil {
		return fmt.Errorf("choose model or identifier hints, not both")
	}
	// JSON escaping can expand every source byte sixfold. Bound the encoded request
	// as well as the decoded catalog, query and identifier sizes.
	const limit = 16 << 20
	data, err := io.ReadAll(io.LimitReader(in, limit+1))
	if err != nil {
		return err
	}
	if len(data) > limit {
		return fmt.Errorf("encoded request exceeds 16 MiB")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var req request
	if err := dec.Decode(&req); err != nil {
		return fmt.Errorf("invalid request JSON: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("require exactly one JSON request")
	}
	if strings.TrimSpace(req.Snapshot) == "" || len(req.Snapshot) > 128 {
		return fmt.Errorf("require snapshot_id of 1..128 bytes")
	}
	if len(req.Documents) == 0 || len(req.Documents) > hintsearch.MaxDocuments {
		return fmt.Errorf("require 1..4096 documents")
	}
	ids := make([]string, len(req.Documents))
	texts := make([]string, len(req.Documents))
	for i, d := range req.Documents {
		if strings.TrimSpace(d.ID) == "" || len(d.ID) > 128 {
			return fmt.Errorf("require document id of 1..128 bytes")
		}
		ids[i] = d.ID
		texts[i] = d.Text
	}
	sorted := append([]string(nil), ids...)
	slices.Sort(sorted)
	if len(slices.Compact(sorted)) != len(ids) {
		return fmt.Errorf("duplicate document id")
	}
	idx, err := hintsearch.New(texts)
	if err != nil {
		return err
	}
	ranking, err := idx.Rank(req.Query)
	if err != nil {
		return err
	}
	order := ranking.Order
	policy := "bm25_complete"
	if identifierHints {
		if req.Hints != nil {
			return fmt.Errorf("choose external or identifier hints, not both")
		}
		normalized := make([]string, len(texts))
		for i, text := range texts {
			normalized[i] = lexicalhint.NormalizeText(text)
		}
		aux, auxErr := hintsearch.New(normalized)
		if auxErr == nil {
			var additional hintsearch.Ranking
			additional, auxErr = aux.Rank(lexicalhint.NormalizeText(req.Query))
			if auxErr == nil {
				order, err = hintsearch.InterleaveBaselineFirst(order, additional.Order)
				if err != nil {
					return err
				}
				policy = "identifier_bm25_baseline_first"
			}
		}
		if auxErr != nil {
			policy = "bm25_identifier_input_out_of_scope"
		}
	}
	if weights != nil {
		if req.Hints != nil {
			return fmt.Errorf("choose model or external hints, not both")
		}
		learned, _, modelErr := hintlearn.Rank(weights, req.Query, texts)
		if modelErr != nil {
			policy = "bm25_model_input_out_of_scope"
		} else {
			order, err = hintsearch.Interleave(order, learned)
			if err != nil {
				return err
			}
			policy = "model_bm25_interleave"
		}
	}
	if req.Hints != nil {
		h := req.Hints
		if h.Snapshot != req.Snapshot || h.Query != req.Query {
			return fmt.Errorf("hint snapshot or query mismatch")
		}
		if len(h.IDs) > len(ids) {
			return fmt.Errorf("too many hints")
		}
		positions := make([]int, len(h.IDs))
		for i, id := range h.IDs {
			positions[i] = slices.Index(ids, id)
			if positions[i] < 0 {
				return fmt.Errorf("unknown hint candidate")
			}
		}
		order, err = hintsearch.Interleave(order, positions)
		if err != nil {
			return err
		}
		policy = "hint_bm25_interleave"
	}
	start, end, page, err := selectPage(req, modelHash, identifierHints, policy, order, ranking.Scores, paging)
	if err != nil {
		return err
	}
	res := response{ModelHash: modelHash, Schema: "riido-hints-v1", Snapshot: req.Snapshot, Status: "unverified", Policy: policy, Candidates: make([]candidate, end-start), Page: page}
	for i := start; i < end; i++ {
		d := order[i]
		res.Candidates[i-start] = candidate{ids[d], i + 1, ranking.Scores[d]}
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
func main() {
	modelPath := flag.String("model", "", "optional research .hbin model")
	expected := flag.String("sha256", "", "required pinned model SHA-256")
	identifierHints := flag.Bool("identifier-hints", false, "experimental identifier-split hints; preserve baseline first candidate")
	pageLimit := flag.Int("limit", 0, "optional candidates per page (1..4096); 0 returns all")
	cursor := flag.String("cursor", "", "next_cursor from the same request and policy; requires --limit")
	session := flag.Bool("session", false, "JSONL session: rank once, then read cursor/limit continuations")
	flag.Parse()
	var weights []float64
	var err error
	if *modelPath != "" || *expected != "" {
		weights, err = loadModel(*modelPath, *expected)
	}
	if err == nil {
		if *session {
			if *cursor != "" {
				err = fmt.Errorf("session starts with a new query; omit --cursor")
			} else {
				err = runSession(os.Stdin, os.Stdout, weights, *expected, *identifierHints, *pageLimit)
			}
		} else {
			err = runPage(os.Stdin, os.Stdout, weights, *expected, *identifierHints, pageOptions{*pageLimit, *cursor})
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadModel(path, expected string) ([]float64, error) {
	if path == "" || len(expected) != 64 {
		return nil, fmt.Errorf("model and pinned sha256 must be provided together")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 24+hintlearn.Dimension*4+1))
	if err != nil {
		return nil, err
	}
	if hintlearn.Hash(b) != expected {
		return nil, fmt.Errorf("model SHA-256 mismatch")
	}
	return hintlearn.Decode(b)
}
