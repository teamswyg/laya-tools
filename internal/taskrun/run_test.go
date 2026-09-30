package taskrun

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/internal/taskverify"
)

// These authored shell fixtures exercise owned process/capture behavior only.
// They are not a Codex inference or evidence of real sandbox enforcement.
func fixtureRequest(t *testing.T, body string) Request {
	t.Helper()
	root := t.TempDir()
	base, e := filepath.Abs("../taskverify/testdata/base")
	if e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(root, "authored-cli")
	script := `#!/bin/sh
if [ "$1" = "--no-daemon" ]; then shift; fi
if [ "$1" = "--version" ]; then printf 'codex-cli 0.158.0\n'; exit 0; fi
if [ "$1" = "sandbox" ]; then printf public > .riido-permission-probe; exit 0; fi
` + body
	if os.WriteFile(binary, []byte(script), 0700) != nil {
		t.Fatal("fixture_write_failed")
	}
	req := Request{Execute: true, TaskID: "comment-preview-authority", Model: "fixture-model", Reasoning: "low", BaseDir: base, PrivateDir: filepath.Join(root, "run"), CodexBinary: binary, Timeout: 2 * time.Second, ExpectedCLIHash: digest([]byte(script)), ExpectedCLIVersion: SupportedVersion, AttemptOrdinal: 1}
	spec, _ := taskverify.TaskSpec(req.TaskID)
	req.ExpectedSpecSHA256 = digestJSON(spec)
	writePlan(t, &req, root)
	return req
}

func writePlan(t *testing.T, req *Request, root string) {
	t.Helper()
	p := plan{Schema: "riido-task-outcome-pilot-plan-v1", Status: "precommitted_development_pilot_not_final_routing_evaluation", PublicBase: taskverify.BaseRevision}
	p.CLI.Hash = req.ExpectedCLIHash
	p.CLI.Version = req.ExpectedCLIVersion
	p.Profiles = append(p.Profiles, struct {
		ID        string `json:"id"`
		Model     string `json:"requested_model"`
		Reasoning string `json:"requested_reasoning"`
	}{"fixture-low", req.Model, req.Reasoning})
	p.Tasks = append(p.Tasks, struct {
		ID   string `json:"id"`
		Spec string `json:"spec_sha256"`
	}{req.TaskID, req.ExpectedSpecSHA256})
	p.Attempts = append(p.Attempts, struct {
		Ordinal int    `json:"ordinal"`
		Task    string `json:"task"`
		Profile string `json:"profile"`
	}{1, req.TaskID, "fixture-low"})
	p.Execution.Max = 1
	p.Execution.Concurrency = 1
	p.Execution.Timeout = int(req.Timeout / time.Second)
	p.Execution.Verify = 45
	p.Execution.Stdout = 64 << 20
	p.Execution.Stderr = MaxStderrBytes
	p.Execution.Fresh = true
	b, e := json.Marshal(p)
	if e != nil {
		t.Fatal(e)
	}
	req.PlanFile = filepath.Join(root, "authored-plan.json")
	req.PlanSHA256 = digest(b)
	if os.WriteFile(req.PlanFile, b, 0600) != nil {
		t.Fatal("plan_write_failed")
	}
}

const completeTrace = `printf '%s\n' '{"type":"thread.started","thread_id":"private-fixture-id"}' '{"type":"turn.started"}' '{"type":"item.completed","item":{"text":"private fixture body must never escape"}}' '{"type":"turn.completed","usage":{"input_tokens":123,"cached_input_tokens":45,"output_tokens":6}}'
`

func requireDarwin(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("actual owned fixture launch is intentionally unavailable on other OS")
	}
}

