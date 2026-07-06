//go:build linux

package console

import "golang.org/x/sys/unix"

const (
	terminalGetTermiosRequest = unix.TCGETS
	terminalSetTermiosRequest = unix.TCSETS
)
