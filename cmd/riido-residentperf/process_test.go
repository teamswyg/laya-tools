package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const lifecycleChildSource = "package main\nimport (\n \"bufio\"\n \"bytes\"\n \"crypto/sha256\"\n \"encoding/binary\"\n \"encoding/hex\"\n \"encoding/json\"\n \"io\"\n \"os\"\n \"strings\"\n \"time\"\n)\ntype candidate struct { ID string \"json:\\\"id\\\"\"; Text string \"json:\\\"text\\\"\" }\ntype input struct { Schema string \"json:\\\"schema\\\"\"; Request string \"json:\\\"request\\\"\"; Candidates []candidate \"json:\\\"candidates\\\"\"; Provenance string \"json:\\\"provenance\\\"\" }\ntype scored struct { ID string \"json:\\\"id\\\"\"; Score float64 \"json:\\\"score\\\"\" }\ntype output struct { Schema string \"json:\\\"schema\\\"\"; Status string \"json:\\\"status\\\"\"; Baseline string \"json:\\\"baseline\\\"\"; InputSHA256 string \"json:\\\"input_sha256\\\"\"; FallbackReason string \"json:\\\"fallback_reason,omitempty\\\"\"; Candidates []scored \"json:\\\"verification_order\\\"\" }\nfunc digest(in input) string {\n h:=sha256.New()\n write:=func(s string){var b [8]byte;binary.BigEndian.PutUint64(b[:],uint64(len(s)));h.Write(b[:]);io.WriteString(h,s)}\n write(in.Schema);write(in.Request);write(in.Provenance)\n for _,c:=range in.Candidates{write(c.ID);write(c.Text)}\n return hex.EncodeToString(h.Sum(nil))\n}\nfunc stall(){for {time.Sleep(time.Second)}}\nfunc main(){\n if len(os.Args)!=4||os.Args[1]!=\"--stream\"||os.Args[2]!=\"--baseline\" {os.Exit(9)}\n scanner:=bufio.NewScanner(os.Stdin);scanner.Buffer(make([]byte,16384),16384)\n count:=0;mode:=\"\";var last []byte\n for scanner.Scan(){\n  count++\n  var in input\n  if json.Unmarshal(scanner.Bytes(),&in)!=nil{os.Exit(10)}\n  mode=strings.Fields(in.Request)[1]\n  if mode==\"stall-first\"&&count==1||mode==\"stall-warm\"&&count==2||mode==\"stall-timed\"&&count==22 {stall()}\n  if mode==\"early-exit\" {os.Exit(7)}\n  if mode==\"truncated\" {io.WriteString(os.Stdout,\"{\\\"schema\\\":\");return}\n  if mode==\"oversized\" {io.WriteString(os.Stdout,strings.Repeat(\"x\",12289));stall()}\n  if mode==\"bad-json\" {io.WriteString(os.Stdout,\"{\\\"authored_private_marker\\\":\\\"x\\\"}\\n\");return}\n  if mode==\"stderr-overflow\" {io.WriteString(os.Stderr,strings.Repeat(\"authored-private-marker\",300));stall()}\n  if mode==\"stderr\"&&count==1 {io.WriteString(os.Stderr,\"authored-private-marker\")}\n  if os.Getenv(\"GOMAXPROCS\")!=\"1\"||os.Getenv(\"GOMEMLIMIT\")!=\"256MiB\"||os.Getenv(\"GODEBUG\")!=\"\"||os.Getenv(\"GOTRACEBACK\")!=\"none\" {os.Exit(11)}\n  out:=output{Schema:\"riido-shortclaim-order-v1\",Status:\"unverified_heuristic\",Baseline:os.Args[3],InputSHA256:digest(in),Candidates:[]scored{{ID:\"candidate-0\"},{ID:\"candidate-1\"},{ID:\"candidate-2\"}}}\n  var buffer bytes.Buffer\n  json.NewEncoder(&buffer).Encode(out)\n  last=append(last[:0],buffer.Bytes()...)\n  if mode==\"split\" {\n   os.Stdout.Write(last[:3]);os.Stdout.Write(last[3:len(last)-1]);os.Stdout.Write(last[len(last)-1:])\n  } else {os.Stdout.Write(last)}\n }\n if mode==\"terminal-extra\" {os.Stdout.Write(last)}\n if mode==\"terminal-nonzero\" {os.Exit(4)}\n if mode==\"terminal-hang\" {os.Stdout.Close();os.Stderr.Close();stall()}\n}\n"

