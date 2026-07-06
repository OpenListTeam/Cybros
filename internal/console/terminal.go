//go:build darwin || linux

package console

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"cybros/consts"

	"golang.org/x/sys/unix"
)

func readHiddenLine(ctx context.Context, reader *bufio.Reader) (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := unix.IoctlGetTermios(fd, terminalGetTermiosRequest)
	if err != nil {
		return "", fmt.Errorf(consts.ErrorGetTerminalState, err)
	}

	newState := *oldState
	newState.Lflag &^= unix.ECHO
	err = unix.IoctlSetTermios(fd, terminalSetTermiosRequest, &newState)
	if err != nil {
		return "", fmt.Errorf(consts.ErrorDisableTerminalEcho, err)
	}
	defer func() {
		_ = unix.IoctlSetTermios(fd, terminalSetTermiosRequest, oldState)
		fmt.Println()
	}()

	return readLine(ctx, reader)
}
