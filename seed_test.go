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
	"os"
	"path/filepath"
	"testing"
)

func TestSeedfile(t *testing.T) {
	tempDir := os.TempDir()
	defer os.Remove(tempDir)
	seedFileName := filepath.Join(tempDir, "seed")

	// check that the seed file is created
	rng, err := NewRNG(seedFileName)
	if err != nil {
		t.Fatal(err)
	}
	err = rng.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(seedFileName); os.IsNotExist(err) {
		t.Error("seed file not found")
	}

	// check that .updateSeedFile() sets the seed and updates the file
	rng, err = NewRNG(seedFileName)
	if err != nil {
		t.Fatal(err)
	}
	rng.gen.reset()
	before, err := os.ReadFile(seedFileName)
	if err != nil {
		t.Error(err)
	}
	err = rng.updateSeedFile()
	if err != nil {
		t.Error(err)
	}
	after, err := os.ReadFile(seedFileName)
	if err != nil {
		t.Error(err)
	}
	// the following would panic if the seed is not reset
	rng.RandomData(1)
	err = rng.Close()
	if err != nil {
		t.Error(err)
	}
	if len(before) != seedFileSize || bytes.Equal(before, after) {
		t.Error("seed file not correctly updated")
	}

	// check that insecure seed files are detected
	err = os.Chmod(seedFileName, os.FileMode(0644))
	if err != nil {
		t.Fatal(err)
	}
	rng, err = NewRNG(seedFileName)
	if err != ErrInsecureSeed {
		t.Error("insecure seed file not detected")
	}
	if rng != nil {
		rng.Close()
	}
	err = os.Chmod(seedFileName, os.FileMode(0600))
	if err != nil {
		t.Fatal(err)
	}

	// check that seed files of wrong length are detected
	err = os.WriteFile(seedFileName, []byte("Hello"), os.FileMode(0600))
	if err != nil {
		t.Error(err)
	}
	rng, err = NewRNG(seedFileName)
	if err != ErrCorruptedSeed {
		t.Error("corrupted seed file not detected:", err)
	}
	if rng != nil {
		rng.Close()
	}
}
