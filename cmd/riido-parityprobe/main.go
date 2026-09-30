// riido-parityprobe compares a pinned Python reference with native Go execution.
// This maintainer diagnostic has no labels and cannot establish model quality.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/assets"
	"github.com/teamswyg/laya-tools/internal/inference"
)

type row struct {
	State, Kind, Instruction           string
	Options                            []string
	IDs, Markers                       []int64
	Probabilities, SourceProbabilities []float64
}
type reference struct {
	Schema       string
	Versions     map[string]string
	Pins         map[string]string
	Cases        []row
	Seconds      float64
	PeakRSSBytes int64
}
type report struct {
	Schema, ReferenceSHA256                                                                    string
	Cases, ExactInputMatches, GoPythonWinnerDisagreements, SourceINT8WinnerDisagreements       int
	GoPythonMaxProbabilityDelta, SourceINT8MaxProbabilityDelta, SourceINT8MeanProbabilityDelta float64
	GoPythonTolerance                                                                          float64
	ImplementationParityPassed, ModelQualityEstablished                                        bool
	Seconds                                                                                    float64
}

func distance(a, b []float64) (float64, bool, error) {
	if len(a) != 2 || len(b) != 2 {
		return 0, false, fmt.Errorf("need two probabilities")
	}
	for _, p := range [][]float64{a, b} {
		for _, v := range p {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return 0, false, fmt.Errorf("invalid probability")
			}
		}
		if math.Abs(p[0]+p[1]-1) > 1e-6 {
			return 0, false, fmt.Errorf("probabilities must sum to one")
		}
	}
	// Match argmax tie handling: index zero wins an exact tie.
	return math.Max(math.Abs(a[0]-b[0]), math.Abs(a[1]-b[1])), (a[1] > a[0]) != (b[1] > b[0]), nil
}
func run() error {
	start := time.Now()
	input := flag.String("reference", "", "reference JSON from laya_parity_reference.py")
	model := flag.String("model-dir", ".cache/models-v2/code", "pinned native model")
	runtime := flag.String("runtime", "", "pinned macOS ARM64 ONNX Runtime library")
	flag.Parse()
	f, err := os.Open(*input)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 2*1024*1024+1))
	if err != nil {
		return err
	}
	if len(data) > 2*1024*1024 {
		return fmt.Errorf("reference exceeds 2 MiB")
	}
	var ref reference
	if err = json.Unmarshal(data, &ref); err != nil {
		return err
	}
	if ref.Schema != "riido-laya-parity-reference-v1" || len(ref.Cases) != 12 {
		return fmt.Errorf("unexpected reference contract")
	}
	pinsData, err := os.ReadFile("experiments/laya-parity/assets-18.json")
	if err != nil {
		return err
	}
	var pins map[string]string
	if err = json.Unmarshal(pinsData, &pins); err != nil {
		return err
	}
	if !reflect.DeepEqual(pins, ref.Pins) {
		return fmt.Errorf("reference asset pins differ")
	}
	if err = assets.Verify(*model, map[string]string{"model.onnx": pins["onnx"], "tokenizer.json": pins["tokenizer"], "config.json": pins["config"]}); err != nil {
		return err
	}
	if err = assets.Verify(filepath.Dir(*runtime), map[string]string{filepath.Base(*runtime): pins["runtime"]}); err != nil {
		return err
	}
	enc, err := inference.LoadEncoder(filepath.Join(*model, "tokenizer.json"))
	if err != nil {
		return err
	}
	engine, err := inference.New(inference.Options{ModelDir: *model, Runtime: *runtime, Provider: "cpu", Threads: 4, MaxTokens: 512})
	if err != nil {
		return err
	}
	defer engine.Close()
	h := sha256.Sum256(data)
	out := report{Schema: "riido-laya-parity-result-v1", ReferenceSHA256: hex.EncodeToString(h[:]), Cases: len(ref.Cases), GoPythonTolerance: 1e-5}
	for _, c := range ref.Cases {
		rawRSS, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(os.Getpid())).Output()
		if err != nil {
			return err
		}
		rss, err := strconv.ParseInt(strings.TrimSpace(string(rawRSS)), 10, 64)
		if err != nil || rss < 0 {
			return fmt.Errorf("invalid sampled RSS")
		}
		if rss > 8*1024*1024 {
			return fmt.Errorf("sampled RSS guard reached")
		}
		if time.Since(start) > 15*time.Minute {
			return fmt.Errorf("time guard reached")
		}
		if c.Kind != "noul" || len(c.Options) != 2 || len(c.State) > 16384 || len(c.Instruction) > 1024 {
			return fmt.Errorf("unexpected fixture shape")
		}
		seq, err := enc.Build(c.State, c.Kind, c.Instruction, c.Options, 512)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(seq.IDs, c.IDs) && reflect.DeepEqual(seq.Markers, c.Markers) {
			out.ExactInputMatches++
		}
		p, err := engine.Predict(c.State, c.Kind, c.Instruction, c.Options)
		if err != nil {
			return err
		}
		delta, flip, err := distance(p.Probabilities, c.Probabilities)
		if err != nil {
			return err
		}
		out.GoPythonMaxProbabilityDelta = math.Max(out.GoPythonMaxProbabilityDelta, delta)
		if flip {
			out.GoPythonWinnerDisagreements++
		}
		delta, flip, err = distance(c.SourceProbabilities, c.Probabilities)
		if err != nil {
			return err
		}
		out.SourceINT8MaxProbabilityDelta = math.Max(out.SourceINT8MaxProbabilityDelta, delta)
		out.SourceINT8MeanProbabilityDelta += delta
		if flip {
			out.SourceINT8WinnerDisagreements++
		}
	}
	out.SourceINT8MeanProbabilityDelta /= float64(out.Cases)
	out.ImplementationParityPassed = out.ExactInputMatches == out.Cases && out.GoPythonMaxProbabilityDelta <= out.GoPythonTolerance && out.GoPythonWinnerDisagreements == 0
	out.Seconds = time.Since(start).Seconds()
	if err = json.NewEncoder(os.Stdout).Encode(out); err != nil {
		return err
	}
	if !out.ImplementationParityPassed {
		return fmt.Errorf("implementation parity gate failed")
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
