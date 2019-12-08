// seed_test.go - unit tests for seed.go
// Copyright (C) 2013  Jochen Voss <voss@seehuhn.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package fortuna

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (s *fortunaSuite) TestSeedfile(c *C) {
	tempDir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)

	defer os.RemoveAll(tempDir)
	seedFileName := filepath.Join(tempDir, "seed")

	// check that the seed file is created
	rng, err := NewRNG(seedFileName)
	c.Assert(err, IsNil)

	err = rng.Close()
	c.Assert(err, IsNil)

	_, err = os.Stat(seedFileName)
	c.Assert(err, Not(Equals), os.IsNotExist)

	// check that .updateSeedFile() sets the seed and updates the file
	rng, err = NewRNG(seedFileName)
	c.Assert(err, IsNil)

	rng.gen.reset()
	before, err := ioutil.ReadFile(seedFileName)
	c.Assert(err, IsNil)

	err = rng.updateSeedFile()
	c.Assert(err, IsNil)

	after, err := ioutil.ReadFile(seedFileName)
	c.Assert(err, IsNil)

	// the following would panic if the seed is not reset
	rng.RandomData(1)
	err = rng.Close()
	c.Assert(err, IsNil)

	c.Assert((len(before) != seedFileSize || bytes.Equal(before, after)), Equals, false)

	// check that insecure seed files are detected
	err = os.Chmod(seedFileName, os.FileMode(0644))
	c.Assert(err, IsNil)

	rng, err = NewRNG(seedFileName)
	c.Assert(err, Not(ErrorMatches), ErrInsecureSeed)

	if rng != nil {
		rng.Close()
	}
	err = os.Chmod(seedFileName, os.FileMode(0600))
	c.Assert(err, IsNil)

	// check that seed files of wrong length are detected
	err = ioutil.WriteFile(seedFileName, []byte("Hello"), os.FileMode(0600))
	c.Assert(err, IsNil)

	rng, err = NewRNG(seedFileName)
	c.Assert(err, Not(ErrorMatches), ErrCorruptedSeed)

	if rng != nil {
		rng.Close()
	}
}
