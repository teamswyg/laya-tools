package inference

import (
	"encoding/json"
	"fmt"
	ort "github.com/yalue/onnxruntime_go"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Options struct {
	ModelDir, Runtime, Provider, Profile string
	Threads, MaxTokens                   int
}
type Engine struct {
	mu      sync.Mutex
	session *ort.DynamicAdvancedSession
	enc     *Encoder
	cfg     config
	opts    Options
	LoadMS  float64
}
type config struct {
	Temperature []float64          `json:"temperature"`
	ByOptions   map[string]float64 `json:"temperature_by_options"`
}
type Prediction struct {
	Probabilities []float64 `json:"probabilities"`
	Winner        int       `json:"winner"`
	Truncated     bool      `json:"truncated"`
	Tokens        int       `json:"tokens"`
	MS            float64   `json:"ms"`
}

var envOnce sync.Once
var envErr error

func New(o Options) (*Engine, error) {
	start := time.Now()
	if o.Threads < 1 || o.Threads > 64 {
		return nil, fmt.Errorf("threads must be 1..64")
	}
	if o.MaxTokens == 0 {
		o.MaxTokens = 512
	}
	if o.Provider != "cpu" && o.Provider != "coreml" {
		return nil, fmt.Errorf("provider must be cpu or coreml")
	}
	envOnce.Do(func() { ort.SetSharedLibraryPath(o.Runtime); envErr = ort.InitializeEnvironment() })
	if envErr != nil {
		return nil, fmt.Errorf("load native runtime: %w (run laya setup)", envErr)
	}
	enc, err := LoadEncoder(filepath.Join(o.ModelDir, "tokenizer.json"))
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(o.ModelDir, "config.json"))
	if err != nil {
		return nil, err
	}
	var cfg config
	if err = json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Temperature) != 3 {
		return nil, fmt.Errorf("invalid model temperature config")
	}
	so, err := ort.NewSessionOptions()
	if err != nil {
		return nil, err
	}
	defer so.Destroy()
	for _, err = range []error{so.SetIntraOpNumThreads(o.Threads), so.SetInterOpNumThreads(1), so.SetGraphOptimizationLevel(ort.GraphOptimizationLevelEnableAll), so.SetCpuMemArena(false), so.SetMemPattern(false)} {
		if err != nil {
			return nil, err
		}
	}
	if o.Profile != "" {
		if err = so.EnableProfiling(o.Profile); err != nil {
			return nil, err
		}
	}
	if o.Provider == "coreml" {
		if err = so.AppendExecutionProviderCoreMLV2(map[string]string{"ModelFormat": "MLProgram", "MLComputeUnits": "ALL", "RequireStaticInputShapes": "0"}); err != nil {
			return nil, err
		}
	}
	session, err := ort.NewDynamicAdvancedSession(filepath.Join(o.ModelDir, "model.onnx"), []string{"input_ids", "attention_mask", "marker_pos", "marker_mask", "qtype"}, []string{"logits", "act_logits"}, so)
	if err != nil {
		return nil, err
	}
	return &Engine{session: session, enc: enc, cfg: cfg, opts: o, LoadMS: float64(time.Since(start).Microseconds()) / 1000}, nil
}
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.session == nil {
		return nil
	}
	err := e.session.Destroy()
	e.session = nil
	return err
}
func (e *Engine) Predict(state, kind, instruction string, options []string) (Prediction, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	start := time.Now()
	seq, err := e.enc.Build(state, kind, instruction, options, e.opts.MaxTokens)
	if err != nil {
		return Prediction{}, err
	}
	q := int64(0)
	if kind == "noul" {
		q = 2
	} else if kind != "choice" {
		return Prediction{}, fmt.Errorf("unsupported question type")
	}
	var inputs []ort.Value
	defer func() {
		for _, v := range inputs {
			v.Destroy()
		}
	}()
	addInt := func(shape ort.Shape, data []int64) error {
		t, err := ort.NewTensor(shape, data)
		if err == nil {
			inputs = append(inputs, t)
		}
		return err
	}
	mask := make([]int64, len(seq.IDs))
	for i := range mask {
		mask[i] = 1
	}
	for _, x := range []struct {
		s ort.Shape
		d []int64
	}{{ort.NewShape(1, int64(len(seq.IDs))), seq.IDs}, {ort.NewShape(1, int64(len(mask))), mask}, {ort.NewShape(1, int64(len(options))), seq.Markers}} {
		if err = addInt(x.s, x.d); err != nil {
			return Prediction{}, err
		}
	}
	mm := make([]bool, len(options))
	for i := range mm {
		mm[i] = true
	}
	t, err := ort.NewTensor(ort.NewShape(1, int64(len(options))), mm)
	if err != nil {
		return Prediction{}, err
	}
	inputs = append(inputs, t)
	if err = addInt(ort.NewShape(1), []int64{q}); err != nil {
		return Prediction{}, err
	}
	outputs := make([]ort.Value, 2)
	defer func() {
		for _, v := range outputs {
			if v != nil {
				v.Destroy()
			}
		}
	}()
	if err = e.session.Run(inputs, outputs); err != nil {
		return Prediction{}, err
	}
	logits, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return Prediction{}, fmt.Errorf("expected float32 logits")
	}
	data := logits.GetData()
	if len(data) != len(options) {
		return Prediction{}, fmt.Errorf("unexpected logits shape")
	}
	bucket := "2"
	if len(options) >= 3 {
		bucket = "3-5"
	}
	if len(options) >= 6 {
		bucket = "6-10"
	}
	temp := e.cfg.Temperature[q]
	if v, ok := e.cfg.ByOptions[kind+":"+bucket]; ok {
		temp = v
	}
	if math.IsNaN(temp) || math.IsInf(temp, 0) {
		temp = 1
	}
	temp = math.Max(.5, math.Min(5, temp))
	probs := make([]float64, len(data))
	maxLog := float64(data[0])
	winner := 0
	for i, v := range data {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return Prediction{}, fmt.Errorf("nonfinite logits")
		}
		if float64(v) > maxLog {
			maxLog = float64(v)
			winner = i
		}
	}
	sum := 0.0
	for i, v := range data {
		probs[i] = math.Exp((float64(v) - maxLog) / temp)
		sum += probs[i]
	}
	for i := range probs {
		probs[i] /= sum
	}
	return Prediction{Probabilities: probs, Winner: winner, Truncated: seq.Truncated, Tokens: len(seq.IDs), MS: float64(time.Since(start).Microseconds()) / 1000}, nil
}