func TestOwnedUnchangedCompleteTraceIsRejected(t *testing.T) {
	requireDarwin(t)
	req := fixtureRequest(t, completeTrace)
	r, e := Run(context.Background(), req)
	if e != nil {
		t.Fatal(e)
	}
	if !r.Started || r.ProcessStatus != "exited_zero" || r.GroupCleanup != "group_terminated" || r.VerificationStatus != "rejected" || r.Verification.Accepted || !r.WholeAttemptUsageComplete {
		t.Fatalf("wrong owned outcome: %+v", r)
	}
	if r.Summary == nil || r.Summary.ObservedModel != "unknown" || r.Summary.ProcessExit != "unknown" || r.Summary.TaskAcceptance != "unknown" || r.ObservedModel != "unknown" {
		t.Fatal("imported summary upgraded to execution claims")
	}
	if r.Stdout.SHA256 != r.Summary.TraceSHA256 || r.Stdout.Bytes != r.Summary.TraceBytes {
		t.Fatal("trace binding mismatch")
	}
	b, _ := json.Marshal(r)
	for _, secret := range []string{"private-fixture-id", "private fixture body", req.PrivateDir, req.BaseDir, req.CodexBinary} {
		if bytes.Contains(b, []byte(secret)) {
			t.Fatal("private content escaped")
		}
	}
	st, _ := os.Stat(req.PrivateDir)
	if st.Mode().Perm() != 0700 {
		t.Fatal("private permissions")
	}
	for _, name := range []string{"stdout.jsonl", "stderr.txt", "record.json", "started-record.json"} {
		st, e := os.Stat(filepath.Join(req.PrivateDir, name))
		if e != nil || st.Mode().Perm() != 0600 {
			t.Fatal("private file permissions")
		}
	}
	if !r.DurableStart || r.AuthCleanup != "removed_or_not_present" || r.VerificationWallMillis == nil {
		t.Fatal("owned durability or cleanup absent")
	}
}

func TestAuthoredCorrectCommentAcceptedIndependently(t *testing.T) {
	requireDarwin(t)
	body := `/usr/bin/sed -i '' 's#// candidate, suggest, abstain; never execution authority#// candidate, suggest, or abstain; a suggestion never authorizes execution.#' pkg/reporouter/router.go
printf 'unassessed' > extra-authored-file
` + completeTrace
	req := fixtureRequest(t, body)
	r, e := Run(context.Background(), req)
	if e != nil {
		t.Fatal(e)
	}
	if r.VerificationStatus != "accepted" || !r.Verification.Accepted || r.CandidateStatus != "captured_task_closure" || r.CandidateFileCount != 1 || r.OutsideClosure != "unassessed_including_runtime_cache_and_added_files" {
		t.Fatalf("wrong independent outcome: %+v", r)
	}
	if _, e = os.Stat(filepath.Join(req.PrivateDir, "candidate", "extra-authored-file")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("outside closure snapshot")
	}
}

func TestNonzeroCompleteTraceDoesNotMeanWholeAttemptUsage(t *testing.T) {
	requireDarwin(t)
	req := fixtureRequest(t, completeTrace+"exit 7\n")
	r, e := Run(context.Background(), req)
	if e != nil {
		t.Fatal(e)
	}
	if r.ProcessStatus != "nonzero_exit" || r.ExitCode == nil || *r.ExitCode != 7 || r.Summary == nil || !r.Summary.UsageComplete || r.WholeAttemptUsageComplete || r.VerificationStatus != "rejected" {
		t.Fatal("complete supplied trace misrepresented as complete successful attempt")
	}
}

func TestFailedAndMalformedTracesPreserveStartedRecords(t *testing.T) {
	requireDarwin(t)
	for _, tc := range []struct {
		body, status string
		summary      bool
	}{{`printf '%s\n' '{"type":"thread.started"}' '{"type":"turn.started"}' '{"type":"turn.failed"}'`, "parsed", true}, {`printf 'private malformed body'`, "parse_failed", false}, {`printf '%s\n' '{"type":"thread.started"}' '{"type":"turn.started"}'`, "parsed", true}} {
		t.Run(tc.status+tc.body[:7], func(t *testing.T) {
			req := fixtureRequest(t, tc.body)
			r, e := Run(context.Background(), req)
			if e != nil {
				t.Fatal(e)
			}
			if !r.Started || r.SummaryStatus != tc.status || (r.Summary != nil) != tc.summary || r.WholeAttemptUsageComplete {
				t.Fatal("failed usage invented")
			}
			if _, e = os.Stat(filepath.Join(req.PrivateDir, "record.json")); e != nil {
				t.Fatal("record lost after launch")
			}
		})
	}
}

