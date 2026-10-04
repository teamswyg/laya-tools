// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 teamswyg contributors.
// Finite exposed development replay, not training or independent Golden data.
// Only fixed public packet files are copied. Startup/internal upstream calls are
// outside the observer ledger; GOMEMLIMIT is a soft Go heap target, not RSS.
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type pin struct {
	name string
	size int
	sha  string
	dest string
}

var pins = [...]pin{
	{"source/decimal/const.go.txt", 4105, "08bfb8421064cf5c2cfd84e53b15e8f3c76ddd629e0b4f148d935aba46f2d441", "decimal/const.go"},
	{"source/decimal/decimal-go.go.txt", 11299, "e3da06b81105dd26de72443f5980242cd821458d65d5558b44891d1a108cd97f", "decimal/decimal-go.go"},
	{"source/decimal/decimal.go.txt", 67565, "7b47645b284bcd9d85b8de8086c001c30cb9ab10773c0e9d9f9db40c1ba9d863", "decimal/decimal.go"},
	{"source/decimal/rounding.go.txt", 5191, "1db1de0fd77a66d2ed5282ef4c95fcb42967c1426dd7cc53ce0c9159978b6d1a", "decimal/rounding.go"},
	{"source/decimal/go.mod.txt", 46, "fc8a8e1fa14e49884ec9f962563a4ef51fa0a9bbfb7dbcb2ec8958e1f00988c9", "decimal/go.mod"},
	{"source/decimal/LICENSE.txt", 2245, "b92ba0f6ee02f2309628bfdadb123668a17c016e475ba477b857d33470d9d625", "decimal/LICENSE"},
	{"source/GO-BSD-NOTICE.txt", 1479, "2d36597f7117c38b006835ae7f537487207d8ec407aa9d9980794b2030cbc067", "decimal/GO-BSD-NOTICE"},
	{"observer.go.txt", 11282, "2019736abf02f2c99fc18680573c9870308fd053f0e007c47aec355cd0e46f46", "main.go"},
	{"INPUT.public.v1.json", 1469, "ea30987c3824e9db81990d758b33ca025a4cfc96bdc4c3d7ff28562057143008", "input.json"},
	{"HARNESS.go.mod.txt", 151, "a6dc47e93ddfb620f7ee58881540a02a677d89c2df57ad4851066149ca893da0", "go.mod"},
	{"evidence/OBSERVATIONS.actual.public.v1.json.gz", 2253, "a73966172f65417a697e07fb5e1d75dfc7410116fd325a889ceea8c8b1c7154f", ""},
	{"WANTS.public.v1.json", 3805, "c8b6f855ea811e3c00d648a0929bf03835dd9062c7c8c34924158228209fd64e", ""},
}

const inputPin, observationPin, wantPin = 8, 10, 11

