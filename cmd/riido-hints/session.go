package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// readSessionLine limits allocation even for a hostile unterminated line.
func readSessionLine(r *bufio.Reader, limit int) ([]byte, error) {
	var data []byte
	for {
		part, err := r.ReadSlice('\n')
		if len(data)+len(part) > limit {
			return nil, fmt.Errorf("session message exceeds byte limit")
		}
		data = append(data, part...)
		if err == bufio.ErrBufferFull {
			continue
		}
		if err == io.EOF && len(data) > 0 {
			return data, nil
		}
		return data, err
	}
}

// One process owns one immutable query result. No shared cache, files or locks.
// The catalog is accepted only in the first line; later lines carry a cursor.
func runSession(in io.Reader, out io.Writer, weights []float64, modelHash string, identifiers bool, limit int) error {
	if limit < 1 || limit > 4096 {
		return fmt.Errorf("session requires --limit 1..4096")
	}
	r := bufio.NewReaderSize(in, 4096)
	full, err := prepareSession(r, weights, modelHash, identifiers)
	if err != nil {
		return err
	}
	digest := full.Page.RankingSHA256
	enc := json.NewEncoder(out) // One compact JSON response per input line.
	emit := func(p pageOptions) error {
		start, end, info, err := selectDigestPage(digest, len(full.Candidates), p)
		if err != nil {
			return err
		}
		page := full
		page.Candidates, page.Page = full.Candidates[start:end], info
		return enc.Encode(page)
	}
	if err := emit(pageOptions{Limit: limit}); err != nil {
		return err
	}
	for {
		line, err := readSessionLine(r, 1024)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		var next struct {
			Cursor string `json:"cursor"`
			Limit  int    `json:"limit"`
		}
		d := json.NewDecoder(bytes.NewReader(line))
		d.DisallowUnknownFields()
		if err := d.Decode(&next); err != nil {
			return fmt.Errorf("invalid continuation: %w", err)
		}
		if err := d.Decode(new(any)); err != io.EOF {
			return fmt.Errorf("extra continuation content")
		}
		if next.Cursor == "" {
			return fmt.Errorf("continuation requires cursor")
		}
		if err := emit(pageOptions{Limit: next.Limit, Cursor: next.Cursor}); err != nil {
			return err
		}
	}
}

func prepareSession(r *bufio.Reader, weights []float64, hash string, identifiers bool) (response, error) {
	var full response
	line, err := readSessionLine(r, 16<<20)
	if err != nil {
		return full, err
	}
	var buf bytes.Buffer
	if err := runPage(bytes.NewReader(line), &buf, weights, hash, identifiers, pageOptions{Limit: 4096}); err != nil {
		return full, err
	}
	err = json.Unmarshal(buf.Bytes(), &full)
	return full, err
}