func TestExclusiveDirectoryAndPinsRefuseBeforeLaunch(t *testing.T) {
	req := fixtureRequest(t, completeTrace)
	for _, tc := range []struct {
		name   string
		mutate func(*Request)
	}{{"missing_execute", func(r *Request) { r.Execute = false }}, {"bad_plan_hash", func(r *Request) { r.PlanSHA256 = strings.Repeat("0", 64) }}, {"bad_spec", func(r *Request) { r.ExpectedSpecSHA256 = strings.Repeat("0", 64) }}, {"wrong_ordinal", func(r *Request) { r.AttemptOrdinal = 2 }}, {"profile_mismatch", func(r *Request) { r.Model = "other-fixture-model" }}, {"bad_cli", func(r *Request) { r.ExpectedCLIHash = strings.Repeat("0", 64) }}} {
		t.Run(tc.name, func(t *testing.T) {
			copy := req
			tc.mutate(&copy)
			r, e := Run(context.Background(), copy)
			if e == nil || r.Started {
				t.Fatal("prelaunch pin accepted")
			}
			if strings.Contains(e.Error(), req.PrivateDir) {
				t.Fatal("error disclosed path")
			}
		})
	}
	if runtime.GOOS != "darwin" {
		return
	}
	if os.Mkdir(req.PrivateDir, 0700) != nil {
		t.Fatal("mkdir")
	}
	r, e := Run(context.Background(), req)
	if e != Error("private_directory_must_be_new") || r.Started {
		t.Fatal("existing private directory accepted")
	}
}

func TestPermissionProbeFailureRefusesBeforeAttempt(t *testing.T) {
	requireDarwin(t)
	req := fixtureRequest(t, completeTrace)
	b, e := os.ReadFile(req.CodexBinary)
	if e != nil {
		t.Fatal(e)
	}
	b = bytes.Replace(b, []byte(`printf public > .riido-permission-probe; exit 0`), []byte(`exit 1`), 1)
	os.WriteFile(req.CodexBinary, b, 0700)
	req.ExpectedCLIHash = digest(b)
	writePlan(t, &req, filepath.Dir(req.PrivateDir))
	r, e := Run(context.Background(), req)
	if e != Error("permission_probe_failed") || r.Started {
		t.Fatal("failed probe launched main attempt")
	}
	if _, e = os.Stat(filepath.Join(req.PrivateDir, "stdout.jsonl")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("main capture exists on refused probe")
	}
}

func TestBinaryChangeDuringProbeRefusesMainAttempt(t *testing.T) {
	requireDarwin(t)
	req := fixtureRequest(t, completeTrace)
	b, e := os.ReadFile(req.CodexBinary)
	if e != nil {
		t.Fatal(e)
	}
	// Authored trusted helper changes its own original executable in preflight.
	b = bytes.Replace(b, []byte(`printf public > .riido-permission-probe; exit 0`), []byte(`printf public > .riido-permission-probe; printf '\n# fixture changed\n' >> "$0"; exit 0`), 1)
	os.WriteFile(req.CodexBinary, b, 0700)
	req.ExpectedCLIHash = digest(b)
	writePlan(t, &req, filepath.Dir(req.PrivateDir))
	r, e := Run(context.Background(), req)
	if e != Error("trusted_executable_changed_during_preflight") || r.Started {
		t.Fatal("changed executable launched")
	}
}