type Fixture struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}
type Candidate struct {
	ID      string `json:"id"`
	Caption string `json:"caption"`
}
type Input struct {
	Schema     string      `json:"schema"`
	Fixtures   []Fixture   `json:"fixtures"`
	Candidates []Candidate `json:"candidates"`
}
type Call struct {
	Phase     string  `json:"phase"`
	Method    string  `json:"method"`
	Attempted bool    `json:"attempted"`
	Returned  bool    `json:"returned"`
	Error     bool    `json:"error"`
	Panic     bool    `json:"panic"`
	ErrorType *string `json:"error_type"`
	PanicType *string `json:"panic_type"`
	Boolean   *bool   `json:"boolean"`
	Integer   *string `json:"integer"`
	Text      *string `json:"text"`
}
type Snapshot struct {
	Coefficient *string `json:"coefficient"`
	Exponent    *int32  `json:"exponent"`
}
type CandidateResult struct {
	Code  string  `json:"code"`
	Value *string `json:"value"`
}
type Trial struct {
	FixtureID                 string           `json:"fixture_id"`
	CandidateID               string           `json:"candidate_id"`
	Status                    string           `json:"status"`
	ReturnedNormally          bool             `json:"returned_normally"`
	CandidateReturnedNormally bool             `json:"candidate_returned_normally"`
	UpstreamParseError        bool             `json:"upstream_parse_error"`
	UpstreamParseErrorType    *string          `json:"upstream_parse_error_type"`
	PanicPhase                *string          `json:"panic_phase"`
	PanicType                 *string          `json:"panic_type"`
	Result                    *CandidateResult `json:"result"`
	Before                    *Snapshot        `json:"before"`
	After                     *Snapshot        `json:"after"`
	CoefficientPreserved      *bool            `json:"coefficient_preserved"`
	ExponentPreserved         *bool            `json:"exponent_preserved"`
	Calls                     []Call           `json:"calls"`
}
type Counts struct {
	Fixtures               int `json:"fixtures"`
	Candidates             int `json:"candidates"`
	Trials                 int `json:"trials"`
	NormalTrialReturns     int `json:"normal_trial_returns"`
	NormalCandidateReturns int `json:"normal_candidate_returns"`
	UpstreamParseErrors    int `json:"upstream_parse_errors"`
	OwnedCandidateErrors   int `json:"owned_candidate_errors"`
	TrialPanics            int `json:"trial_panics"`
	APIAttempted           int `json:"api_attempted"`
	APIReturned            int `json:"api_returned"`
	APIErrors              int `json:"api_errors"`
	APIPanics              int `json:"api_panics"`
}
type Report struct {
	Schema                        string  `json:"schema"`
	Status                        string  `json:"status"`
	InputSHA256                   string  `json:"input_sha256"`
	Input                         Input   `json:"input"`
	Counts                        Counts  `json:"counts"`
	Trials                        []Trial `json:"trials"`
	LedgerScope                   string  `json:"ledger_scope"`
	PreMainAndHiddenCallsIncluded bool    `json:"pre_main_and_hidden_calls_included"`
	ModelCalls                    int     `json:"model_calls"`
	FitCalls                      int     `json:"fit_calls"`
	Qualified                     bool    `json:"qualified"`
	RoleAssigned                  bool    `json:"role_assigned"`
	DefaultActivated              bool    `json:"default_activated"`
	GPU                           bool    `json:"gpu"`
}
type Want struct {
	ParseSuccess         bool    `json:"parse_success"`
	Value                *string `json:"value_decimal_string"`
	OwnedError           *string `json:"owned_error"`
	UpstreamParseError   bool    `json:"upstream_parse_error"`
	CoefficientPreserved bool    `json:"coefficient_preserved"`
	ExponentPreserved    bool    `json:"exponent_preserved"`
	NormalReturn         bool    `json:"normal_return"`
	Panic                bool    `json:"panic"`
}
type Wants struct {
	Schema             string          `json:"schema"`
	State              string          `json:"state"`
	Request            string          `json:"request"`
	Parents            int             `json:"parents"`
	Role               json.RawMessage `json:"role"`
	Qualified          bool            `json:"qualified"`
	Labels             int             `json:"learning_labels"`
	Fit                int             `json:"Fit"`
	IndependentParents int             `json:"new_independent_golden_parents"`
	OwnedCodes         []string        `json:"owned_error_codes"`
	Fixtures           []struct {
		ID    string `json:"id"`
		Input string `json:"input"`
		Want  Want   `json:"Want"`
	} `json:"fixtures"`
	Scope string `json:"scope"`
}
type packet struct {
	Files [len(pins)][]byte
	Input Input
	Wants Wants
	Saved Report
}

func sha(b []byte) string { x := sha256.Sum256(b); return hex.EncodeToString(x[:]) }
func read(path string, limit int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("packet unavailable")
	}
	b, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	closeErr := f.Close()
	if err != nil || closeErr != nil || len(b) > limit {
		return nil, errors.New("packet read/bound failure")
	}
	return b, nil
}

