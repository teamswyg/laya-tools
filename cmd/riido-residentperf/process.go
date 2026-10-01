package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"syscall"
	"time"
)

type ReadyPayload struct {
	Payload   Payload
	Line      []byte
	Responses [4][]byte
}
type ReadyCorpus struct{ Payloads []ReadyPayload }

func prepareWireTable(c Corpus) (ReadyCorpus, error) {
	out := ReadyCorpus{Payloads: make([]ReadyPayload, len(c.Payloads))}
	for i, p := range c.Payloads {
		prepared, e := loadPayload(p.Wire)
		if e != nil {
			return ReadyCorpus{}, e
		}
		out.Payloads[i].Payload = p
		out.Payloads[i].Payload.Wire = slices.Clone(p.Wire)
		out.Payloads[i].Line = append(slices.Clone(p.Wire), '\n')
		for j, kind := range baselineKinds() {
			out.Payloads[i].Responses[j], _, e = expectedResponse(prepared, kind)
			if e != nil {
				return ReadyCorpus{}, e
			}
		}
	}
	return out, nil
}

// Each phase starts at pool index zero. The timed requests cycle a bounded pool.
func payloadIndex(workload string, ordinal, count int) int {
	if workload == "same" || count <= 0 {
		return 0
	}
	if ordinal >= 21 {
		ordinal -= 21
	} else if ordinal > 0 {
		ordinal--
	}
	return ordinal % count
}
func phaseIndex(ordinal int) int {
	if ordinal == 0 {
		return 0
	}
	if ordinal < 21 {
		return 1
	}
	return 2
}
func pendingRow(index int, plan RowPlan) RowResult {
	return RowResult{Index: index, Plan: plan, Status: "not_started", Resource: unavailableResource(), Phases: [3]PhaseCounts{{Planned: 1, NotAttempted: 1}, {Planned: 20, NotAttempted: 20}, {Planned: timedPerRow, NotAttempted: timedPerRow}}}
}
func countsValid(phases [3]PhaseCounts) bool {
	for i, p := range phases {
		if p.Planned != [3]int{1, 20, timedPerRow}[i] || p.Attempted < 0 || p.NotAttempted < 0 || p.Failed < 0 || p.Incomplete < 0 || p.Validated < 0 || p.FullyWritten < 0 || p.Received < 0 || p.Planned != p.Attempted+p.NotAttempted || p.Attempted != p.Validated+p.Failed+p.Incomplete || p.Received > p.FullyWritten || p.FullyWritten > p.Attempted || p.Validated > p.Received {
			return false
		}
	}
	return true
}

type frame struct {
	Raw      []byte
	Received time.Time
	Error    string
	EOF      bool
}
type readReceipt struct {
	Bytes  int64
	SHA256 string
	EOF    bool
	Error  string
}
type meter struct {
	Source io.Reader
	Hash   io.Writer
	Bytes  int64
}

func (m *meter) Read(p []byte) (int, error) {
	n, e := m.Source.Read(p)
	if n > 0 {
		m.Bytes += int64(n)
		_, _ = m.Hash.Write(p[:n])
	}
	return n, e
}

