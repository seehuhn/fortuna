// +build darwin freebsd linux netbsd openbsd

package fortuna

import (
	"os"
	"syscall"
)

const (
	// ErrAlreadyLocked is an error code indicating that a file could
	// not be locked because a lock was already present.
	errAlreadyLocked = syscall.EWOULDBLOCK
)

// Flock tries to acquire an exclusive lock to the given file.  If the
// file is already locked, ErrAlreadyLocked is returned.
//
// The Flock() function is not available on all operating systems.  On
// systems where Flock() is not available, it is replaced with a stub
// function which always returns nil.
func flock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

// Funlock can be used to release a file lock which was previously
// acquired using Flock()
func funlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
