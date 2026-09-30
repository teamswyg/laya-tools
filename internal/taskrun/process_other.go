//go:build !darwin

package taskrun

import "os/exec"

func executionSupported() bool            { return false }
func configureProcess(cmd *exec.Cmd)      {}
func cleanupProcess(cmd *exec.Cmd) string { return "unavailable" }
