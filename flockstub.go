// +build !darwin,!freebsd,!linux,!netbsd,!openbsd

package fortuna

import (
	"errors"
	"os"
)

var (
	// ErrAlreadyLocked is an error code indicating that a file could
	// not be locked because a lock was already present.
	errAlreadyLocked = errors.New("resource temporarily unavailable")
)

// Flock is a dummy function which always returns nil on this system.
//
// On systems which support file locking, Flock() tries to acquire an
// exclusive lock to the given file.  If the file is already locked,
// ErrAlreadyLocked is returned.
func flock(file *os.File) error {
	return nil
}

// Funlock always returns nil on this system.
//
// On systems which support file lockgin, Funlock() can be used to
// release a file lock which was previously acquired using Flock()
func funlock(file *os.File) error {
	return nil
}
