// riido-sourceprobe prepares unlabeled validation requests and verifies/replays
// original-weight reference scores. It never runs a model or exports labels.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"

	"github.com/teamswyg/laya-tools/internal/assets"
	"github.com/teamswyg/laya-tools/internal/inference"
	"github.com/teamswyg/laya-tools/internal/layaprobe"
	"github.com/teamswyg/laya-tools/internal/paireval"
)

const planSHA = "4eb07b72f5029f8313dce98663d1dffb79b398c7685f70e9225a21b62b20399e"
const modelSHA = "e759adc16a499aef407d4d5672b912ad20baf15d44248e835b3ee3ff5da91b4a"
const membershipSHA = "d90eda8e8a423eff44847ff46ef19b468a57ac2791d45c57e7e8ba39a4ee9902"

type request struct {
	State, Kind, Instruction string
	Options                  []string
	IDs, Markers             []int64
	Truncated                bool
}
type bundle struct {
	Schema, PlanSHA256, ModelSHA256, SourceSHA256, MembershipSHA256 string
	Cases                                                           []request
}
type response struct {
	RequestSHA256 string
	Prediction    inference.Prediction
}
type scored struct {
	Schema, RequestsSHA256, PlanSHA256, ModelSHA256 string
	Cases                                           []response
	TokenParityMatches                              int
	Seconds                                         float64
	PeakRSSBytes                                    int64
	Versions                                        map[string]string
}