// Standard JSON shape/EOF validation; not generic canonical JSON validation.
func decode(b []byte, v any) error {
	if !utf8.Valid(b) {
		return errors.New("invalid UTF8")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return errors.New("invalid JSON shape")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errors.New("JSON trailing value")
	}
	return nil
}
func unpack(b []byte) ([]byte, error) {
	r := bytes.NewReader(b)
	z, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.New("gzip header failure")
	}
	z.Multistream(false)
	raw, err := io.ReadAll(io.LimitReader(z, 262145))
	closeErr := z.Close()
	if err != nil || closeErr != nil || len(raw) > 262144 || r.Len() != 0 {
		return nil, errors.New("gzip CRC/EOF/bound failure")
	}
	return raw, nil
}
func loadPacket(root string) (p packet, err error) {
	for i, pin := range pins {
		p.Files[i], err = read(filepath.Join(root, pin.name), pin.size)
		if err != nil {
			return p, err
		}
		if len(p.Files[i]) != pin.size || sha(p.Files[i]) != pin.sha {
			return p, errors.New("packet pin mismatch")
		}
	}
	if decode(p.Files[inputPin], &p.Input) != nil || decode(p.Files[wantPin], &p.Wants) != nil {
		return p, errors.New("frozen JSON invalid")
	}
	in, w := p.Input, p.Wants
	if in.Schema != "riido-decimal-frozen-v1" || len(in.Fixtures) != 8 || len(in.Candidates) != 4 || w.Schema != "riido-decimal-literal-wants-v1" || len(w.Fixtures) != 8 || w.Parents != 1 || w.Qualified || w.Labels != 0 || w.Fit != 0 || w.IndependentParents != 0 || string(w.Role) != "null" || !reflect.DeepEqual(w.OwnedCodes, []string{"non_integer", "out_of_range"}) {
		return p, errors.New("frozen authority mismatch")
	}
	for i, f := range in.Fixtures {
		if f.ID != w.Fixtures[i].ID || f.Value != w.Fixtures[i].Input {
			return p, errors.New("fixture identity mismatch")
		}
	}
	ids := [4]string{"strict_fraction_first", "strict_range_first", "truncate_checked", "intpart_unchecked"}
	for i, c := range in.Candidates {
		if c.ID != ids[i] {
			return p, errors.New("candidate identity mismatch")
		}
	}
	raw, err := unpack(p.Files[observationPin])
	if err != nil {
		return p, err
	}
	if len(raw) != 74491 || sha(raw) != "3a45b8bcd54b44f6ae352781455d18d9861eae9cf63a3f1c5536288be6f8b49c" {
		return p, errors.New("saved raw pin mismatch")
	}
	if decode(raw, &p.Saved) != nil || validate(p.Saved, in, w) != nil {
		return p, errors.New("saved report invalid")
	}
	return p, nil
}
func trueValue(v *bool) bool   { return v != nil && *v }
func present(s *Snapshot) bool { return s != nil && s.Coefficient != nil && s.Exponent != nil }
func callShape(e Call) bool {
	if !e.Attempted || !e.Returned || e.Error || e.Panic || e.ErrorType != nil || e.PanicType != nil {
		return false
	}
	switch e.Method {
	case "decimal.NewFromString", "decimal.Decimal.Coefficient", "decimal.Decimal.BigInt":
		return e.Boolean == nil && e.Integer == nil && e.Text == nil
	case "big.Int.String":
		return e.Text != nil && e.Boolean == nil && e.Integer == nil
	case "decimal.Decimal.IsInteger", "big.Int.IsInt64":
		return e.Boolean != nil && e.Integer == nil && e.Text == nil
	case "decimal.Decimal.Exponent", "decimal.Decimal.IntPart", "big.Int.Int64":
		if e.Integer == nil || e.Boolean != nil || e.Text != nil {
			return false
		}
		bits := 64
		if e.Method == "decimal.Decimal.Exponent" {
			bits = 32
		}
		v, err := strconv.ParseInt(*e.Integer, 10, bits)
		return err == nil && strconv.FormatInt(v, 10) == *e.Integer
	}
	return false
}
func snapshotCalls(c []Call, phase string, s *Snapshot) bool {
	return len(c) == 3 && c[0].Phase == phase && c[0].Method == "decimal.Decimal.Coefficient" && c[1].Phase == phase && c[1].Method == "big.Int.String" && c[2].Phase == phase && c[2].Method == "decimal.Decimal.Exponent" && *c[1].Text == *s.Coefficient && *c[2].Integer == strconv.FormatInt(int64(*s.Exponent), 10)
}

