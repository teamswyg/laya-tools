//go:build darwin

package taskverify

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func sandboxSupported() bool {
	st, err := os.Stat("/usr/bin/sandbox-exec")
	return err == nil && st.Mode().IsRegular()
}

func sandboxCommand(ctx context.Context, goBinary string, args []string, toolchain, scratch string) (*exec.Cmd, error) {
	// Candidate code sees only its public closure, toolchain and OS libraries.
	// It cannot read the checkout, the user's home, credential files or the network.
	q := strconv.Quote
	profile := "(version 1)(deny default)" +
		"(allow process-fork)(allow sysctl-read)(allow mach-lookup)" +
		"(allow process-exec (subpath " + q(toolchain) + ") (subpath " + q(scratch) + "))" +
		"(allow file-read* (subpath " + q(toolchain) + ") (subpath " + q(scratch) + ")" +
		// dyld needs to open the root directory; this grants no file contents
		// beneath it. Without this literal, trusted Go itself aborts at startup.
		" (literal \"/\")" +
		" (subpath \"/System/Library\") (subpath \"/usr/lib\") (subpath \"/private/var/db/dyld\")" +
		" (literal \"/dev/null\") (literal \"/dev/random\") (literal \"/dev/urandom\")" +
		" (literal \"/private/etc/localtime\"))" +
		"(allow file-write* (subpath " + q(scratch) + ") (literal \"/dev/null\"))"
	argv := append([]string{"-p", profile, goBinary}, args...)
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", argv...)
	// Deadline/output cancellation terminates compiler/test descendants too,
	// rather than leaving a child running after only the outer Go command dies.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	return cmd, nil
}