func hash(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return paireval.Hash(b)
}
func bounded(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, (16<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 16<<20 {
		return nil, fmt.Errorf("file exceeds 16 MiB")
	}
	return b, nil
}

type builder interface {
	Build(string, string, string, []string, int) (inference.Sequence, error)
}

func prepare(rows []paireval.Row, split paireval.Split, enc builder) (bundle, error) {
	out := bundle{Schema: "riido-source-requests-v1", PlanSHA256: planSHA, ModelSHA256: modelSHA, SourceSHA256: paireval.SourceSHA, MembershipSHA256: split.MembershipSHA256}
	if len(rows) != len(split.Rows) {
		return out, fmt.Errorf("split length mismatch")
	}
	for i, row := range rows {
		if split.Rows[i].Split != "validation" || !paireval.InScope(row) {
			continue
		}
		options := []string{"false: no, the statement does not hold", "true: yes, the statement holds"}
		for order := 0; order < 2; order++ {
			req := request{State: "File: [unavailable]\n" + row.Code, Kind: "noul", Instruction: "Is this source code relevant to the software change: \"" + row.Query + "\"?", Options: append([]string(nil), options...)}
			seq, err := enc.Build(req.State, req.Kind, req.Instruction, req.Options, 512)
			if err != nil {
				return out, err
			}
			req.IDs, req.Markers, req.Truncated = seq.IDs, seq.Markers, seq.Truncated
			out.Cases = append(out.Cases, req)
			options[0], options[1] = options[1], options[0]
		}
	}
	return out, nil
}

type replay struct {
	requests []request
	scores   []response
	at       int
}

func (r *replay) Predict(state, kind, instruction string, options []string) (inference.Prediction, error) {
	if r.at >= len(r.requests) || r.at >= len(r.scores) {
		return inference.Prediction{}, fmt.Errorf("score sequence exhausted")
	}
	req := r.requests[r.at]
	s := r.scores[r.at]
	if state != req.State || kind != req.Kind || instruction != req.Instruction || !reflect.DeepEqual(options, req.Options) || s.RequestSHA256 != hash(req) {
		return inference.Prediction{}, fmt.Errorf("request identity mismatch")
	}
	p := s.Prediction
	if p.Truncated != req.Truncated || p.Tokens != len(req.IDs) || math.IsNaN(p.MS) || math.IsInf(p.MS, 0) || p.MS < 0 {
		return p, fmt.Errorf("prediction metadata mismatch")
	}
	r.at++
	return p, nil
}
func run() error {
	stage := flag.String("stage", "prepare", "prepare or evaluate")
	input := flag.String("input", ".cache/semantic-scale/cosqa-all.json", "pinned CoSQA source")
	model := flag.String("model-dir", ".cache/models-v2/code", "pinned tokenizer directory")
	requests := flag.String("requests", ".cache/source19-requests.json", "local unlabeled requests")
	scores := flag.String("scores", ".cache/source19-scores.json", "local original-weight predictions")
	flag.Parse()
	plan, err := os.ReadFile("experiments/laya-original/plan-19.json")
	if err != nil {
		return err
	}
	if paireval.Hash(plan) != planSHA {
		return fmt.Errorf("plan mismatch")
	}
	pinBytes, err := os.ReadFile("experiments/laya-parity/assets-18.json")
	if err != nil {
		return err
	}
	if paireval.Hash(pinBytes) != "fa8193110a1e4f4332176f901798c3cf35e839f9044b03e4fc9a71c99d6bdac9" {
		return fmt.Errorf("asset manifest mismatch")
	}
	var pins map[string]string
	if err = json.Unmarshal(pinBytes, &pins); err != nil {
		return err
	}
	if err = assets.Verify(*model, map[string]string{"tokenizer.json": pins["tokenizer"]}); err != nil {
		return err
	}
	rows, err := paireval.Load(*input)
	if err != nil {
		return err
	}
	split := paireval.Partition(rows)
	if split.MembershipSHA256 != membershipSHA || split.Counts["validation"] != 607 {
		return fmt.Errorf("membership mismatch")
	}
	enc, err := inference.LoadEncoder(filepath.Join(*model, "tokenizer.json"))
	if err != nil {
		return err
	}
	req, err := prepare(rows, split, enc)
	if err != nil {
		return err
	}
	if len(req.Cases) != 1106 {
		return fmt.Errorf("unexpected request count")
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if *stage == "prepare" {
		f, err := os.OpenFile(*requests, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		_, err = f.Write(encoded)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		return closeErr
	}
	if *stage != "evaluate" {
		return fmt.Errorf("unknown stage")
	}
	frozen, err := bounded(*requests)
	if err != nil {
		return err
	}
	if string(frozen) != string(encoded) {
		return fmt.Errorf("prepared requests changed")
	}
	raw, err := bounded(*scores)
	if err != nil {
		return err
	}
	var results scored
	if err = json.Unmarshal(raw, &results); err != nil {
		return err
	}
	if results.Schema != "riido-source-scores-v1" || results.PlanSHA256 != planSHA || results.ModelSHA256 != modelSHA || results.RequestsSHA256 != paireval.Hash(encoded) || len(results.Cases) != len(req.Cases) || results.TokenParityMatches != len(req.Cases) {
		return fmt.Errorf("score provenance mismatch")
	}
	replay := &replay{requests: req.Cases, scores: results.Cases}
	report, err := layaprobe.RunProtocol(rows, split, replay, func() error { return nil }, func(int, int) {}, "noul")
	if err != nil {
		return err
	}
	if replay.at != len(req.Cases) {
		return fmt.Errorf("unused predictions")
	}
	report.Schema = "riido-original-relevance-v1"
	report.PlanSHA256 = planSHA
	report.ModelSHA256 = modelSHA
	return json.NewEncoder(os.Stdout).Encode(struct {
		Report                       layaprobe.Report
		RequestsSHA256, ScoresSHA256 string
		ReferenceSeconds             float64
		ReferencePeakRSSBytes        int64
		TokenParityMatches           int
		Versions                     map[string]string
		TimingNote                   string
	}{report, paireval.Hash(encoded), paireval.Hash(raw), results.Seconds, results.PeakRSSBytes, results.TokenParityMatches, results.Versions, "Report.ProbeSeconds is Go replay/evaluation time; per-call times and ReferenceSeconds are Python CPU reference measurements"})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
