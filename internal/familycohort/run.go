package familycohort

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/teamswyg/laya-tools/internal/reviewpacket"
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func encode(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, Error("output_json")
	}
	return append(b, '\n'), nil
}

func publish(dir, name string, data []byte) error {
	tmp := filepath.Join(dir, name+".pending")
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return Error("output_create")
	}
	n, we := f.Write(data)
	se, ce := f.Sync(), f.Close()
	if we != nil || se != nil || ce != nil || n != len(data) {
		return Error("output_write")
	}
	if os.Link(tmp, filepath.Join(dir, name)) != nil || os.Remove(tmp) != nil {
		return Error("output_publish")
	}
	d, err := os.Open(dir)
	if err != nil {
		return Error("output_directory")
	}
	se, ce = d.Sync(), d.Close()
	if se != nil || ce != nil {
		return Error("output_sync")
	}
	return nil
}

func diskBytes(dir string) (int64, error) {
	var total int64
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return Error("output_accounting")
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return Error("output_accounting")
			}
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// Run captures the one independently pinned bundle exactly once, then checks
// its structure. References inside it are never opened. Successful private
// output retains the captured bundle bytes unchanged, not normalized wording.
func Run(root string, input File, out string) (s Summary, err error) {
	s = initial()
	if !validFile(input) || input.Bytes == 0 || input.Bytes > MaxInputBytes {
		s.Code = "input_pin"
		return s, Error(s.Code)
	}
	root, err = filepath.Abs(root)
	if err == nil {
		root, err = filepath.EvalSymlinks(root)
	}
	if err != nil {
		s.Code = "input_root"
		return s, Error(s.Code)
	}
	inputPath := filepath.Join(root, filepath.FromSlash(input.Path))
	parent, pe := filepath.EvalSymlinks(filepath.Dir(inputPath))
	within, re := filepath.Rel(root, parent)
	if pe != nil || re != nil || !filepath.IsLocal(within) {
		s.Code = "input_root"
		return s, Error(s.Code)
	}
	if os.Mkdir(out, 0700) != nil {
		s.Code = "new_output_required"
		return s, Error(s.Code)
	}
	defer func() {
		if err != nil {
			s.State = "blocked"
			s.Code = err.Error()
		}
		b, e := encode(s)
		if e == nil {
			used, ae := diskBytes(out)
			if ae != nil || used+int64(len(b)) > MaxOutputBytes {
				e = Error("output_bytes")
			} else {
				e = publish(out, "SUMMARY.json", b)
			}
		}
		if e != nil {
			err = e
			s.State, s.Code = "blocked", e.Error()
		}
	}()
	if os.Mkdir(filepath.Join(out, "receipts"), 0700) != nil {
		return s, Error("output_directory")
	}
	// Preflight the exact input and maximum artifact plus receipt/summary reserve
	// before opening the supplied content. No partial corpus is selected.
	if input.Bytes+MaxArtifactBytes+(64<<10) > MaxOutputBytes {
		return s, Error("output_bytes")
	}
	raw, outcome, captureErr := reviewpacket.Capture(reviewpacket.Config{
		InputPath: inputPath, ReceiptPath: filepath.Join(out, "receipts", "000001.start.json"),
		ExpectedSHA256: input.SHA256, MaxBytes: MaxInputBytes, Actor: "familycohort-adapter",
	})
	if captureErr != nil || !outcome.StartDurable || !outcome.ResultDurable {
		return s, Error("capture_failed")
	}
	s.InputBytes = int64(len(raw))
	if s.InputBytes != input.Bytes || sourcecohort.Hash(raw) != input.SHA256 {
		return s, Error("input_bytes")
	}
	bundle, decodeErr := Decode(raw)
	if decodeErr != nil {
		return s, decodeErr
	}
	s, err = Validate(bundle)
	s.InputBytes = int64(len(raw))
	if err != nil {
		return s, err
	}
	retained := File{Path: "BUNDLE.private.json", SHA256: input.SHA256, Bytes: input.Bytes}
	artifact, e := encode(Structure{"riido-familycohort-structure-v1", input, retained, s})
	if e != nil {
		return s, e
	}
	used, e := diskBytes(out)
	if e != nil || len(artifact) > MaxArtifactBytes || len(raw) > MaxArtifactBytes || used+int64(len(raw)+len(artifact))+(64<<10) > MaxOutputBytes {
		return s, Error("output_bytes")
	}
	if e := publish(out, "BUNDLE.private.json", raw); e != nil {
		return s, e
	}
	if e := publish(out, "STRUCTURE.private.json", artifact); e != nil {
		return s, e
	}
	return s, nil
}
