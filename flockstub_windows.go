//go:build windows
// +build windows

package fortuna

import (
	"os"
)

// TODO find windows machine and create tests

func flock(file *os.File) error {
	return syscall.LockFileEx(int(file.Fd()), 0, 0, ^uint32(0), ^uint32(0), new(syscall.Overlapped))
}

func funlock(file *os.File) error {
	return syscall.UnlockFileEx(int(file.Fd()), 0, 0, ^uint32(0), ^uint32(0), new(syscall.Overlapped))
}
