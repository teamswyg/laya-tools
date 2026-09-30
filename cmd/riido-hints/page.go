package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

type pageOptions struct {
	Limit  int
	Cursor string
}
type pageInfo struct {
	Total         int    `json:"total_candidates"`
	Offset        int    `json:"offset"`
	Returned      int    `json:"returned"`
	RankingSHA256 string `json:"ranking_sha256"`
	NextCursor    string `json:"next_cursor,omitempty"`
}
type pageCursor struct {
	Version int    `json:"v"`
	Offset  int    `json:"o"`
	Hash    string `json:"h"`
}

func (p pageOptions) validate() error {
	if p.Limit < 0 || p.Limit > hintsearch.MaxDocuments || len(p.Cursor) > 256 || (p.Cursor != "" && p.Limit == 0) {
		return fmt.Errorf("require --limit 1..4096 with cursor; omit both for all candidates")
	}
	return nil
}

// The digest binds contents, query, snapshot, policy/configuration, complete
// ordering and scores. It is a consistency check, not authentication or access
// control. No request data or server-side cursor state is stored.
func selectPage(req request, modelHash string, identifiers bool, policy string, order []int, scores []float64, p pageOptions) (int, int, *pageInfo, error) {
	if e := p.validate(); e != nil {
		return 0, 0, nil, e
	}
	if p.Limit == 0 {
		return 0, len(order), nil, nil
	}
	h := sha256.New()
	e := json.NewEncoder(h).Encode(struct {
		Version     string
		Request     request
		ModelHash   string
		Identifiers bool
		Policy      string
		Order       []int
		Scores      []float64
	}{"riido-hints-page-v1", req, modelHash, identifiers, policy, order, scores})
	if e != nil {
		return 0, 0, nil, e
	}
	digest := hex.EncodeToString(h.Sum(nil))
	return selectDigestPage(digest, len(order), p)
}

func selectDigestPage(digest string, total int, p pageOptions) (int, int, *pageInfo, error) {
	if err := p.validate(); err != nil || p.Limit == 0 {
		return 0, 0, nil, fmt.Errorf("require a valid page limit")
	}
	start := 0
	if p.Cursor != "" {
		b, e := base64.RawURLEncoding.DecodeString(p.Cursor)
		if e != nil {
			return 0, 0, nil, fmt.Errorf("invalid cursor encoding")
		}
		var c pageCursor
		d := json.NewDecoder(bytes.NewReader(b))
		d.DisallowUnknownFields()
		if e := d.Decode(&c); e != nil {
			return 0, 0, nil, fmt.Errorf("invalid cursor")
		}
		if e := d.Decode(new(any)); e != io.EOF {
			return 0, 0, nil, fmt.Errorf("extra cursor content")
		}
		if c.Version != 1 || c.Hash != digest || c.Offset <= 0 || c.Offset >= total {
			return 0, 0, nil, fmt.Errorf("cursor does not match current request, ranking or position")
		}
		start = c.Offset
	}
	end := min(start+p.Limit, total)
	info := &pageInfo{Total: total, Offset: start, Returned: end - start, RankingSHA256: digest}
	if end < total {
		b, e := json.Marshal(pageCursor{1, end, digest})
		if e != nil {
			return 0, 0, nil, e
		}
		info.NextCursor = base64.RawURLEncoding.EncodeToString(b)
	}
	return start, end, info, nil
}