// One reader owns stdout hashing/framing. A capacity-one receipt channel is
// bounded delivery, not an open-loop queue. Never accumulate whole responses.
func drainFrames(ctx context.Context, input io.Reader, frames chan<- frame, finished chan<- readReceipt) {
	h := sha256.New()
	m := meter{Source: input, Hash: h}
	receipt := readReceipt{}
	defer func() { receipt.Bytes, receipt.SHA256 = m.Bytes, hex.EncodeToString(h.Sum(nil)); finished <- receipt }()
	reader := bufio.NewReaderSize(&m, wireLimit+1)
	send := func(f frame) bool {
		select {
		case frames <- f:
			return true
		default:
		}
		select {
		case frames <- f:
			return true
		case <-ctx.Done():
			return false
		}
	}
	for {
		raw, e := reader.ReadSlice('\n')
		at := time.Now()
		if e == io.EOF && len(raw) == 0 {
			receipt.EOF = true
			send(frame{EOF: true, Received: at})
			return
		}
		if e != nil || len(raw) < 2 || len(raw) > wireLimit || raw[len(raw)-1] != '\n' {
			receipt.Error = "response_framing_or_io_failed"
			send(frame{Error: receipt.Error, Received: at})
			return
		}
		if m.Bytes > int64((requestsPerRow+1)*wireLimit) {
			receipt.Error = "child_total_output_limit"
			send(frame{Error: receipt.Error, Received: at})
			return
		}
		if !send(frame{Raw: slices.Clone(raw), Received: at}) {
			receipt.Error = "reader_cancelled_before_delivery"
			return
		}
	}
}

// Preserve complete already-observed timely frames despite a simultaneous
// context wakeup. Never accept a frame stamped after the effective deadline.
func receiveFrame(ctx context.Context, frames <-chan frame) (frame, error) {
	check := func(f frame) (frame, error) {
		if deadline, ok := ctx.Deadline(); ok && f.Received.After(deadline) {
			return frame{}, errors.New("child_or_global_deadline")
		}
		return f, nil
	}
	select {
	case f := <-frames:
		return check(f)
	default:
	}
	select {
	case f := <-frames:
		return check(f)
	case <-ctx.Done():
		select {
		case f := <-frames:
			return check(f)
		default:
			return frame{}, errors.New("child_or_global_deadline")
		}
	}
}

type waitReceipt struct {
	Error    error
	State    *os.ProcessState
	Observed time.Time
}
type stderrReceipt struct {
	Bytes     int64
	Overflow  bool
	IOFailure bool
}

func unavailableResource() Resource {
	return Resource{CPUError: "cpu_usage_unavailable", RSSError: "peak_rss_unavailable"}
}
func finiteNonnegative(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 }
func resourceForState(state *os.ProcessState) Resource {
	if state == nil {
		return unavailableResource()
	}
	r := unavailableResource()
	r.UserCPUSeconds, r.SystemCPUSeconds = state.UserTime().Seconds(), state.SystemTime().Seconds()
	r.CPUAvailable = finiteNonnegative(r.UserCPUSeconds) && finiteNonnegative(r.SystemCPUSeconds)
	if r.CPUAvailable {
		r.CPUError = ""
	}
	if u, ok := state.SysUsage().(*syscall.Rusage); ok {
		rss := resourceRSS(u.Maxrss)
		r.RSSAvailable, r.RSSError = rss.RSSAvailable, rss.RSSError
		r.LifetimePeakRSSRaw, r.LifetimePeakRSSUnit, r.LifetimePeakRSSBytes = rss.LifetimePeakRSSRaw, rss.LifetimePeakRSSUnit, rss.LifetimePeakRSSBytes
	}
	r.Available = r.CPUAvailable && r.RSSAvailable
	return r
}
func resourceRSS(raw int64) Resource { return resourceRSSForOS(raw, runtime.GOOS) }
func resourceRSSForOS(raw int64, platform string) Resource {
	r := Resource{LifetimePeakRSSRaw: raw, RSSError: "peak_rss_unavailable"}
	if platform == "darwin" {
		r.LifetimePeakRSSUnit = "bytes"
	} else if platform == "linux" {
		r.LifetimePeakRSSUnit = "KiB"
	} else {
		r.RSSError = "peak_rss_platform_unsupported"
		return r
	}
	if raw <= 0 {
		return r
	}
	if platform == "linux" {
		if raw > math.MaxInt64/1024 {
			r.RSSError = "peak_rss_conversion_overflow"
			return r
		}
		r.LifetimePeakRSSBytes = raw * 1024
	} else {
		r.LifetimePeakRSSBytes = raw
	}
	r.RSSAvailable, r.RSSError = true, ""
	return r
}
func selfUsage() (syscall.Rusage, bool) {
	var u syscall.Rusage
	e := syscall.Getrusage(syscall.RUSAGE_SELF, &u)
	return u, e == nil
}
func selfResource(before, after syscall.Rusage, valid bool) Resource {
	if !valid {
		return unavailableResource()
	}
	r := resourceRSS(after.Maxrss)
	r.UserCPUSeconds = float64(after.Utime.Sec-before.Utime.Sec) + float64(after.Utime.Usec-before.Utime.Usec)/1e6
	r.SystemCPUSeconds = float64(after.Stime.Sec-before.Stime.Sec) + float64(after.Stime.Usec-before.Stime.Usec)/1e6
	r.CPUAvailable = finiteNonnegative(r.UserCPUSeconds) && finiteNonnegative(r.SystemCPUSeconds)
	if !r.CPUAvailable {
		r.CPUError = "cpu_usage_delta_invalid"
	}
	r.Available = r.CPUAvailable && r.RSSAvailable
	return r
}
func childEnvironment() []string {
	return []string{"GOMAXPROCS=1", "GOMEMLIMIT=256MiB", "GODEBUG=", "GOTRACEBACK=none"}
}
func setPrimaryError(r *RowResult, reason string) {
	if r.Error == "" {
		r.Error = reason
	}
}
func setCleanupError(r *RowResult, reason string) {
	if r.CleanupError == "" {
		r.CleanupError = reason
	}
	setPrimaryError(r, reason)
}

