// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Read-only, unlearned literal/event evidence shadow. No semantic model or app client.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/statehintfamily"
)

const (
	inputBudget  = 1 << 20
	maxEvents    = 100
	maxTextBytes = 4096
	anchor       = ".cache/statehint-assertion-shadow"
)

var errShadow = errors.New("assertion shadow input, checksum or private output invalid")

type observation struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // text_observation or observed_event
	Kind      string `json:"kind,omitempty"`
	Role      string `json:"role,omitempty"` // body_snapshot or comment
	Text      string `json:"text,omitempty"`
	Actor     string `json:"actor,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Version   string `json:"version,omitempty"`
	Source    string `json:"source"`
}
type input struct {
	Schema       string        `json:"schema"`
	Observations []observation `json:"observations"`
}
type provenance struct {
	Actor     string `json:"actor,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Version   string `json:"version,omitempty"`
	Source    string `json:"source"`
}
type evidence struct {
	ObservationID      string     `json:"observation_id"`
	Kind               string     `json:"candidate_kind"`
	Role               string     `json:"text_role"`
	Start              int        `json:"start_byte"`
	End                int        `json:"end_byte_exclusive"`
	Literal            string     `json:"literal"`
	Unlearned          bool       `json:"unlearned_literal_evidence"`
	SemanticJudgment   bool       `json:"semantic_judgment"`
	Confidence         *float64   `json:"confidence"`
	DisplayCandidate   bool       `json:"experimental_display_candidate"`
	CompletionWithheld bool       `json:"completion_display_withheld"`
	Reasons            []string   `json:"uncertainty_reasons"`
	Provenance         provenance `json:"provenance"`
}
type eventFact struct {
	ObservationID        string     `json:"observation_id"`
	Kind                 string     `json:"supplied_observed_event_kind"`
	InputAttestationOnly bool       `json:"input_attestation_only"`
	CompletionAuthority  bool       `json:"task_completion_authority"`
	Provenance           provenance `json:"provenance"`
}
type report struct {
	Schema             string      `json:"schema"`
	Status             string      `json:"status"`
	InputSHA           string      `json:"input_sha256"`
	Observations       int         `json:"observations"`
	CommentTexts       int         `json:"comment_text_observations"`
	BodySnapshots      int         `json:"body_snapshot_observations"`
	InputGap           []string    `json:"input_gaps"`
	Evidence           []evidence  `json:"literal_candidates"`
	Facts              []eventFact `json:"supplied_event_facts"`
	CompletionWithheld bool        `json:"completion_display_withheld"`
	ModelCalls         int         `json:"model_calls"`
	FitCalls           int         `json:"fit_calls"`
	StateWrites        int         `json:"annotation_reaction_status_writes"`
	Qualified          bool        `json:"semantic_quality_qualified"`
	ExistingGates      [2]float64  `json:"unchanged_model_confidence_margin"`
	GateApplied        bool        `json:"probabilistic_model_gate_applied"`
	ElapsedNS          int64       `json:"elapsed_nanoseconds"`
	GoHeap             uint64      `json:"go_heap_alloc_bytes_not_os_rss"`
	MemoryNote         string      `json:"memory_measurement_note"`
}

func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func local(name string) bool {
	return filepath.IsLocal(name) && filepath.Clean(name) == name && name != "." && !strings.Contains(name, "\\")
}
func shaValid(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && hex.EncodeToString(b) == s
}
func bounded(s string, n int) bool { return len(s) <= n && utf8.ValidString(s) }
func knownFact(s string) bool {
	switch s {
	case "CommentObserved", "CommentCreated", "ProgressObserved", "CommandSucceeded", "TurnEnded":
		return true
	}
	return false
}
func decode(b []byte) (input, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var in input
	if d.Decode(&in) != nil {
		return in, errShadow
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return in, errShadow
	}
	if in.Schema != "riido-assertion-shadow-input-v1" || len(in.Observations) < 1 || len(in.Observations) > maxEvents {
		return in, errShadow
	}
	for i, o := range in.Observations {
		if o.ID == "" || o.Source == "" || !bounded(o.ID, 128) || !bounded(o.Source, 256) || !bounded(o.Actor, 128) || !bounded(o.Unit, 256) || !bounded(o.Timestamp, 64) || !bounded(o.Version, 128) || !bounded(o.Text, maxTextBytes) {
			return in, errShadow
		}
		if o.Timestamp != "" {
			if _, e := time.Parse(time.RFC3339Nano, o.Timestamp); e != nil {
				return in, errShadow
			}
		}
		for _, prior := range in.Observations[:i] {
			if prior.ID == o.ID {
				return in, errShadow
			}
		}
		switch o.Type {
		case "text_observation":
			if o.Kind != "" || (o.Role != "body_snapshot" && o.Role != "comment") || strings.TrimSpace(o.Text) == "" {
				return in, errShadow
			}
		case "observed_event":
			if !knownFact(o.Kind) || o.Role != "" || o.Text != "" {
				return in, errShadow
			}
		default:
			return in, errShadow
		}
	}
	return in, nil
}

