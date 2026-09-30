package filelabels

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const MaxDevelopmentBytes = 256 << 20
const MaxDevelopmentLine = 16 << 20

type DevelopmentLabel struct {
	Label
	PatchBytes  int64
	Oversized   bool
	FailureKind string
}

// ReadDevelopment verifies all bytes before decoding any labels. Members must
// come from the frozen development-only membership reader. Returned paths own
// their storage; no full patch buffer is retained between records.
func ReadDevelopment(name, expected string, members []sweaudit.TaskRole) ([]DevelopmentLabel, error) {
	return readDevelopment(name, expected, members, MaxDevelopmentBytes, MaxDevelopmentLine)
}
func readDevelopment(name, expected string, members []sweaudit.TaskRole, maxBytes, maxLine int64) ([]DevelopmentLabel, error) {
	if len(members) == 0 || len(members) > 32768 {
		return nil, fmt.Errorf("development member bound")
	}
	ids := make([]string, len(members))
	for i, m := range members {
		if (m.Role != "train" && m.Role != "validation") || strings.TrimSpace(m.Task.ID) == "" || len(m.Task.ID) > 4096 {
			return nil, fmt.Errorf("nondevelopment or invalid member")
		}
		ids[i] = m.Task.ID
	}
	slices.Sort(ids)
	for i := 1; i < len(ids); i++ {
		if ids[i] == ids[i-1] {
			return nil, fmt.Errorf("duplicate member")
		}
	}
	digest, e := hex.DecodeString(expected)
	if e != nil || len(digest) != sha256.Size {
		return nil, fmt.Errorf("invalid projection digest")
	}
	f, e := os.Open(name)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, maxBytes+1))
	if e != nil {
		return nil, e
	}
	if n > maxBytes || !bytes.Equal(h.Sum(nil), digest) {
		return nil, fmt.Errorf("development projection hash or byte bound")
	}
	if _, e = f.Seek(0, io.SeekStart); e != nil {
		return nil, e
	}
	// Recheck the bytes consumed during decoding to detect replacement/mutation
	// between the preflight hash pass and the decoding pass.
	second := sha256.New()
	scan := bufio.NewScanner(io.TeeReader(io.LimitReader(f, maxBytes+1), second))
	scan.Buffer(make([]byte, 64<<10), int(maxLine)+1)
	seen := make([]bool, len(ids))
	labels := make([]DevelopmentLabel, 0, len(ids))
	for scan.Scan() {
		if len(labels) >= len(ids) || int64(len(scan.Bytes()))+1 > maxLine {
			return nil, fmt.Errorf("development row or line bound")
		}
		var row struct {
			ID         *string `json:"instance_id"`
			Patch      *string `json:"patch"`
			PatchBytes *int64  `json:"patch_bytes"`
			Oversized  *bool   `json:"oversized"`
		}
		d := json.NewDecoder(bytes.NewReader(scan.Bytes()))
		d.DisallowUnknownFields()
		if e = d.Decode(&row); e != nil || row.ID == nil || row.PatchBytes == nil || row.Oversized == nil {
			return nil, fmt.Errorf("development label schema")
		}
		if e = d.Decode(new(any)); e != io.EOF {
			return nil, fmt.Errorf("development label trailing content")
		}
		i, ok := slices.BinarySearch(ids, *row.ID)
		if !ok || seen[i] {
			return nil, fmt.Errorf("unknown or duplicate development label")
		}
		seen[i] = true

		item := DevelopmentLabel{Label: Label{ID: *row.ID}, PatchBytes: *row.PatchBytes, Oversized: *row.Oversized}
		if item.Oversized {
			if row.Patch != nil || item.PatchBytes <= MaxPatchBytes {
				return nil, fmt.Errorf("invalid oversized metadata")
			}
			item.ParseError = true
			item.FailureKind = "oversized"
		} else {
			if row.Patch == nil || item.PatchBytes != int64(len(*row.Patch)) || item.PatchBytes > MaxPatchBytes {
				return nil, fmt.Errorf("invalid normal patch metadata")
			}
			result, parseErr := Parse(*row.Patch)
			item.Result = result
			item.ParseError = parseErr != nil
			if parseErr != nil {
				if strings.TrimSpace(*row.Patch) == "" {
					item.FailureKind = "blank_patch"
				} else {
					switch parseErr.Error() {
					case "unexpected patch preamble":
						item.FailureKind = "non_git_preamble"
					case "no Git diff blocks":
						item.FailureKind = "no_git_blocks"
					default:
						item.FailureKind = "invalid_patch"
					}
				}
			}
		}
		labels = append(labels, item)
	}
	if scan.Err() != nil {
		return nil, fmt.Errorf("development line read or byte bound")
	}
	if !bytes.Equal(second.Sum(nil), digest) {
		return nil, fmt.Errorf("development projection changed during read")
	}
	if len(labels) != len(ids) {
		return nil, fmt.Errorf("missing development labels")
	}
	slices.SortFunc(labels, func(a, b DevelopmentLabel) int { return strings.Compare(a.ID, b.ID) })
	return labels, nil
}
