//go:build darwin

package taskrun

import (
	"syscall"
	"testing"
	"time"
)

func TestGroupCleanupRequiresObservedAbsence(t *testing.T) {
	for _, tc := range []struct {
		name      string
		killError error
		probes    []error
		want      string
	}{
		{"already_absent", syscall.ESRCH, nil, "group_terminated"},
		{"reaping_after_kill", nil, []error{syscall.EPERM, syscall.EPERM, syscall.ESRCH}, "group_terminated"},
		{"reaping_before_cleanup", syscall.EPERM, []error{syscall.EPERM, syscall.ESRCH}, "group_terminated"},
		{"live_then_absent", nil, []error{nil, syscall.ESRCH}, "group_terminated"},
		{"permission_never_resolved", syscall.EPERM, []error{syscall.EPERM}, "cleanup_unknown"},
		{"still_live", nil, []error{nil}, "cleanup_unknown"},
		{"unexpected_kill_error", syscall.EINVAL, nil, "cleanup_failed"},
		{"unexpected_probe_error", nil, []error{syscall.EINVAL}, "cleanup_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clock := time.Unix(0, 0)
			calls := 0
			kill := func(pid int, signal syscall.Signal) error {
				if pid != -12345 {
					t.Fatal("wrong owned group")
				}
				calls++
				if calls == 1 {
					if signal != syscall.SIGKILL {
						t.Fatal("missing kill attempt")
					}
					return tc.killError
				}
				if signal != 0 || len(tc.probes) == 0 {
					t.Fatal("invalid absence observation")
				}
				return tc.probes[min(calls-2, len(tc.probes)-1)]
			}
			got := cleanupGroup(12345, kill, func() time.Time { return clock }, func(d time.Duration) { clock = clock.Add(d) })
			if got != tc.want || calls > 201 || clock.Sub(time.Unix(0, 0)) > 2*time.Second {
				t.Fatalf("wrong bounded cleanup: %s calls%d elapsed%s", got, calls, clock.Sub(time.Unix(0, 0)))
			}
		})
	}
}

func TestGroupCleanupNeverSignalsCurrentGroup(t *testing.T) {
	for _, pid := range []int{0, 1, -1} {
		called := false
		result := cleanupGroup(pid, func(int, syscall.Signal) error { called = true; return nil }, time.Now, time.Sleep)
		if called || result != "cleanup_failed" {
			t.Fatal("invalid group signaled")
		}
	}
}