func TestEnvironmentAndAuthConfigNotInherited(t *testing.T) {
	requireDarwin(t)
	t.Setenv("OPENAI_API_KEY", "authored-private-environment-value")
	t.Setenv("HF_TOKEN", "authored-private-hf-value")
	t.Setenv("MY_PRIVATE_CONFIG", "authored-private-config-value")
	body := `test -z "$OPENAI_API_KEY" && test -z "$HF_TOKEN" && test -z "$MY_PRIVATE_CONFIG" || exit 8
test ! -e "$CODEX_HOME/config.toml" || exit 9
test ! -e "$HOME/AGENTS.md" || exit 10
test -f "$CODEX_HOME/auth.json" || exit 11
test "$HOME" != "$CODEX_HOME" || exit 12
` + completeTrace
	req := fixtureRequest(t, body)
	auth := filepath.Join(filepath.Dir(req.PrivateDir), "auth-source")
	os.Mkdir(auth, 0700)
	// Literal authored credential-shaped fixtures; no real secret is present.
	data := `{"auth_mode":"chatgpt","OPENAI_API_KEY":null,"tokens":{"access_token":"fixture-access","refresh_token":"fixture-refresh","id_token":"fixture-id","account_id":"fixture-account"},"last_refresh":"2026-01-01T00:00:00Z"}`
	os.WriteFile(filepath.Join(auth, "auth.json"), []byte(data), 0600)
	os.WriteFile(filepath.Join(auth, "config.toml"), []byte("private fixture config"), 0600)
	req.AuthSourceDir = auth
	r, e := Run(context.Background(), req)
	if e != nil {
		t.Fatal(e)
	}
	if r.ProcessStatus != "exited_zero" {
		t.Fatal("controlled environment fixture failed")
	}
	if _, e = os.Stat(filepath.Join(req.PrivateDir, "codex-home", "auth.json")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("copied auth retained")
	}
	b, _ := json.Marshal(r)
	for _, v := range []string{"fixture-access", "fixture-refresh", "fixture-id", "authored-private-environment-value", "private fixture config"} {
		if bytes.Contains(b, []byte(v)) {
			t.Fatal("auth/environment escaped")
		}
	}
}

func TestPrelaunchErrorsRemoveCopiedAuth(t *testing.T) {
	requireDarwin(t)
	req := fixtureRequest(t, completeTrace)
	auth := filepath.Join(filepath.Dir(req.PrivateDir), "auth-source")
	os.Mkdir(auth, 0700)
	os.WriteFile(filepath.Join(auth, "auth.json"), []byte(`{"auth_mode":"chatgpt","tokens":{"access_token":"fixture-access","refresh_token":"fixture-refresh","id_token":"fixture-id"}}`), 0600)
	req.AuthSourceDir = auth
	// Keep plan consistent with expected hash while the actual executable differs.
	req.ExpectedCLIHash = strings.Repeat("0", 64)
	writePlan(t, &req, filepath.Dir(req.PrivateDir))
	_, e := Run(context.Background(), req)
	if e != Error("trusted_executable_pin_mismatch") {
		t.Fatal("wrong prelaunch error")
	}
	if _, e = os.Stat(filepath.Join(req.PrivateDir, "codex-home", "auth.json")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("auth not removed on prelaunch refusal")
	}
}

func TestAuthModesAndSymlinksRefused(t *testing.T) {
	for _, auth := range []string{`{"auth_mode":"apikey","OPENAI_API_KEY":"fixture-value"}`, `{"auth_mode":"chatgpt","tokens":null}`, `{"auth_mode":"chatgpt","OPENAI_API_KEY":"fixture-value","tokens":{"access_token":"a","refresh_token":"b","id_token":"c"}}`} {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "auth.json"), []byte(auth), 0600)
		e := stageAuth(dir, t.TempDir())
		if e != Error("chatgpt_auth_required") {
			t.Fatal("unsupported auth admitted")
		}
	}
	dir := t.TempDir()
	target := filepath.Join(t.TempDir(), "auth.json")
	os.WriteFile(target, []byte("fixture"), 0600)
	os.Symlink(target, filepath.Join(dir, "auth.json"))
	if stageAuth(dir, t.TempDir()) != Error("auth_source_unavailable") {
		t.Fatal("symlink auth admitted")
	}
}

func TestCandidateSymlinksAndOutsideFiles(t *testing.T) {
	base := []taskverify.BaseFile{{Path: "source.go", Data: []byte("public")}}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "source.go"), []byte("public"), 0600)
	os.WriteFile(filepath.Join(dir, "unassessed-secret-name"), bytes.Repeat([]byte("x"), 2<<20), 0600)
	files, e := snapshot(dir, t.TempDir(), base)
	if e != nil || len(files) != 1 || files[0].Path != "source.go" {
		t.Fatal("outside scope affected snapshot")
	}
	os.Remove(filepath.Join(dir, "source.go"))
	os.Symlink(filepath.Join(dir, "unassessed-secret-name"), filepath.Join(dir, "source.go"))
	if _, e = snapshot(dir, t.TempDir(), base); e == nil {
		t.Fatal("symlink candidate admitted")
	}
}

