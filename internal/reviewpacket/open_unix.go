//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package reviewpacket

import (
	"os"
	"syscall"
)

func openInput(path string) (inputFile, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
}
