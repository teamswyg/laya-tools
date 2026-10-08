//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package reviewpacket

import (
	"os"
	"syscall"
	"testing"
)

func TestFIFOAndSymlinkSubstitutionNeverReadOrHang(t *testing.T) {
	for _, kind := range []string{"fifo-existing", "fifo-substitution", "symlink-substitution"} {
		t.Run(kind, func(t *testing.T) {
			cfg := fixtureConfig(t)
			original := cfg.InputPath + ".original"
			if err := os.Rename(cfg.InputPath, original); err != nil {
				t.Fatal(err)
			}
			substitute := func() {
				if kind == "symlink-substitution" {
					if err := os.Symlink(original, cfg.InputPath); err != nil {
						t.Fatal(err)
					}
				} else if err := syscall.Mkfifo(cfg.InputPath, 0600); err != nil {
					t.Fatal(err)
				}
			}
			ops := diskOps()
			if kind == "fifo-existing" {
				substitute()
			} else {
				if err := os.WriteFile(cfg.InputPath, []byte(publicFixture), 0600); err != nil {
					t.Fatal(err)
				}
				ops.open = func(path string) (inputFile, error) {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					substitute()
					return openInput(path)
				}
			}
			payload, outcome, err := capture(cfg, ops)
			code := "input_identity_changed"
			if kind == "fifo-existing" {
				code = "input_type_denied"
			}
			if kind == "symlink-substitution" {
				code = "input_open_failed"
			}
			result := assertFailed(t, payload, outcome, err, cfg, code)
			if result.BytesRead != 0 {
				t.Fatal("denied FIFO or symlink input was read")
			}
		})
	}
}
