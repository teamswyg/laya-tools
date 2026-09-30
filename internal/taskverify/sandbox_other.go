//go:build !darwin

package taskverify

import (
	"context"
	"fmt"
	"os/exec"
)

func sandboxSupported() bool { return false }

func sandboxCommand(context.Context, string, []string, string, string) (*exec.Cmd, error) {
	return nil, fmt.Errorf("isolation_unavailable")
}