// Only this independently authored synthetic protocol helper is built/executed.
// Never invokes the official corpus/replay, candidate truth, or an ML model.
func buildLifecycleChild(t *testing.T) string {
	t.Helper()
	if runtime.Version() != goVersion {
		t.Fatal("test_toolchain_not_pinned")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "synthetic.go")
	binary := filepath.Join(dir, "synthetic-child")
	if os.WriteFile(source, []byte(lifecycleChildSource), 0600) != nil {
		t.Fatal("helper_source_create_failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-trimpath", "-buildvcs=false", "-o", binary, source)
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "GOWORK=off", "GOENV=off", "GOFLAGS=-p=1", "CGO_ENABLED=0"}
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	if cmd.Run() != nil {
		t.Fatal("helper_build_failed")
	}
	return binary
}
func lifecyclePlan() Plan {
	var plan Plan
	plan.ChildTimeoutSeconds, plan.CleanupGraceMilliseconds = 15, 1000
	plan.Rows[0] = RowPlan{Baseline: "fixed_order", Workload: "distinct", Repeat: 1}
	return plan
}
func lifecycleInputs(t *testing.T, mode string) ReadyCorpus {
	t.Helper()
	var corpus Corpus
	for _, suffix := range []string{"alpha", "beta", "gamma"} {
		input := shortclaim.Input{Schema: textSchema, Request: "control " + mode + " " + suffix, Provenance: provenance, Candidates: []shortclaim.Candidate{{ID: "candidate-0", Text: "Keep input order."}, {ID: "candidate-1", Text: "Reverse input order."}, {ID: "candidate-2", Text: "Preserve all input."}}}
		p, e := shortclaim.Validate(input)
		if e != nil {
			t.Fatal("synthetic_input_invalid")
		}
		wire, e := json.Marshal(input)
		if e != nil {
			t.Fatal("synthetic_wire_invalid")
		}
		payload := Payload{Wire: wire, InputDigest: inputDigest(p)}
		for i, kind := range baselineKinds() {
			_, payload.Expected[i], e = expectedResponse(p, kind)
			if e != nil {
				t.Fatal("synthetic_expectation_invalid")
			}
		}
		corpus.Payloads = append(corpus.Payloads, payload)
	}
	ready, e := prepareWireTable(corpus)
	if e != nil {
		t.Fatal("synthetic_prepare_failed")
	}
	return ready
}
func assertLifecycleCleanup(t *testing.T, r RowResult) {
	t.Helper()
	if !r.Started || !r.Reaped || r.WaitCalls != 1 || !r.StdoutJoined || !r.StderrJoined || !r.WatcherJoined || !r.StdoutReceiptAvailable || !r.StderrReceiptAvailable || !r.CleanupComplete || r.CleanupError != "" {
		t.Fatalf("synthetic child or goroutine cleanup not confirmed: %+v", r)
	}
	if !countsValid(r.Phases) {
		t.Fatal("request_count_conservation_failed")
	}
	if !r.Resource.CPUAvailable || !r.Resource.RSSAvailable || !r.Resource.Available {
		t.Fatal("reaped_child_resource_missing")
	}
	raw, e := json.Marshal(r)
	if e != nil || bytes.Contains(raw, []byte("authored-private-marker")) || bytes.Contains(raw, []byte("authored_private_marker")) {
		t.Fatal("raw_child_diagnostics_retained")
	}
}
func TestResidentLifecycleSyntheticPipeControls(t *testing.T) {
	binary := buildLifecycleChild(t)
	for _, mode := range []string{"normal", "split"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			ready := lifecycleInputs(t, mode)
			r := measureRow(ctx, binary, lifecyclePlan(), ready, 0)
			assertLifecycleCleanup(t, r)
			if r.Status != "complete" || r.Error != "" || !r.EOFObserved || r.ExitCode != 0 || !r.ExitAvailable || !r.CPUPerRequestAvailable || r.CPURequestDenominator != 1045 || r.TimedRTT.Count != 1024 || r.TimedValidation.Count != 1024 || r.TimedValidationWork.Count != 1024 || r.TimedResponseDeliveryGap.Count != 1024 {
				t.Fatalf("synthetic complete protocol row failed: %+v", r)
			}
			for i, want := range [3]int{1, 20, 1024} {
				if r.Phases[i] != (PhaseCounts{Planned: want, Attempted: want, FullyWritten: want, Received: want, Validated: want}) {
					t.Fatal("full_row_phase_count_failed")
				}
			}
			if r.ActualInputStreamSHA256 != r.ExpectedInputStreamSHA256 || r.ActualOutputStreamSHA256 != r.ExpectedOutputStreamSHA256 || r.WaitObservedWallNS <= 0 || r.ExpectedHashPreparationNS <= 0 {
				t.Fatal("stream_identity_or_observation_boundary_missing")
			}
			h := sha256.New()
			// Independently literal first+warm20+timed1024 pool indices. The two
			// phases reset, rather than sharing ordinal modulo count.
			for _, i := range []int{0} {
				h.Write(ready.Payloads[i].Line)
			}
			for i := 0; i < 20; i++ {
				h.Write(ready.Payloads[i%3].Line)
			}
			for i := 0; i < 1024; i++ {
				h.Write(ready.Payloads[i%3].Line)
			}
			if hashBytesOfHash(h.Sum(nil)) != r.ExpectedInputStreamSHA256 {
				t.Fatal("phase_local_pool_sequence_not_pinned")
			}
			if r.TimedValidation.P50NS < r.TimedValidationWork.P50NS || r.TimedValidation.P50NS < r.TimedResponseDeliveryGap.P50NS {
				t.Fatal("receipt_to_validation_window_omits_delivery")
			}
		})
	}
	for _, tc := range []struct {
		mode             string
		phase, validated int
	}{
		{"stall-first", 0, 0}, {"stall-warm", 1, 1}, {"stall-timed", 2, 21},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
			defer cancel()
			r := measureRow(ctx, binary, lifecyclePlan(), lifecycleInputs(t, tc.mode), 0)
			assertLifecycleCleanup(t, r)
			validated := 0
			for _, p := range r.Phases {
				validated += p.Validated
			}
			if validated != tc.validated || r.Phases[tc.phase].Incomplete != 1 || r.Status == "complete" || !r.ContextDeadlineExceeded || r.CPUPerRequestAvailable || r.CPURequestDenominator != 0 {
				t.Fatalf("deadline lost partial counts: %+v", r)
			}
		})
	}
	for _, mode := range []string{"early-exit", "truncated", "oversized", "bad-json", "stderr-overflow", "stderr", "terminal-extra", "terminal-nonzero"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			r := measureRow(ctx, binary, lifecyclePlan(), lifecycleInputs(t, mode), 0)
			assertLifecycleCleanup(t, r)
			if r.Status == "complete" || r.Error == "" || r.CPUPerRequestAvailable || r.CPURequestDenominator != 0 {
				t.Fatal("invalid protocol row became a normal CPU average")
			}
			if mode == "terminal-extra" || mode == "terminal-nonzero" || mode == "stderr" {
				for i, want := range [3]int{1, 20, 1024} {
					if r.Phases[i].Validated != want || r.Phases[i].Failed != 0 || r.Phases[i].Incomplete != 0 {
						t.Fatal("terminal_failure_fabricated_request_failure")
					}
				}
			}
			if mode == "stderr" || mode == "stderr-overflow" {
				if r.StderrBytes <= 0 || r.StderrError == "" {
					t.Fatal("stderr_count_missing")
				}
			}
			if mode == "stderr-overflow" && !r.StderrLimitExceeded {
				t.Fatal("truncated_stderr_count_not_marked")
			}
			if (mode == "terminal-nonzero" || mode == "early-exit") && r.WaitError != "child_exit_failed" {
				t.Fatal("secondary_wait_error_lost")
			}
			if (mode == "truncated" || mode == "oversized") && r.StdoutError == "" {
				t.Fatal("secondary_reader_error_lost")
			}
		})
	}
	t.Run("first_failure_preserves_all_planned_rows", func(t *testing.T) {
		plan := policyPlan()
		report := replayVerified(plan, lifecycleInputs(t, "bad-json"), binary, strings.Repeat("a", 64))
		if report.PlannedRows != 24 || report.PlannedRequests != 25080 || report.Rows[0].Status == "complete" || !report.Rows[0].Started || !report.Rows[0].Reaped || report.Rows[0].WaitCalls != 1 {
			t.Fatal("synthetic stop-on-first-failure control differs")
		}
		for _, row := range report.Rows[1:] {
			if row.Started || row.Status != "not_started" || row.WaitCalls != 0 || row.Error != "prior_row_failure_no_retry" || !countsValid(row.Phases) {
				t.Fatal("first failure discarded or launched a later planned row")
			}
			for _, phase := range row.Phases {
				if phase.Attempted != 0 || phase.NotAttempted != phase.Planned {
					t.Fatal("not-started row fabricated attempted requests")
				}
			}
		}
	})
	t.Run("terminal_eof_does_not_mean_reaped", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r := measureRow(ctx, binary, lifecyclePlan(), lifecycleInputs(t, "terminal-hang"), 0)
		assertLifecycleCleanup(t, r)
		if !r.EOFObserved || r.Status == "complete" || !r.ContextDeadlineExceeded || r.Error == "" || r.CPUPerRequestAvailable {
			t.Fatal("terminal_eof_hid_nonterminated_child")
		}
		for _, p := range r.Phases {
			if p.Validated != p.Planned || p.Failed != 0 || p.Incomplete != 0 {
				t.Fatal("terminal_wait_changed_request_counts")
			}
		}
	})
}