// Partial writes count real bytes: no assumption that a failed child saw none.
func writeRequest(w io.Writer, line []byte, meter io.Writer) (int, error) {
	written := 0
	for written < len(line) {
		n, e := w.Write(line[written:])
		if n < 0 || n > len(line)-written {
			return written, errors.New("request_write_failed")
		}
		if n > 0 {
			_, _ = meter.Write(line[written : written+n])
			written += n
		}
		if e != nil || n == 0 {
			return written, errors.New("request_write_failed")
		}
	}
	return written, nil
}
func acceptWait(r *RowResult, receipt waitReceipt, start time.Time) {
	r.Reaped, r.ExitAvailable = true, receipt.State != nil
	r.WaitObservedWallNS = receipt.Observed.Sub(start).Nanoseconds()
	if receipt.State != nil {
		r.ExitCode, r.Resource = receipt.State.ExitCode(), resourceForState(receipt.State)
	}
	if receipt.Error != nil {
		r.WaitError = "child_exit_failed"
		setPrimaryError(r, "child_exit_failed")
	}
}

func acceptReadReceipt(r *RowResult, receipt readReceipt) {
	r.StdoutJoined, r.StdoutReceiptAvailable, r.OutputBytes, r.ActualOutputStreamSHA256 = true, true, receipt.Bytes, receipt.SHA256
	r.StdoutError = receipt.Error
	if receipt.Error != "" {
		setPrimaryError(r, receipt.Error)
	}
}

func acceptStderrReceipt(r *RowResult, receipt stderrReceipt) {
	// Bytes is the observed prefix, capped at 4,097; it is not claimed to be
	// total child stderr when overflow or IO failure prevented complete drain.
	r.StderrJoined, r.StderrReceiptAvailable, r.StderrBytes, r.StderrLimitExceeded = true, true, receipt.Bytes, receipt.Overflow
	if receipt.Overflow {
		r.StderrError = "child_stderr_limit_exceeded"
	} else if receipt.IOFailure {
		r.StderrError = "child_stderr_io_failed"
	} else if receipt.Bytes != 0 {
		r.StderrError = "child_stderr_not_empty"
	}
	if r.StderrError != "" {
		setPrimaryError(r, "child_stderr_or_io_failed")
	}
}