// Check recorded branch consumption, not a second execution of the upstream API.
func candidateCalls(t Trial, c []Call) bool {
	pos := 0
	take := func(method string) (Call, bool) {
		if pos == len(c) || c[pos].Phase != "candidate" || c[pos].Method != method {
			return Call{}, false
		}
		e := c[pos]
		pos++
		return e, true
	}
	owned := func(code string) bool { return pos == len(c) && t.Result.Code == code && t.Result.Value == nil }
	finish := func(method string) bool {
		e, ok := take(method)
		return ok && pos == len(c) && t.Result.Code == "success" && reflect.DeepEqual(e.Integer, t.Result.Value)
	}
	if t.CandidateID == "intpart_unchecked" {
		return finish("decimal.Decimal.IntPart")
	}
	if t.CandidateID == "strict_fraction_first" {
		e, ok := take("decimal.Decimal.IsInteger")
		if !ok {
			return false
		}
		if !*e.Boolean {
			return owned("non_integer")
		}
	}
	if _, ok := take("decimal.Decimal.BigInt"); !ok {
		return false
	}
	e, ok := take("big.Int.IsInt64")
	if !ok {
		return false
	}
	if !*e.Boolean {
		return owned("out_of_range")
	}
	if t.CandidateID == "strict_range_first" {
		e, ok = take("decimal.Decimal.IsInteger")
		if !ok {
			return false
		}
		if !*e.Boolean {
			return owned("non_integer")
		}
	}
	return finish("big.Int.Int64")
}
func validate(r Report, in Input, w Wants) error {
	if r.Schema != "riido-decimal-actual-development-audit-v1" || r.Status != "observed_development_audit" || r.InputSHA256 != pins[inputPin].sha || !reflect.DeepEqual(r.Input, in) || r.LedgerScope != "explicit_wrapped_decimal_and_bigint_calls_only" || r.PreMainAndHiddenCallsIncluded || r.ModelCalls != 0 || r.FitCalls != 0 || r.Qualified || r.RoleAssigned || r.DefaultActivated || r.GPU || len(r.Trials) != 32 {
		return errors.New("report authority mismatch")
	}
	counts := Counts{Fixtures: 8, Candidates: 4}
	var passed [4]int
	for k, t := range r.Trials {
		i, j := k/4, k%4
		f, want := in.Fixtures[i], w.Fixtures[i].Want
		if t.FixtureID != f.ID || t.CandidateID != in.Candidates[j].ID || t.Status != "observed" || !t.ReturnedNormally || !t.CandidateReturnedNormally || t.UpstreamParseError || t.UpstreamParseErrorType != nil || t.PanicPhase != nil || t.PanicType != nil || t.Result == nil || !present(t.Before) || !present(t.After) || !trueValue(t.CoefficientPreserved) || !trueValue(t.ExponentPreserved) || len(t.Calls) < 8 {
			return errors.New("trial shape/return mismatch")
		}
		// These fixed decimal literals have no exponents/leading zero ambiguity.
		coefficient, exponent := strings.Replace(f.Value, ".", "", 1), int32(0)
		if dot := strings.IndexByte(f.Value, '.'); dot >= 0 {
			exponent = -int32(len(f.Value) - dot - 1)
		}
		if *t.Before.Coefficient != coefficient || *t.Before.Exponent != exponent || !reflect.DeepEqual(t.Before, t.After) {
			return errors.New("input representation preservation mismatch")
		}
		for _, e := range t.Calls {
			if !callShape(e) {
				return errors.New("ledger result shape mismatch")
			}
			if e.Attempted {
				counts.APIAttempted++
			}
			if e.Returned {
				counts.APIReturned++
			}
			if e.Error {
				counts.APIErrors++
			}
			if e.Panic {
				counts.APIPanics++
			}
		}
		end := len(t.Calls) - 3
		if t.Calls[0].Phase != "setup" || t.Calls[0].Method != "decimal.NewFromString" || !snapshotCalls(t.Calls[1:4], "before", t.Before) || !snapshotCalls(t.Calls[end:], "after", t.After) || !candidateCalls(t, t.Calls[4:end]) {
			return errors.New("ledger phase/branch mismatch")
		}
		counts.Trials++
		if t.ReturnedNormally {
			counts.NormalTrialReturns++
		}
		if t.CandidateReturnedNormally {
			counts.NormalCandidateReturns++
		}
		if t.UpstreamParseError {
			counts.UpstreamParseErrors++
		}
		if t.PanicType != nil {
			counts.TrialPanics++
		}
		if t.Result.Code != "success" {
			counts.OwnedCandidateErrors++
		}
		code := "success"
		if want.OwnedError != nil {
			code = *want.OwnedError
		}
		if want.ParseSuccess == !t.Calls[0].Error && reflect.DeepEqual(t.Result.Value, want.Value) && t.Result.Code == code && t.UpstreamParseError == want.UpstreamParseError && *t.CoefficientPreserved == want.CoefficientPreserved && *t.ExponentPreserved == want.ExponentPreserved && t.ReturnedNormally == want.NormalReturn && (t.PanicType != nil) == want.Panic {
			passed[j]++
		}
	}
	expected := Counts{Fixtures: 8, Candidates: 4, Trials: 32, NormalTrialReturns: 32, NormalCandidateReturns: 32, OwnedCandidateErrors: 13, APIAttempted: 298, APIReturned: 298}
	if counts != expected || counts != r.Counts || passed != [4]int{8, 7, 5, 3} {
		return errors.New("counts/literal-Want matrix mismatch")
	}
	return nil
}

// Named buffer deliberately does not expose bytes.Buffer.ReadFrom to io.Copy.
type bounded struct {
	buf bytes.Buffer
	max int
}

