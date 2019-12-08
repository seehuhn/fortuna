// +build darwin freebsd linux netbsd openbsd

package fortuna

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (s *fortunaSuite) TestFlock(c *C) {
	tempDir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)

	defer os.RemoveAll(tempDir)
	testFileName := filepath.Join(tempDir, "test")

	file1, err := os.Create(testFileName)
	c.Assert(err, IsNil)

	defer file1.Close()
	_, err = file1.Write([]byte("test"))
	c.Assert(err, IsNil)

	file2, err := os.Open(testFileName)
	c.Assert(err, IsNil)

	defer file2.Close()

	err = flock(file1)
	c.Assert(err, IsNil)

	err = flock(file2)
	c.Assert(err, Not(ErrorMatches), errAlreadyLocked)

	err = funlock(file1)
	c.Assert(err, IsNil)

	err = flock(file2)
	c.Assert(err, IsNil)

	err = flock(file1)
	c.Assert(err, Not(ErrorMatches), errAlreadyLocked)

	err = funlock(file2)
	c.Assert(err, IsNil)
}

func (s *fortunaSuite) TestFlockClose(c *C) {
	tempDir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)

	defer os.RemoveAll(tempDir)
	testFileName := filepath.Join(tempDir, "test")

	file1, err := os.Create(testFileName)
	c.Assert(err, IsNil)

	_, err = file1.Write([]byte("test"))
	c.Assert(err, IsNil)

	file2, err := os.Open(testFileName)
	c.Assert(err, IsNil)

	defer file2.Close()

	// Verify that the file lock is release on close.
	err = flock(file1)
	c.Assert(err, IsNil)

	file1.Close()
	err = flock(file2)
	c.Assert(err, IsNil)

	err = funlock(file2)
	c.Assert(err, IsNil)
}

func (s *fortunaSuite) TestSeedFileSharing(c *C) {
	tempDir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)

	defer os.RemoveAll(tempDir)
	seedFileName := filepath.Join(tempDir, "seed")

	rng1, err := NewRNG(seedFileName)
	c.Assert(err, IsNil)

	defer rng1.Close()

	rng2, err := NewRNG(seedFileName)
	c.Assert(err, NotNil)

	if err == nil {
		defer rng2.Close()
	}
}