func hashBytesOfHash(raw []byte) string {
	const alphabet = "0123456789abcdef"
	out := make([]byte, len(raw)*2)
	for i, b := range raw {
		out[i*2], out[i*2+1] = alphabet[b>>4], alphabet[b&15]
	}
	return string(out)
}
func TestPhaseLocalSequenceAndPendingCountConservation(t *testing.T) {
	for _, tc := range []struct{ ordinal, index int }{{0, 0}, {1, 0}, {2, 1}, {20, 1}, {21, 0}, {22, 1}, {1044, 0}} {
		if payloadIndex("distinct", tc.ordinal, 3) != tc.index || payloadIndex("same", tc.ordinal, 3) != 0 {
			t.Fatal("phase_local_sequence_drift")
		}
	}
	r := pendingRow(12, RowPlan{Baseline: "bm25", Workload: "same", Repeat: 3})
	if !countsValid(r.Phases) || r.Started || r.Status != "not_started" {
		t.Fatal("unstarted_planned_row_not_preserved")
	}
	changed := r.Phases
	changed[2].NotAttempted--
	if countsValid(changed) {
		t.Fatal("missing_request_count_accepted")
	}
}
func TestReceiveTimelyBufferedFrameSurvivesCancellation(t *testing.T) {
	base := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), base.Add(time.Second))
	cancel()
	frames := make(chan frame, 1)
	frames <- frame{Raw: []byte("{}\n"), Received: base}
	if f, e := receiveFrame(ctx, frames); e != nil || string(f.Raw) != "{}\n" {
		t.Fatal("buffered_observed_response_lost_to_cancel")
	}
	late := make(chan frame, 1)
	late <- frame{Raw: []byte("{}\n"), Received: base.Add(2 * time.Second)}
	if _, e := receiveFrame(ctx, late); e == nil {
		t.Fatal("late_frame_accepted")
	}
	if _, e := receiveFrame(ctx, make(chan frame)); e == nil {
		t.Fatal("cancel_without_observed_frame_accepted")
	}
}