func (b *bounded) Write(p []byte) (int, error) {
	if len(p) > b.max-b.buf.Len() {
		return 0, errors.New("command output bound")
	}
	return b.buf.Write(p)
}
func (b *bounded) Bytes() []byte { return b.buf.Bytes() }
func environment(tmp string) []string {
	values := []string{"CGO_ENABLED=0", "GOPROXY=off", "GOSUMDB=off", "GOTOOLCHAIN=local", "GOMAXPROCS=2", "GOMEMLIMIT=96MiB", "GOWORK=off", "GOENV=off", "GOFLAGS=", "TMPDIR=" + filepath.Join(tmp, "build-temp")}
	var out []string
	for _, e := range os.Environ() {
		keep := true
		for _, v := range values {
			key := v[:strings.IndexByte(v, '=')+1]
			if strings.HasPrefix(e, key) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, e)
		}
	}
	return append(out, values...)
}
func command(exe string, args []string, dir string, seconds, outMax int) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, exe, args...)
	c.Dir, c.Env, c.WaitDelay = dir, environment(dir), time.Second
	out, stderr := &bounded{max: outMax}, &bounded{max: 8192}
	c.Stdout, c.Stderr = out, stderr
	if c.Run() != nil || ctx.Err() != nil {
		return nil, errors.New("command failed/timeout/output bound")
	}
	return out.Bytes(), nil
}
func run() (err error) {
	flags := flag.NewFlagSet("decimal-replay", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "experiments/short-claim/next-cohort-decimal-audit", "public packet root")
	goExe := flags.String("go", "go", "local Go executable")
	if flags.Parse(os.Args[1:]) != nil || flags.NArg() != 0 {
		return errors.New("invalid argument")
	}
	tool, err := exec.LookPath(*goExe)
	if err != nil {
		return errors.New("Go unavailable")
	}
	tool, err = filepath.Abs(tool)
	if err != nil {
		return errors.New("Go locator invalid")
	}
	p, err := loadPacket(*root)
	if err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "riido-decimal-replay-")
	if err != nil {
		return errors.New("temporary directory failed")
	}
	complete := false
	defer func() {
		if os.RemoveAll(tmp) != nil && err == nil {
			err = errors.New("temporary cleanup failed")
		}
		if complete && err == nil {
			if _, e := fmt.Fprintln(os.Stdout, "Decimal replay passed: parent=1 trials=32 explicit_calls=298 owned_errors=13 passes=8/7/5/3 Fit=0 model=0"); e != nil {
				err = errors.New("summary output failed")
			}
		}
	}()
	if os.Mkdir(filepath.Join(tmp, "decimal"), 0700) != nil {
		return errors.New("temporary Decimal directory failed")
	}
	if os.Mkdir(filepath.Join(tmp, "build-temp"), 0700) != nil {
		return errors.New("temporary build directory failed")
	}
	for i, pin := range pins {
		if pin.dest != "" && os.WriteFile(filepath.Join(tmp, pin.dest), p.Files[i], 0600) != nil {
			return errors.New("temporary source copy failed")
		}
	}
	version, err := command(tool, []string{"version"}, tmp, 5, 8192)
	if err != nil {
		return errors.New("version stage failed/timeout/output bound")
	}
	fields := strings.Fields(string(version))
	if len(fields) != 4 || fields[0] != "go" || fields[1] != "version" || fields[2] != "go1.27.1" {
		return errors.New("Go 1.27.1 required")
	}
	worker := filepath.Join(tmp, "worker")
	if _, err = command(tool, []string{"build", "-p=1", "-trimpath", "-buildvcs=false", "-o", worker, "."}, tmp, 60, 65536); err != nil {
		return errors.New("build stage failed/timeout/output bound")
	}
	output, err := command(worker, []string{filepath.Join(tmp, "input.json")}, tmp, 20, 65536)
	if err != nil {
		return errors.New("worker stage failed/timeout/output bound")
	}
	raw, err := unpack(output)
	if err != nil {
		return err
	}
	var actual Report
	if decode(raw, &actual) != nil || validate(actual, p.Input, p.Wants) != nil || !reflect.DeepEqual(p.Saved, actual) {
		return errors.New("complete typed replay mismatch")
	}
	complete = true
	return nil
}
func guardedRun() (err error) {
	complete := false
	defer func() {
		if !complete {
			_ = recover()
			err = errors.New("replay panic")
		}
	}()
	err = run()
	complete = true
	return
}
func main() {
	if err := guardedRun(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