type region struct {
	start, end int
	reason     string
}

func appendRegion(v []region, start, end int, reason string) []region {
	if end > start {
		return append(v, region{start, end, reason})
	}
	return v
}
func visible(text string) (string, []region) {
	b := []byte(text)
	var regions []region
	lowerBytes := []byte(text)
	for i, v := range lowerBytes {
		if v >= 'A' && v <= 'Z' {
			lowerBytes[i] += 'a' - 'A'
		}
	}
	lower := string(lowerBytes)
	// Tag spelling and attributes are never literal message text. Preserve byte offsets.
	for i := 0; i < len(text); i++ {
		if text[i] != '<' {
			continue
		}
		j := strings.IndexByte(text[i:], '>')
		if j < 0 {
			regions = appendRegion(regions, i, len(text), "unsupported_markup")
			break
		}
		j += i
		for k := i; k <= j; k++ {
			if b[k] != '\n' {
				b[k] = ' '
			}
		}
		tag := strings.TrimSpace(lower[i+1 : j])
		nameEnd := strings.IndexAny(tag, " \t\r\n\f")
		if nameEnd >= 0 {
			tag = tag[:nameEnd]
		}
		for _, name := range []string{"pre", "code", "blockquote", "script", "style"} {
			if tag == name {
				closing := "</" + name + ">"
				end := strings.Index(lower[j+1:], closing)
				regionEnd := len(text)
				if end >= 0 {
					regionEnd = j + 1 + end
				}
				reason := "quoted_or_code"
				if name == "script" || name == "style" {
					reason = "unsupported_markup"
				}
				regions = appendRegion(regions, j+1, regionEnd, reason)
			}
		}
		i = j
	}
	for _, delimiter := range []string{"```", "`", "\"", "“"} {
		close := delimiter
		if delimiter == "“" {
			close = "”"
		}
		cursor := 0
		for cursor < len(text) {
			i := strings.Index(text[cursor:], delimiter)
			if i < 0 {
				break
			}
			i += cursor
			j := strings.Index(text[i+len(delimiter):], close)
			if j < 0 {
				regions = appendRegion(regions, i, len(text), "quoted_or_code")
				break
			}
			j += i + len(delimiter)
			regions = appendRegion(regions, i, j+len(close), "quoted_or_code")
			cursor = j + len(close)
		}
	}
	offset := 0
	for _, line := range strings.SplitAfter(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			regions = appendRegion(regions, offset, offset+len(line), "quoted_or_code")
		}
		offset += len(line)
	}
	return string(b), regions
}
func reasonAdd(v []string, s string) []string {
	for _, x := range v {
		if x == s {
			return v
		}
	}
	return append(v, s)
}
func sentence(text string, start, end int) string {
	a, b := start, end
	for a > 0 && !strings.ContainsRune(".!?\n", rune(text[a-1])) {
		a--
	}
	for b < len(text) && !strings.ContainsRune(".!?\n", rune(text[b])) {
		b++
	}
	return text[a:b]
}
func provenanceOf(o observation) provenance {
	return provenance{o.Actor, o.Unit, o.Timestamp, o.Version, o.Source}
}
func contextReasons(o observation) []string {
	v := []string{}
	if o.Role == "body_snapshot" {
		v = append(v, "body_snapshot_is_not_a_comment")
	}
	for _, x := range []struct{ value, reason string }{{o.Actor, "actor_unknown"}, {o.Unit, "unit_unknown"}, {o.Timestamp, "time_unknown"}, {o.Version, "version_unknown"}} {
		if x.value == "" {
			v = append(v, x.reason)
		}
	}
	return v
}
func extract(o observation) []evidence {
	text, regions := visible(o.Text)
	specs := []struct{ kind, pattern string }{
		{"reply_claim", `(?i)\b(?:please\s+(?:explain|clarify|confirm|tell|choose|advise)|(?:can|could|would)\s+(?:you|someone|anyone)\s+(?:please\s+)?(?:explain|clarify|confirm|review|answer|tell|choose|advise))\b|(?:설명해|알려|확인해|선택해)\s*주세요`},
		{"activity_claim", `(?i)\b(?:working|implementing|testing|checking|fixing|building)\b|(?:작업|구현|진행|확인)\s*중`},
		{"closure_claim", `(?i)\b(?:finished|completed|closed|wrapped\s+up)\b|(?:완료했습니다|마쳤습니다|끝냈습니다|완료했어요)`},
	}
	neg := regexp.MustCompile(`(?i)\b(?:not|never|didn't|haven't|hasn't|isn't|wasn't)\b|(?:않았|못했|아니|안\s*했)`)
	future := regexp.MustCompile(`(?i)\b(?:will|plan|planned|tomorrow|intend|going\s+to|if|when|might)\b|(?:예정|계획|내일|하면)`)
	remainder := regexp.MustCompile(`(?i)\b(?:still|remaining|yet|required|must|need)\b|(?:아직|남아|해야|필수)`)
	var result []evidence
	for _, spec := range specs {
		for _, span := range regexp.MustCompile(spec.pattern).FindAllStringIndex(text, -1) {
			reasons := contextReasons(o)
			for _, r := range regions {
				if span[0] < r.end && span[1] > r.start {
					reasons = reasonAdd(reasons, r.reason)
				}
			}
			clause := sentence(text, span[0], span[1])
			if neg.MatchString(clause) {
				reasons = reasonAdd(reasons, "negated_context")
			}
			if future.MatchString(clause) {
				reasons = reasonAdd(reasons, "future_or_conditional_context")
			}
			if remainder.MatchString(text) && spec.kind == "closure_claim" {
				reasons = reasonAdd(reasons, "required_remainder_scope_unresolved")
			}
			// Activity word presence alone has no reliably resolved actor/current-unit binding.
			if spec.kind == "activity_claim" {
				direct := regexp.MustCompile(`(?i)\b(?:I\s+am|I'm|we\s+are|we're)\s+(?:currently\s+)?(?:working|implementing|testing|checking|fixing|building)\b|(?:지금|현재).*(?:작업|구현|진행|확인)\s*중`)
				if !direct.MatchString(clause) {
					reasons = reasonAdd(reasons, "literal_activity_scope_unresolved")
				}
				if o.Unit != "" && !strings.Contains(strings.ToLower(clause), strings.ToLower(o.Unit)) {
					reasons = reasonAdd(reasons, "unit_scope_unresolved")
				}
			}
			withheld := spec.kind == "closure_claim"
			if withheld {
				reasons = reasonAdd(reasons, "completion_display_always_withheld")
			}
			result = append(result, evidence{o.ID, spec.kind, o.Role, span[0], span[1], o.Text[span[0]:span[1]], true, false, nil, len(reasons) == 0 && !withheld, withheld, reasons, provenanceOf(o)})
		}
	}
	return result
}
func shadow(in input, pin string) report {
	start := time.Now()
	r := report{Schema: "riido-assertion-shadow-report-v1", Status: "unlearned read-only research shadow", InputSHA: pin, Observations: len(in.Observations), InputGap: []string{}, Evidence: []evidence{}, Facts: []eventFact{}, CompletionWithheld: true, ExistingGates: [2]float64{statehintfamily.ConfidenceFloor, statehintfamily.MarginFloor}, MemoryNote: "Go heap is not peak OS RSS. Measure whole-command peak RSS with an external process wrapper; no model/native/GPU runtime."}
	for _, o := range in.Observations {
		if o.Type == "observed_event" {
			r.Facts = append(r.Facts, eventFact{o.ID, o.Kind, true, false, provenanceOf(o)})
			continue
		}
		if o.Role == "comment" {
			r.CommentTexts++
		} else {
			r.BodySnapshots++
		}
		r.Evidence = append(r.Evidence, extract(o)...)
	}
	if r.CommentTexts == 0 {
		r.InputGap = append(r.InputGap, "no_comment_text_observations; body snapshots cannot establish live work-state truth")
	}
	r.ElapsedNS = time.Since(start).Nanoseconds()
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	r.GoHeap = memory.HeapAlloc
	return r
}
func readInput(root *os.Root, name, pin string) ([]byte, error) {
	if !local(name) || !shaValid(pin) {
		return nil, errShadow
	}
	before, e := root.Lstat(name)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > inputBudget {
		return nil, errShadow
	}
	f, e := root.Open(name)
	if e != nil {
		return nil, errShadow
	}
	defer f.Close()
	after, e := f.Stat()
	if e != nil || !os.SameFile(before, after) {
		return nil, errShadow
	}
	b, e := io.ReadAll(io.LimitReader(f, inputBudget+1))
	if e != nil || len(b) > inputBudget || digest(b) != pin {
		return nil, errShadow
	}
	return b, nil
}
func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errShadow
	}
	for _, p := range []string{".cache", anchor} {
		if e := root.Mkdir(p, 0700); e != nil && !errors.Is(e, os.ErrExist) {
			return nil, errShadow
		}
		s, e := root.Lstat(p)
		if e != nil || !s.IsDir() || s.Mode()&os.ModeSymlink != 0 || p == anchor && s.Mode().Perm() != 0700 {
			return nil, errShadow
		}
	}
	if root.Mkdir(name, 0700) != nil {
		return nil, errShadow
	}
	return root.OpenRoot(name)
}
func run(args []string, out, errOut io.Writer) error {
	flags := flag.NewFlagSet("riido-statehint-assertion-shadow", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	name := flags.String("input", "", "explicit local bounded observation JSON")
	pin := flags.String("input-sha256", "", "exact input byte SHA")
	dest := flags.String("out", "", "new private report directory")
	check := flags.Bool("check", false, "validate only, no output")
	if e := flags.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errOut, "riido-statehint-assertion-shadow --input FILE --input-sha256 SHA --check; or --out .cache/statehint-assertion-shadow/NEW-RUN. No models, DB/app client, status writes or confidence/quality claim.")
			return e
		}
		return errShadow
	}
	if flags.NArg() != 0 || *check && *dest != "" || !*check && (!local(*dest) || filepath.Dir(*dest) != anchor) {
		return errShadow
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errShadow
	}
	defer root.Close()
	b, e := readInput(root, *name, *pin)
	if e != nil {
		return e
	}
	in, e := decode(b)
	if e != nil {
		return e
	}
	if *check {
		return json.NewEncoder(out).Encode(struct {
			Status       string `json:"status"`
			Observations int    `json:"observations"`
		}{"checked; no model/app calls or output", len(in.Observations)})
	}
	result := shadow(in, *pin)
	private, e := privateOutput(root, *dest)
	if e != nil {
		return e
	}
	defer private.Close()
	data, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		return errShadow
	}
	data = append(data, '\n')
	file, e := private.OpenFile("report.json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errShadow
	}
	n, e := file.Write(data)
	if e == nil && n != len(data) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = file.Sync()
	}
	closeErr := file.Close()
	if e != nil || closeErr != nil {
		return errShadow
	}
	return json.NewEncoder(out).Encode(struct {
		Status                          string `json:"status"`
		Candidates, Facts, CommentTexts int
		CompletionWithheld              bool `json:"completion_display_withheld"`
	}{result.Status, len(result.Evidence), len(result.Facts), result.CommentTexts, true})
}
func main() {
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "assertion shadow failed; check bounded pinned input and fresh private output")
		os.Exit(1)
	}
}