func TestCaptureBoundHashesFullObservedBytes(t *testing.T) {
	canceled := false
	w, e := newCapture(filepath.Join(t.TempDir(), "trace"), 4, func() { canceled = true })
	if e != nil {
		t.Fatal(e)
	}
	defer w.file.Close()
	w.Write([]byte("abc"))
	w.Write([]byte("def"))
	r := w.report()
	if !canceled || r.Complete || r.Bytes != 6 || r.RetainedBytes != 4 || r.SHA256 != digest([]byte("abcdef")) || r.Status != "byte_limit" {
		t.Fatal("bounded hash accounting")
	}
	w.file.Close()
	data, _ := os.ReadFile(w.file.Name())
	if string(data) != "abcd" {
		t.Fatal("unbounded retained capture")
	}
}

func TestCanonicalArgsProfileAndHashes(t *testing.T) {
	profile, canonical, e := permissionsConfig(runtime.GOROOT())
	if e != nil {
		t.Fatal(e)
	}
	args := invocationArgs("fixture-model", "low", "$WORKSPACE", canonical, "$GOROOT")
	joined := strings.Join(args, "\n")
	for _, flag := range []string{"--no-daemon", "--ignore-user-config", "--ignore-rules", "--ephemeral", "--json", "--skip-git-repo-check", "--disable", "apps", "hooks", "multi_agent", "memories", "skill_mcp_dependency_install", `web_search="disabled"`, `approval_policy="never"`, `default_permissions="riido-task"`, `permissions.riido-task.network.enabled=false`, `shell_environment_policy.inherit="none"`, `shell_environment_policy.set=`, `"GOTMPDIR"=`, `"GOCACHE"=`, `"GOPROXY"="off"`, `:root`, `:tmpdir`, `:slash_tmp`} {
		if !strings.Contains(joined, flag) {
			t.Fatal("missing required isolation argument")
		}
	}
	if strings.Contains(joined, "--sandbox") || strings.Contains(joined, "--resume") || strings.Contains(joined, "API_KEY") || strings.Contains(joined, runtime.GOROOT()) || !strings.Contains(profile, strconvQuoteRoot()) {
		t.Fatal("unsafe or path-dependent arguments")
	}
	if digestJSON(args) != digestJSON(invocationArgs("fixture-model", "low", "$WORKSPACE", canonical, "$GOROOT")) {
		t.Fatal("canonical hash drift")
	}
}

func TestTimeoutTerminatesOwnedDescendantBeforeSnapshot(t *testing.T) {
	requireDarwin(t)
	body := `(/bin/sleep 30; printf changed > pkg/reporouter/router.go) &
printf '%s' "$!" > .authored-descendant-pid
/bin/sleep 30
`
	req := fixtureRequest(t, body)
	req.Timeout = time.Second
	writePlan(t, &req, filepath.Dir(req.PrivateDir))
	r, e := Run(context.Background(), req)
	if e != nil {
		t.Fatal(e)
	}
	if r.ProcessStatus != "timeout_or_canceled" || r.GroupCleanup != "group_terminated" || r.WholeAttemptUsageComplete || r.VerificationStatus != "rejected" || r.WallMillis < 900 || r.WallMillis > 5000 {
		t.Fatalf("bad timeout/cleanup: %+v", r)
	}
	// An unchanged closure is independently observed only after group absence.
	if r.CandidateSHA256 != r.BaseSHA256 {
		t.Fatal("timed-out descendant changed frozen candidate")
	}
}

func TestInheritedPipeTimeoutInvalidatesCapture(t *testing.T) {
	requireDarwin(t)
	req := fixtureRequest(t, completeTrace+"/bin/sleep 30 &\nexit 0\n")
	req.Timeout = 4 * time.Second
	writePlan(t, &req, filepath.Dir(req.PrivateDir))
	r, e := Run(context.Background(), req)
	if e != nil {
		t.Fatal(e)
	}
	if r.ProcessStatus != "inherited_pipe_timeout" || r.GroupCleanup != "group_terminated" || r.Stdout.Complete || r.Stderr.Complete || r.WholeAttemptUsageComplete || r.SummaryStatus != "capture_incomplete" {
		t.Fatalf("pipe-close treated as natural EOF: %+v", r)
	}
}

