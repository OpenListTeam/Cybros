//go:build darwin

package console

import "golang.org/x/sys/unix"

const (
	terminalGetTermiosRequest = unix.TIOCGETA
	terminalSetTermiosRequest = unix.TIOCSETA
)
