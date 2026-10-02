// riido-shortclaim orders bounded candidates for subsequent verification.
// It does not load a model, decide whether an action is authorized, or discard
// candidates. All scores are unverified heuristics rather than probabilities.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

type candidateOutput struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

type output struct {
	Schema         string            `json:"schema"`
	Status         string            `json:"status"`
	Baseline       string            `json:"baseline"`
	InputSHA256    string            `json:"input_sha256"`
	FallbackReason string            `json:"fallback_reason,omitempty"`
	Candidates     []candidateOutput `json:"verification_order"`
}

func digest(p shortclaim.Prepared) string {
	h := sha256.New()
	var length [8]byte
	write := func(s string) {
		binary.BigEndian.PutUint64(length[:], uint64(len(s)))
		_, _ = h.Write(length[:])
		_, _ = io.WriteString(h, s)
	}
	write(p.Schema)
	write(p.Request)
	write(p.Provenance)
	for i := 0; i < p.Count; i++ {
		write(p.Candidates[i].ID)
		write(p.Candidates[i].Text)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func emit(in io.Reader, out io.Writer, kind string) error {
	validated, err := shortclaim.LoadValidated(in)
	if err != nil {
		return err
	}
	ranking, err := validated.Rank(kind)
	if err != nil {
		return err
	}
	p := validated.Prepared()
	r := output{Schema: "riido-shortclaim-order-v1", Status: "unverified_heuristic", Baseline: kind, InputSHA256: digest(p), FallbackReason: ranking.FallbackReason}
	r.Candidates = make([]candidateOutput, p.Count)
	for i := 0; i < ranking.Count; i++ {
		index := ranking.Order[i]
		r.Candidates[i] = candidateOutput{ID: p.Candidates[index].ID, Score: ranking.Scores[index]}
	}
	return json.NewEncoder(out).Encode(r)
}

func run(args []string, in io.Reader, out, diagnostic io.Writer) error {
	fs := flag.NewFlagSet("riido-shortclaim", flag.ContinueOnError)
	// flag's default diagnostics echo supplied values. Keep parser errors fixed;
	// only the static help text may be written to diagnostics.
	fs.SetOutput(io.Discard)
	kind := fs.String("baseline", "bm25", "fixed_order, bm25, lexical_ordered or narrow_rule")
	stream := fs.Bool("stream", false, "read bounded JSON lines; keep one process resident; stop on first invalid record")
	bench := fs.Bool("benchmark", false, "measure one original public eight-candidate fixture without a model")
	iterations := fs.Int("iterations", 2000, "benchmark repetitions per stage, 100..10000")
	profile := fs.String("cpuprofile", "", "benchmark CPU profile; LOCAL ONLY, never publish raw profiles")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(diagnostic)
			fmt.Fprintln(diagnostic, "Usage: riido-shortclaim [options] < input.json")
			fs.PrintDefaults()
			return flag.ErrHelp
		}
		return errors.New("invalid_arguments")
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected_arguments")
	}
	switch *kind {
	case "fixed_order", "bm25", "lexical_ordered", "narrow_rule":
	default:
		return errors.New("unsupported_baseline")
	}
	if *bench {
		if *stream {
			return errors.New("benchmark_stream_conflict")
		}
		return benchmark(in, out, *kind, *iterations, *profile)
	}
	if *profile != "" {
		return errors.New("profile_requires_benchmark")
	}
	if !*stream {
		return emit(in, out, *kind)
	}
	// ReadSlice caps memory even for a sender that never terminates its line.
	// A newline is framing rather than part of the 12 KiB JSON payload.
	reader := bufio.NewReaderSize(in, shortclaim.MaxJSONBytes+2)
	for {
		line, err := reader.ReadSlice('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return errors.New("stream_record_limit_or_io")
		}
		if len(line) == 0 && errors.Is(err, io.EOF) {
			return nil
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = bytes.TrimSuffix(line, []byte{'\n'})
			line = bytes.TrimSuffix(line, []byte{'\r'})
		}
		if len(line) > shortclaim.MaxJSONBytes {
			return errors.New("stream_record_limit")
		}
		if len(line) == 0 {
			return errors.New("empty_stream_record")
		}
		if e := emit(bytes.NewReader(line), out, *kind); e != nil {
			return e
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
	}
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
