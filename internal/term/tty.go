// Package term provides terminal utilities for TTY detection.
package term

import (
	"os"

	"golang.org/x/sys/unix"
)

// IsTTY returns whether stdout is a TTY.
func IsTTY() bool {
	fd := int(os.Stdout.Fd())
	_, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	return err == nil
}

