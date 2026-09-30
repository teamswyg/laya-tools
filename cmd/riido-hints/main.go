// riido-hints ranks an explicitly supplied, permission-filtered catalog.
// It provides unverified hints only and never reads repository files itself.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

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
	Candidates []candidate `json:"candidates"`
}

func run(in io.Reader, out io.Writer) error {
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
	dec := json.NewDecoder(strings.NewReader(string(data)))
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
	res := response{Schema: "riido-hints-v1", Snapshot: req.Snapshot, Status: "unverified", Policy: policy, Candidates: make([]candidate, len(order))}
	for i, d := range order {
		res.Candidates[i] = candidate{ids[d], i + 1, ranking.Scores[d]}
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
