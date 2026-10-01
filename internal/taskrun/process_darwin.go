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
	return cleanupGroup(cmd.Process.Pid, syscall.Kill, time.Now, time.Sleep)
}

// A Darwin group can briefly contain only unreaped zombies. Its group lookup
// succeeds, but killpg can return EPERM when no signalable member is found.
// EPERM proves neither absence nor successful killing. Keep the existing bounded
// wait, accepting termination only after an explicit ESRCH observation.
func cleanupGroup(pid int, kill func(int, syscall.Signal) error, now func() time.Time, sleep func(time.Duration)) string {
	if pid <= 1 {
		return "cleanup_failed"
	}
	e := kill(-pid, syscall.SIGKILL)
	if e == syscall.ESRCH {
		return "group_terminated"
	}
	if e != nil && e != syscall.EPERM {
		return "cleanup_failed"
	}
	// Kill delivery is not a proof of termination. Wait until the owned process
	// group disappears before snapshotting any candidate bytes.
	deadline := now().Add(2 * time.Second)
	for now().Before(deadline) {
		e = kill(-pid, 0)
		if e == syscall.ESRCH {
			return "group_terminated"
		}
		if e != nil && e != syscall.EPERM {
			return "cleanup_failed"
		}
		sleep(10 * time.Millisecond)
	}
	return "cleanup_unknown"
}