func measureRow(parent context.Context, executable string, plan Plan, ready ReadyCorpus, index int) RowResult {
	if index < 0 || index >= rowCount {
		return RowResult{Index: index, Status: "not_started", Error: "prepared_dispatch_invalid", Resource: unavailableResource()}
	}
	r := pendingRow(index, plan.Rows[index])
	r.Resource = unavailableResource()
	kinds := baselineKinds()
	k := slices.Index(kinds[:], r.Plan.Baseline)
	if k < 0 || r.Plan.Workload != "same" && r.Plan.Workload != "distinct" || len(ready.Payloads) < 2 || len(ready.Payloads) > 72 || plan.ChildTimeoutSeconds != 15 || plan.CleanupGraceMilliseconds != 1000 {
		r.Error = "prepared_dispatch_invalid"
		return r
	}
	if parent == nil || parent.Err() != nil {
		r.Error = "global_deadline_before_start"
		return r
	}
	// Before child lifetime, but INSIDE global replay CPU/wall. This cost is
	// explicit preparation, not silently counted as child warmup or omitted.
	hashStart := time.Now()
	expectedInput, expectedOutput := sha256.New(), sha256.New()
	for ordinal := 0; ordinal < requestsPerRow; ordinal++ {
		p := ready.Payloads[payloadIndex(r.Plan.Workload, ordinal, len(ready.Payloads))]
		if len(p.Line) < 2 || len(p.Line) > wireLimit || p.Line[len(p.Line)-1] != '\n' || len(p.Responses[k]) == 0 {
			r.Error = "prepared_dispatch_invalid"
			return r
		}
		_, _ = expectedInput.Write(p.Line)
		_, _ = expectedOutput.Write(p.Responses[k])
	}
	r.ExpectedHashPreparationNS = time.Since(hashStart).Nanoseconds()
	r.ExpectedInputStreamSHA256, r.ExpectedOutputStreamSHA256 = hex.EncodeToString(expectedInput.Sum(nil)), hex.EncodeToString(expectedOutput.Sum(nil))
	if parent.Err() != nil {
		r.Error = "global_deadline_before_start"
		return r
	}
	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()
	stdinR, stdinW, e := os.Pipe()
	if e != nil {
		r.Error = "pipe_create_failed"
		return r
	}
	defer stdinR.Close()
	defer stdinW.Close()
	stdoutR, stdoutW, e := os.Pipe()
	if e != nil {
		r.Error = "pipe_create_failed"
		return r
	}
	defer stdoutR.Close()
	defer stdoutW.Close()
	stderrR, stderrW, e := os.Pipe()
	if e != nil {
		r.Error = "pipe_create_failed"
		return r
	}
	defer stderrR.Close()
	defer stderrW.Close()
	cmd := exec.CommandContext(ctx, executable, "--stream", "--baseline", r.Plan.Baseline)
	cmd.Env, cmd.Stdin, cmd.Stdout, cmd.Stderr = childEnvironment(), stdinR, stdoutW, stderrW
	start := time.Now()
	e = cmd.Start()
	r.SpawnCallNS = time.Since(start).Nanoseconds()
	if e != nil {
		r.Error = "child_start_failed"
		r.LifetimeWallNS = time.Since(start).Nanoseconds()
		return r
	}
	r.Started, r.Status = true, "incomplete"
	_ = stdinR.Close()
	_ = stdoutW.Close()
	_ = stderrW.Close()
	frames, readerDone := make(chan frame, 1), make(chan readReceipt, 1)
	go drainFrames(ctx, stdoutR, frames, readerDone)
	stderrDone := make(chan stderrReceipt, 1)
	go func() {
		n, err := io.CopyN(io.Discard, stderrR, 4097)
		stderrDone <- stderrReceipt{Bytes: n, Overflow: n > 4096, IOFailure: err != nil && !errors.Is(err, io.EOF)}
		if n > 4096 {
			cancel()
		}
	}()
	waitDone := make(chan waitReceipt, 1)
	r.WaitCalls = 1
	// Exactly one Wait owner. No other goroutine reads cmd.ProcessState.
	go func() {
		err := cmd.Wait()
		waitDone <- waitReceipt{Error: err, State: cmd.ProcessState, Observed: time.Now()}
	}()
	watcherStop, watcherDone := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(watcherDone)
		select {
		case <-ctx.Done():
			_ = stdinW.Close()
			_ = stdoutR.Close()
			_ = stderrR.Close()
		case <-watcherStop:
		}
	}()
	actualInput := sha256.New()
	var rtt, blocking, validation, work, gap [timedPerRow]int64
	timedCount := 0
	for ordinal := 0; ordinal < requestsPerRow && r.Error == ""; ordinal++ {
		// A timely buffered preceding response may be counted, but cancellation
		// never starts a new request.
		if ctx.Err() != nil {
			setPrimaryError(&r, "child_or_global_deadline")
			break
		}
		phase := &r.Phases[phaseIndex(ordinal)]
		phase.Attempted++
		phase.NotAttempted--
		p := ready.Payloads[payloadIndex(r.Plan.Workload, ordinal, len(ready.Payloads))]
		wireStart := time.Now()
		written, e := writeRequest(stdinW, p.Line, actualInput)
		r.InputBytes += int64(written)
		writeNS := time.Since(wireStart).Nanoseconds()
		if e != nil {
			phase.Failed++
			setPrimaryError(&r, "request_write_failed")
			break
		}
		phase.FullyWritten++
		f, e := receiveFrame(ctx, frames)
		if e != nil {
			phase.Incomplete++
			setPrimaryError(&r, "child_or_global_deadline")
			break
		}
		if f.Error != "" {
			phase.Failed++
			setPrimaryError(&r, f.Error)
			break
		}
		if f.EOF {
			phase.Incomplete++
			setPrimaryError(&r, "response_missing_eof")
			break
		}
		phase.Received++
		if f.Received.Before(wireStart) {
			phase.Failed++
			setPrimaryError(&r, "unsolicited_response")
			break
		}
		if ordinal == 0 {
			r.StartupAndFirstResponseNS = f.Received.Sub(start).Nanoseconds()
		}
		validationStart := time.Now()
		if e := checkResponse(f.Raw, p.Payload, r.Plan.Baseline, p.Payload.Expected[k]); e != nil {
			phase.Failed++
			setPrimaryError(&r, e.Error())
			break
		}
		validatedAt := time.Now()
		phase.Validated++
		if phaseIndex(ordinal) == 2 {
			rtt[timedCount], blocking[timedCount] = f.Received.Sub(wireStart).Nanoseconds(), writeNS
			validation[timedCount], work[timedCount], gap[timedCount] = validatedAt.Sub(f.Received).Nanoseconds(), validatedAt.Sub(validationStart).Nanoseconds(), validationStart.Sub(f.Received).Nanoseconds()
			timedCount++
		}
	}
	if r.Error == "" {
		_ = stdinW.Close()
		f, e := receiveFrame(ctx, frames)
		if e != nil {
			setPrimaryError(&r, "terminal_eof_deadline")
		} else if !f.EOF || f.Error != "" {
			setPrimaryError(&r, "extra_or_incomplete_terminal_response")
		} else {
			r.EOFObserved = true
		}
	}
	// EOF is not reaping. Await process termination until execution deadline.
	if r.Error == "" {
		select {
		case receipt := <-waitDone:
			acceptWait(&r, receipt, start)
		case <-ctx.Done():
			select {
			case receipt := <-waitDone:
				acceptWait(&r, receipt, start)
			default:
				setPrimaryError(&r, "terminal_wait_deadline")
			}
		}
	}
	cleanupStart := time.Now()
	if r.Error != "" {
		cancel()
		_ = cmd.Process.Kill()
		_ = stdinW.Close()
		_ = stdoutR.Close()
		_ = stderrR.Close()
	}
	close(watcherStop)
	// Wait/stdout/stderr/watcher share ONE total grace, never 3 independent ones.
	cleanupTimer := time.NewTimer(time.Second)
	defer cleanupTimer.Stop()
	cleanupExpired := false
	if !r.Reaped {
		select {
		case receipt := <-waitDone:
			acceptWait(&r, receipt, start)
		case <-cleanupTimer.C:
			cleanupExpired = true
			setCleanupError(&r, "child_not_reaped")
		}
	}
	if !cleanupExpired {
		select {
		case receipt := <-readerDone:
			acceptReadReceipt(&r, receipt)
		case <-cleanupTimer.C:
			cleanupExpired = true
			setCleanupError(&r, "stdout_reader_not_joined")
		}
	}
	if !cleanupExpired {
		select {
		case receipt := <-stderrDone:
			acceptStderrReceipt(&r, receipt)
		case <-cleanupTimer.C:
			cleanupExpired = true
			setCleanupError(&r, "stderr_reader_not_joined")
		}
	}
	if !cleanupExpired {
		select {
		case <-watcherDone:
			r.WatcherJoined = true
		case <-cleanupTimer.C:
			cleanupExpired = true
			setCleanupError(&r, "cancellation_watcher_not_joined")
		}
	}
	if cleanupExpired {
		cancel()
		_ = cmd.Process.Kill()
		_ = stdinW.Close()
		_ = stdoutR.Close()
		_ = stderrR.Close()
		// Recover already-available evidence without another grace.
		if !r.Reaped {
			select {
			case receipt := <-waitDone:
				acceptWait(&r, receipt, start)
			default:
			}
		}
		if !r.StdoutJoined {
			select {
			case receipt := <-readerDone:
				acceptReadReceipt(&r, receipt)
			default:
			}
		}
		if !r.StderrJoined {
			select {
			case receipt := <-stderrDone:
				acceptStderrReceipt(&r, receipt)
			default:
			}
		}
		select {
		case <-watcherDone:
			r.WatcherJoined = true
		default:
		}
	}
	r.CleanupWallNS = time.Since(cleanupStart).Nanoseconds()
	r.CleanupComplete = r.Reaped && r.StdoutJoined && r.StderrJoined && r.WatcherJoined && !cleanupExpired
	r.ActualInputStreamSHA256 = hex.EncodeToString(actualInput.Sum(nil))
	r.LifetimeWallNS = time.Since(start).Nanoseconds()
	r.Cancelled, r.ContextDeadlineExceeded = ctx.Err() != nil, errors.Is(ctx.Err(), context.DeadlineExceeded)
	r.ChildDeadlineExceeded = r.LifetimeWallNS > int64(plan.ChildTimeoutSeconds)*int64(time.Second)
	if r.Error == "" && (!countsValid(r.Phases) || !r.Resource.Available || !r.CleanupComplete || !r.EOFObserved || r.ExitCode != 0 || timedCount != timedPerRow || r.ActualInputStreamSHA256 != r.ExpectedInputStreamSHA256 || r.ActualOutputStreamSHA256 != r.ExpectedOutputStreamSHA256 || r.ChildDeadlineExceeded || r.Cancelled) {
		r.Error = "terminal_resource_budget_or_stream_mismatch"
	}
	if r.Error == "" {
		r.Status, r.CPUPerRequestAvailable, r.CPURequestDenominator = "complete", true, requestsPerRow
		r.CPUSecondsPerRequest = (r.Resource.UserCPUSeconds + r.Resource.SystemCPUSeconds) / requestsPerRow
	}
	r.TimedRTT, r.TimedWrite, r.TimedValidation = distribution(rtt[:timedCount]), distribution(blocking[:timedCount]), distribution(validation[:timedCount])
	r.TimedValidationWork, r.TimedResponseDeliveryGap = distribution(work[:timedCount]), distribution(gap[:timedCount])
	return r
}
