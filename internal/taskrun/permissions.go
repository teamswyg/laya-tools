package taskrun

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const profileName = "riido-task"

func permissionsConfig() (string, string, error) {
	toolchain, e := filepath.EvalSymlinks(runtime.GOROOT())
	if e != nil || !filepath.IsAbs(toolchain) {
		return "", "", Error("trusted_toolchain_unavailable")
	}
	st, e := os.Stat(filepath.Join(toolchain, "bin", "go"))
	if e != nil || !st.Mode().IsRegular() {
		return "", "", Error("trusted_toolchain_unavailable")
	}
	// The Go toolchain has one narrowly named read-only root. Neither the user
	// home nor auth/config/cache directories receive a read exception.
	prefix := `permissions.riido-task.filesystem={":root"="deny",":minimal"="read",":workspace_roots"="write",":tmpdir"="deny",":slash_tmp"="deny",`
	profile := prefix + strconv.Quote(toolchain) + `="read"}`
	canonical := prefix + `"$GOROOT"="read"}`
	return profile, canonical, nil
}

func invocationArgs(model, reasoning, workspace, profile string) []string {
	q := strconv.Quote
	root := filepath.Join(workspace, ".riido-runtime")
	commandEnv := `shell_environment_policy.set={"PATH"=` + q(filepath.Join(runtime.GOROOT(), "bin")+":/usr/bin:/bin:/usr/sbin:/sbin") + `,"HOME"=` + q(filepath.Join(root, "home")) + `,"CODEX_HOME"=` + q(filepath.Join(root, "unauthed-codex-home")) + `,"TMPDIR"=` + q(filepath.Join(root, "tmp")) + `,"GOTMPDIR"=` + q(filepath.Join(root, "tmp")) + `,"GOCACHE"=` + q(filepath.Join(root, "gocache")) + `,"GOTOOLCHAIN"="local","GOPROXY"="off","GOSUMDB"="off","CGO_ENABLED"="0","LANG"="en_US.UTF-8","LC_ALL"="en_US.UTF-8"}`
	if workspace == "$WORKSPACE" {
		commandEnv = strings.ReplaceAll(commandEnv, runtime.GOROOT(), "$GOROOT")
	}
	return []string{"--no-daemon", "exec", "--ignore-user-config", "--ignore-rules", "--ephemeral", "--json", "--skip-git-repo-check", "--disable", "apps", "--disable", "hooks", "--disable", "multi_agent", "--disable", "multi_agent_v2", "--disable", "memories", "--disable", "skill_mcp_dependency_install", "--color", "never", "-C", workspace, "-m", model, "-c", "model_reasoning_effort=" + q(reasoning), "-c", `approval_policy="never"`, "-c", `web_search="disabled"`, "-c", `default_permissions="riido-task"`, "-c", profile, "-c", "permissions.riido-task.network.enabled=false", "-c", `shell_environment_policy.inherit="none"`, "-c", commandEnv, "-"}
}

// This preflight is a fixed offline probe, never model inference. A successful
// version string alone is not evidence that the requested profile is enforced.
func permissionProbe(ctx context.Context, binary, workspace, private, home string, env []string, profile string) error {
	sentinel := filepath.Join(private, "outside-sentinel")
	if os.WriteFile(sentinel, []byte("authored-public-isolation-probe"), 0600) != nil {
		return Error("permission_probe_setup_failed")
	}
	defer os.Remove(sentinel)
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		return Error("permission_probe_setup_failed")
	}
	defer listener.Close()
	server := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { io.WriteString(w, "authored-public-network-probe") })}
	defer server.Close()
	go server.Serve(listener)
	url := "http://" + listener.Addr().String()
	client := http.Client{Timeout: time.Second}
	response, e := client.Get(url)
	if e != nil {
		return Error("permission_probe_control_failed")
	}
	response.Body.Close()
	// Each failure is required. Redirections keep both sentinel and any real
	// copied auth bytes out of all probe output, even if isolation is broken.
	script := `printf public > .riido-permission-probe && test -r .riido-permission-probe && /usr/bin/curl --version >/dev/null 2>&1 && "$4" version >/dev/null 2>&1 && ! /bin/cat "$1" >/dev/null 2>&1 && ! /bin/sh -c 'printf denied > "$1"' probe "$1" >/dev/null 2>&1 && ! /usr/bin/curl --silent --max-time 2 "$2" >/dev/null 2>&1 && ! /bin/cat "$3" >/dev/null 2>&1`
	auth := filepath.Join(home, "auth.json")
	// A real sentinel always exists, so an absent login is not mistaken for
	// proof of auth-path denial. Actual copied auth is checked when present.
	if _, e = os.Stat(auth); e != nil {
		auth = sentinel
	}
	args := []string{"--no-daemon", "sandbox", "-C", workspace, "-P", profileName, "-c", profile, "-c", "permissions.riido-task.network.enabled=false", "--", "/bin/sh", "-c", script, "probe", sentinel, url, auth, filepath.Join(runtime.GOROOT(), "bin", "go")}
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, binary, args...)
	cmd.Env = env
	cmd.Dir = workspace
	configureProcess(cmd)
	cmd.WaitDelay = time.Second
	var output boundedBuffer
	output.limit = 4096
	cmd.Stdout = &output
	cmd.Stderr = &output
	e = cmd.Run()
	cleanupProcess(cmd)
	probeFile := filepath.Join(workspace, ".riido-permission-probe")
	contents, readErr := os.ReadFile(probeFile)
	os.Remove(probeFile)
	// The unchanged outside sentinel also catches sandbox write failures that
	// a shell command might otherwise swallow incorrectly.
	outside, readOutsideErr := os.ReadFile(sentinel)
	if e != nil || output.exceeded || readErr != nil || string(contents) != "public" || readOutsideErr != nil || string(outside) != "authored-public-isolation-probe" || strings.TrimSpace(output.buf.String()) != "" {
		return Error("permission_probe_failed")
	}
	return nil
}
