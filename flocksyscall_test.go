//go:build darwin || freebsd || linux || netbsd || openbsd
// +build darwin freebsd linux netbsd openbsd

package fortuna

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFlock(t *testing.T) {
	tempDir := os.TempDir()
	defer os.Remove(tempDir)
	testFileName := filepath.Join(tempDir, "test")

	file1, err := os.Create(testFileName)
	if err != nil {
		t.Error(err)
	}
	defer file1.Close()
	_, err = file1.Write([]byte("test"))
	if err != nil {
		t.Error(err)
	}

	file2, err := os.Open(testFileName)
	if err != nil {
		t.Error(err)
	}
	defer file2.Close()

	err = flock(file1)
	if err != nil {
		t.Error("flock failed")
	}
	err = flock(file2)
	if err != errAlreadyLocked {
		t.Error("flock wrongly succeeded")
	}
	err = funlock(file1)
	if err != nil {
		t.Error("funlock failed")
	}

	err = flock(file2)
	if err != nil {
		t.Error("flock failed")
	}
	err = flock(file1)
	if err != errAlreadyLocked {
		t.Error("flock wrongly succeeded")
	}
	err = funlock(file2)
	if err != nil {
		t.Error("funlock failed")
	}
}

func TestFlockClose(t *testing.T) {
	tempDir := os.TempDir()
	defer os.Remove(tempDir)
	testFileName := filepath.Join(tempDir, "test")

	file1, err := os.Create(testFileName)
	if err != nil {
		t.Error(err)
	}
	_, err = file1.Write([]byte("test"))
	if err != nil {
		t.Error(err)
	}

	file2, err := os.Open(testFileName)
	if err != nil {
		t.Error(err)
	}
	defer file2.Close()

	// Verify that the file lock is release on close.
	err = flock(file1)
	if err != nil {
		t.Error(err)
	}
	file1.Close()
	err = flock(file2)
	if err != nil {
		t.Error("flock failed")
	}
	err = funlock(file2)
	if err != nil {
		t.Error("funlock failed")
	}
}

func TestSeedFileSharing(t *testing.T) {
	tempDir := os.TempDir()
	defer os.Remove(tempDir)
	seedFileName := filepath.Join(tempDir, "seed")

	rng1, err := NewRNG(seedFileName)
	if err != nil {
		t.Error(err)
	}
	defer rng1.Close()

	rng2, err := NewRNG(seedFileName)
	if err == nil {
		rng2.Close()
		t.Error("shared seed file not detected")
	}
}
