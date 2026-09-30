package taskrun

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const PinnedGoVersion = "go1.27.1"

type trustedToolchain struct{ root, version, hash string }

// A trimpath release can have no compiled-in GOROOT. The caller supplies that
// installation explicitly; we never discover it through an inherited PATH or
// mutate the process-global GOROOT environment.
func resolveGoRoot(ctx context.Context, supplied string) (trustedToolchain, error) {
	if supplied == "" {
		supplied = runtime.GOROOT()
	}
	if supplied == "" || !filepath.IsAbs(supplied) || strings.ContainsAny(supplied, "\x00\r\n") {
		return trustedToolchain{}, Error("trusted_toolchain_unavailable")
	}
	root, e := filepath.EvalSymlinks(supplied)
	if e != nil || !filepath.IsAbs(root) {
		return trustedToolchain{}, Error("trusted_toolchain_unavailable")
	}
	binary := filepath.Join(root, "bin", "go")
	st, e := os.Lstat(binary)
	if e != nil || !st.Mode().IsRegular() || st.Mode()&0111 == 0 {
		return trustedToolchain{}, Error("trusted_toolchain_unavailable")
	}
	hash, e := executableDigest(binary)
	if e != nil {
		return trustedToolchain{}, Error("trusted_toolchain_unavailable")
	}
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, binary, "version")
	cmd.Dir = root
	cmd.Env = []string{"PATH=/usr/bin:/bin:/usr/sbin:/sbin", "GOROOT=" + root, "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "LANG=en_US.UTF-8", "LC_ALL=en_US.UTF-8"}
	configureProcess(cmd)
	cmd.WaitDelay = time.Second
	var output boundedBuffer
	output.limit = 1024
	cmd.Stdout = &output
	cmd.Stderr = &output
	e = cmd.Run()
	cleanup := cleanupProcess(cmd)
	version := strings.TrimSpace(output.buf.String())
	if e != nil || output.exceeded || cleanup != "group_terminated" {
		return trustedToolchain{}, Error("trusted_toolchain_unavailable")
	}
	if version != "go version "+PinnedGoVersion+" "+runtime.GOOS+"/"+runtime.GOARCH {
		return trustedToolchain{}, Error("unsupported_go_version")
	}
	return trustedToolchain{root: root, version: version, hash: hash}, nil
}
