//go:build darwin

package taskrun

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

func executionSupported() bool {
	st, e := os.Stat("/usr/bin/sandbox-exec")
	return e == nil && st.Mode().IsRegular()
}
func configureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		e := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if e == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return e
	}
}
func cleanupProcess(cmd *exec.Cmd) string {
	if cmd.Process == nil {
		return "unavailable"
	}
	e := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if e == syscall.ESRCH {
		return "group_terminated"
	}
	if e != nil {
		return "cleanup_failed"
	}
	// Kill delivery is not a proof of termination. Wait until the owned process
	// group disappears before snapshotting any candidate bytes.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		e = syscall.Kill(-cmd.Process.Pid, 0)
		if e == syscall.ESRCH {
			return "group_terminated"
		}
		if e != nil {
			return "cleanup_failed"
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "cleanup_unknown"
}