func TestStderrBoundStillReturnsAttemptRecord(t *testing.T) {
	requireDarwin(t)
	req := fixtureRequest(t, `/bin/dd if=/dev/zero bs=65536 count=20 1>&2 2>/dev/null`)
	r, e := Run(context.Background(), req)
	if e != nil {
		t.Fatal(e)
	}
	if !r.Started || r.ProcessStatus != "capture_limit" || r.Stderr.Complete || r.Stderr.RetainedBytes != MaxStderrBytes || r.WholeAttemptUsageComplete {
		t.Fatalf("wrong bounded attempt: %+v", r)
	}
	if _, e = os.Stat(filepath.Join(req.PrivateDir, "record.json")); e != nil {
		t.Fatal("bounded attempt record missing")
	}
}

func TestIsolatedPrivatePathRequiresTempAndNoProjectAncestors(t *testing.T) {
	root := t.TempDir()
	if _, e := isolatedPrivatePath(filepath.Join(root, "new-run")); e != nil {
		t.Fatal(e)
	}
	for _, marker := range []string{".git", "AGENTS.md", ".codex", ".agents"} {
		if e := os.WriteFile(filepath.Join(root, marker), []byte("authored marker"), 0600); e != nil {
			t.Fatal(e)
		}
		if _, e := isolatedPrivatePath(filepath.Join(root, "new-run")); e != Error("private_project_ancestor_forbidden") {
			t.Fatal("project ancestor accepted")
		}
		os.Remove(filepath.Join(root, marker))
	}
	if _, e := isolatedPrivatePath("/usr/local/new-run"); e == nil {
		t.Fatal("non-temp path accepted")
	}
}

func TestPrivateRecordNeverOverwritesExistingAttempt(t *testing.T) {
	dir := t.TempDir()
	r := Record{Schema: Schema, Started: true, ProcessStatus: "running"}
	if e := writeRecord(dir, "started-record.json", r); e != nil {
		t.Fatal(e)
	}
	if e := writeRecord(dir, "started-record.json", Record{}); e != Error("record_write_failed") {
		t.Fatal("started attempt overwritten")
	}
	b, e := os.ReadFile(filepath.Join(dir, "started-record.json"))
	if e != nil || !bytes.Contains(b, []byte(`"started":true`)) {
		t.Fatal("durable initial record lost")
	}
}

func strconvQuoteRoot() string { return `"` + runtime.GOROOT() + `"` }

func TestUnsupportedOSNeverLaunches(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("other OS refusal")
	}
	req := fixtureRequest(t, completeTrace)
	r, e := Run(context.Background(), req)
	if e != Error("execution_isolation_unavailable") || r.Started {
		t.Fatal("unsupported OS launched")
	}
}

