package filelabels

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const FullProjectionSHA256 = "a58d19a6e4f3d6f062deeb1dba61630c8f26d739a51bf5d3786f9cff123e096d"
const MultilingualProjectionSHA256 = "be00d9ef6584dbc82594eee83ab41d21049266312915d60854f1082608fa3576"

type Label struct {
	ID         string
	Result     Result
	ParseError bool
}

// Read verifies the exact projection before exposing any evaluation labels.
func Read(name, expected string) ([]Label, error) {
	f, e := os.Open(name)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, 32<<20+1))
	if e != nil {
		return nil, e
	}
	if len(data) > 32<<20 {
		return nil, fmt.Errorf("label projection bound")
	}
	h := sha256.Sum256(data)
	if hex.EncodeToString(h[:]) != expected {
		return nil, fmt.Errorf("label projection hash mismatch")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var labels []Label
	for {
		var row struct {
			ID    string `json:"instance_id"`
			Patch string `json:"patch"`
		}
		e := dec.Decode(&row)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, fmt.Errorf("label projection schema")
		}
		if strings.TrimSpace(row.ID) == "" || len(labels) >= 4096 {
			return nil, fmt.Errorf("label identity or row bound")
		}
		result, e := Parse(row.Patch)
		labels = append(labels, Label{row.ID, result, e != nil})
	}
	slices.SortFunc(labels, func(a, b Label) int { return strings.Compare(a.ID, b.ID) })
	for i := 1; i < len(labels); i++ {
		if labels[i].ID == labels[i-1].ID {
			return nil, fmt.Errorf("duplicate label identity")
		}
	}
	return labels, nil
}