func TestBoundedFramingEOFAndCancelledDeliveryJoin(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		valid       bool
	}{
		{"empty-eof", "", true},
		{"normal", "{}\n", true},
		{"maximum-lf-inclusive", strings.Repeat("x", wireLimit-1) + "\n", true},
		{"empty-line", "\n", false},
		{"partial-eof", "{}", false},
		{"oversized-line", strings.Repeat("x", wireLimit) + "\n", false},
		{"unbounded-no-lf", strings.Repeat("x", 2*wireLimit), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			frames, done := make(chan frame, 1), make(chan readReceipt, 1)
			go drainFrames(ctx, strings.NewReader(tc.input), frames, done)
			first, e := receiveFrame(ctx, frames)
			if e != nil || (first.Error == "") != tc.valid {
				t.Fatal("bounded frame acceptance differs")
			}
			if tc.valid && tc.input != "" {
				if string(first.Raw) != tc.input {
					t.Fatal("framing changed exact response bytes")
				}
				last, e := receiveFrame(ctx, frames)
				if e != nil || !last.EOF {
					t.Fatal("terminal EOF missing")
				}
			}
			select {
			case receipt := <-done:
				if receipt.Bytes < 0 || receipt.Bytes > int64(wireLimit+1) || (receipt.Error == "") != tc.valid || receipt.EOF != tc.valid {
					t.Fatal("bounded reader receipt differs")
				}
				if tc.valid && (receipt.Bytes != int64(len(tc.input)) || receipt.SHA256 != hashBytes([]byte(tc.input))) {
					t.Fatal("complete reader bytes/hash missing")
				}
			case <-ctx.Done():
				t.Fatal("frame reader failed to join")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	frames, done := make(chan frame, 1), make(chan readReceipt, 1)
	go drainFrames(ctx, strings.NewReader("{}\n{}\n{}\n"), frames, done)
	cancel()
	select {
	case <-done:
		if len(frames) > 1 {
			t.Fatal("cancelled reader accumulated an unbounded frame queue")
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled blocked delivery did not join")
	}
}

type shortWriter struct {
	bytes.Buffer
	remaining int
	fail      bool
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if w.remaining == 0 {
		return 0, errors.New("authored-private-marker")
	}
	n := min(len(p), w.remaining)
	w.remaining -= n
	w.Buffer.Write(p[:n])
	if w.fail {
		return n, errors.New("authored-private-marker")
	}
	return n, nil
}
func TestPartialWriteBytesArePreservedAndErrorRedacted(t *testing.T) {
	w := &shortWriter{remaining: 3, fail: true}
	var metered bytes.Buffer
	n, e := writeRequest(w, []byte("abcdef\n"), &metered)
	if n != 3 || metered.String() != "abc" || e == nil || e.Error() != "request_write_failed" {
		t.Fatal("partial_write_was_lost_or_raw_error_escaped")
	}
	w = &shortWriter{remaining: 20}
	metered.Reset()
	if n, e := writeRequest(w, []byte("abcdef\n"), &metered); e != nil || n != 7 || metered.String() != "abcdef\n" {
		t.Fatal("normal_write_failed")
	}
}

func TestResourceAvailabilityAndPlatformUnits(t *testing.T) {
	for _, tc := range []struct {
		platform  string
		raw, want int64
		available bool
		reason    string
	}{
		{"darwin", 123, 123, true, ""}, {"linux", 123, 125952, true, ""},
		{"linux", math.MaxInt64, 0, false, "peak_rss_conversion_overflow"},
		{"darwin", 0, 0, false, "peak_rss_unavailable"}, {"linux", -1, 0, false, "peak_rss_unavailable"},
		{"other", 123, 0, false, "peak_rss_platform_unsupported"},
	} {
		r := resourceRSSForOS(tc.raw, tc.platform)
		if r.RSSAvailable != tc.available || r.LifetimePeakRSSRaw != tc.raw || r.LifetimePeakRSSBytes != tc.want || r.RSSError != tc.reason || r.CPUAvailable || r.Available {
			t.Fatal("platform_unit_or_availability_drift")
		}
	}
	if r := resourceForState(nil); r.Available || r.CPUAvailable || r.RSSAvailable || r.CPUError == "" || r.RSSError == "" {
		t.Fatal("missing_rusage_became_zero_measurement")
	}
	before, after := syscall.Rusage{}, syscall.Rusage{}
	before.Utime.Sec, after.Utime.Sec = 2, 3
	before.Utime.Usec, after.Utime.Usec = 500000, 100000
	before.Stime.Sec, after.Stime.Sec = 1, 1
	before.Stime.Usec, after.Stime.Usec = 100000, 300000
	after.Maxrss = 123
	r := selfResource(before, after, true)
	if !r.Available || math.Abs(r.UserCPUSeconds-.6) > 1e-12 || math.Abs(r.SystemCPUSeconds-.2) > 1e-12 || r.LifetimePeakRSSRaw != 123 {
		t.Fatal("self_cpu_delta_or_lifetime_peak_incorrect")
	}
	if r := selfResource(before, after, false); r.Available || r.CPUAvailable || r.RSSAvailable {
		t.Fatal("failed_getrusage_became_available")
	}
	after.Utime.Sec = 0
	if r := selfResource(before, after, true); r.CPUAvailable || r.Available || !r.RSSAvailable {
		t.Fatal("invalid_cpu_lost_independent_rss_availability")
	}
}

func TestSecondaryReceiptsKeepPrimaryFailureAndFixedDiagnostics(t *testing.T) {
	r := pendingRow(0, RowPlan{})
	r.Error = "response_contract_invalid"
	acceptWait(&r, waitReceipt{Error: errors.New("authored-private-marker"), Observed: time.Now()}, time.Now())
	acceptReadReceipt(&r, readReceipt{Bytes: 3, SHA256: hashBytes([]byte("x\n")), Error: "response_framing_or_io_failed"})
	acceptStderrReceipt(&r, stderrReceipt{Bytes: 4097, Overflow: true, IOFailure: true})
	if r.Error != "response_contract_invalid" || r.WaitError != "child_exit_failed" || r.StdoutError != "response_framing_or_io_failed" || r.StderrError != "child_stderr_limit_exceeded" || !r.StderrLimitExceeded {
		t.Fatal("primary or secondary error evidence was lost")
	}
	raw, e := json.Marshal(r)
	if e != nil || strings.Contains(string(raw), "authored-private-marker") {
		t.Fatal("raw worker diagnostic escaped fixed receipt errors")
	}
}
func TestPreStartFailureKeepsAllRequestsUnattempted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := measureRow(ctx, "unused", lifecyclePlan(), lifecycleInputs(t, "normal"), 0)
	if r.Started || r.WaitCalls != 0 || r.Status != "not_started" || r.Error == "" || !countsValid(r.Phases) {
		t.Fatal("prestart_cancellation_lost_planned_counts")
	}
	r = measureRow(context.Background(), filepath.Join(t.TempDir(), "missing"), lifecyclePlan(), lifecycleInputs(t, "normal"), 0)
	if r.Started || r.WaitCalls != 0 || r.Error != "child_start_failed" || !countsValid(r.Phases) || r.Resource.CPUAvailable || r.Resource.RSSAvailable {
		t.Fatal("start_failure_fabricated_resource_or_requests")
	}
}
func TestReadyCopiesOwnWireAndRegistryArrays(t *testing.T) {
	ready := lifecycleInputs(t, "normal")
	copied := slices.Clone(ready.Payloads[0].Line)
	ready.Payloads[0].Payload.Wire[0] = 'x'
	if !bytes.Equal(ready.Payloads[0].Line, copied) {
		t.Fatal("prepared_wire_slices_share_storage")
	}
	kinds, ids := baselineKinds(), candidateIDs()
	kinds[0], ids[0] = "changed", "changed"
	if baselineKinds()[0] != "fixed_order" || candidateIDs()[0] != "candidate-0" {
		t.Fatal("registry_exposes_shared_mutation")
	}
	if strings.Join(childEnvironment(), "\n") != "GOMAXPROCS=1\nGOMEMLIMIT=256MiB\nGODEBUG=\nGOTRACEBACK=none" {
		t.Fatal("child_environment_scope_changed")
	}
}