func TestTrimpathPackagedExecutorExplicitGoRoot(t *testing.T) {
	requireDarwin(t)
	body := `/usr/bin/sed -i '' 's#// Snapshot for one caller-defined budget period. This is not a reservation ledger.#// Snapshot supplied by the caller for a single budget period; no reservation is made.#' pkg/catalog/catalog.go
` + completeTrace
	req := fixtureRequest(t, body)
	req.TaskID = "comment-budget-period"
	spec, _ := taskverify.TaskSpec(req.TaskID)
	req.ExpectedSpecSHA256 = digestJSON(spec)
	writePlan(t, &req, filepath.Dir(req.PrivateDir))
	root, e := filepath.EvalSymlinks(runtime.GOROOT())
	if e != nil {
		t.Fatal(e)
	}
	packaged := filepath.Join(filepath.Dir(req.PrivateDir), "riido-taskrun-trimpath")
	repo, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	buildCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	build := exec.CommandContext(buildCtx, filepath.Join(root, "bin", "go"), "build", "-trimpath", "-o", packaged, "./cmd/riido-taskrun")
	build.Dir = repo
	build.Env = append(os.Environ(), "GOROOT="+root, "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off")
	if out, e := build.CombinedOutput(); e != nil {
		t.Fatalf("authored packaged fixture build failed: %v (%d diagnostic bytes)", e, len(out))
	}
	args := []string{"--execute", "--task", req.TaskID, "--model", req.Model, "--reasoning", req.Reasoning, "--base-dir", req.BaseDir, "--private-dir", req.PrivateDir, "--codex-bin", req.CodexBinary, "--codex-sha256", req.ExpectedCLIHash, "--codex-version", req.ExpectedCLIVersion, "--plan-file", req.PlanFile, "--plan-sha256", req.PlanSHA256, "--attempt-ordinal", "1", "--task-spec-sha256", req.ExpectedSpecSHA256, "--timeout", "2s"}
	// No inherited GOROOT or PATH-discovered Go toolchain is available to the
	// packaged process. The authored CLI fixture performs no model inference.
	env := []string{"PATH=/usr/bin:/bin:/usr/sbin:/sbin", "HOME=" + filepath.Dir(req.PrivateDir), "TMPDIR=" + os.TempDir(), "LANG=en_US.UTF-8"}
	missing := exec.CommandContext(buildCtx, packaged, args...)
	missing.Env = env
	var missingOut, missingErr bytes.Buffer
	missing.Stdout = &missingOut
	missing.Stderr = &missingErr
	if e := missing.Run(); e == nil || missingOut.Len() != 0 || strings.TrimSpace(missingErr.String()) != "trusted_toolchain_unavailable" {
		t.Fatal("trimpath missing explicit toolchain was not a clean prelaunch refusal")
	}
	cmd := exec.CommandContext(buildCtx, packaged, append(args, "--go-root", root)...)
	cmd.Env = env
	var out, diagnostics bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &diagnostics
	if e := cmd.Run(); e != nil {
		t.Fatalf("authored packaged execution failed: %v (%d redacted diagnostic bytes)", e, diagnostics.Len())
	}
	var r Record
	if json.Unmarshal(out.Bytes(), &r) != nil || !r.Started || r.VerificationStatus != "accepted" || !r.Verification.ExecutionIsolated || r.Verification.IndependentTests == 0 || !r.WholeAttemptUsageComplete || r.GoVersion != "go version "+PinnedGoVersion+" "+runtime.GOOS+"/"+runtime.GOARCH || len(r.GoBinarySHA256) != 64 {
		t.Fatal("packaged authored fixture did not prove behavioral acceptance and toolchain binding")
	}
	if bytes.Contains(out.Bytes(), []byte(root)) || bytes.Contains(out.Bytes(), []byte(req.PrivateDir)) {
		t.Fatal("packaged report leaked local toolchain or private paths")
	}
}

func TestExplicitToolchainValidationRefusesMissingOrNonExecutable(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "bin"), 0700)
	os.WriteFile(filepath.Join(root, "bin", "go"), []byte("authored nonexecutable fixture"), 0600)
	for _, supplied := range []string{"relative-go-root", filepath.Join(root, "missing"), root} {
		if _, e := resolveGoRoot(context.Background(), supplied); e != Error("trusted_toolchain_unavailable") {
			t.Fatal("invalid toolchain accepted")
		}
	}
}

func TestSuppliedPlanToolchainPinsAreVerifiedBeforeAttempt(t *testing.T) {
	requireDarwin(t)
	toolchain, e := resolveGoRoot(context.Background(), runtime.GOROOT())
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct{ name, version, hash string }{{"wrong_hash", toolchain.version, strings.Repeat("0", 64)}, {"wrong_version", "go version go0.0.0 fixture/fixture", toolchain.hash}, {"missing_version", "", toolchain.hash}} {
		t.Run(tc.name, func(t *testing.T) {
			req := fixtureRequest(t, completeTrace)
			b, e := os.ReadFile(req.PlanFile)
			if e != nil {
				t.Fatal(e)
			}
			var p plan
			if json.Unmarshal(b, &p) != nil {
				t.Fatal("fixture plan")
			}
			pins, _ := json.Marshal(struct {
				Version string `json:"version"`
				Hash    string `json:"binary_sha256"`
			}{tc.version, tc.hash})
			p.GoToolchain = pins
			b, _ = json.Marshal(p)
			os.WriteFile(req.PlanFile, b, 0600)
			req.PlanSHA256 = digest(b)
			r, e := Run(context.Background(), req)
			if e != Error("planned_go_toolchain_mismatch") || r.Started {
				t.Fatal("mismatched planned toolchain launched")
			}
		})
	}
}
