//go:build !darwin && !linux && !freebsd && !openbsd && !netbsd && !dragonfly

package reviewpacket

import "errors"

func openInput(string) (inputFile, error) {
	return nil, errors.New("platform lacks supported nonblocking no-follow input open")
}
