// seed.go - seed file handling for the Fortuna generator
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
	"errors"
	"io"
	"os"
)

const (
	seedFileSize = 64
)

var (
	ErrCorruptedSeed = errors.New("seed file corrupted")
	ErrInsecureSeed  = errors.New("seed file with insecure permissions")
)

func doWriteSeed(f *os.File, seed []byte) error {
	_, err := f.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	n, err := f.Write(seed)
	if err != nil || n != len(seed) {
		if err == nil {
			err = &os.PathError{Op: "write", Path: f.Name(), Err: nil}
		}
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

// Read and update the seed file.
//
// If the seed file is empty, reading the seed file is omitted.  After
// (potentially) reading the contents of the seed file, new seed data
// is written to the file.  In case the seed file is corrupted or has
// insecure file permissions, an error is returned.
func (acc *Accumulator) updateSeedFile() error {
	fi, err := acc.seedFile.Stat()
	if err != nil {
		return err
	} else if fi.Mode()&os.FileMode(0077) != 0 {
		return ErrInsecureSeed
	}

	_, err = acc.seedFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	// To prevent attacks we keep the PRNG locked until the new seed
	// file is safely written to disk.
	acc.genMutex.Lock()
	defer acc.genMutex.Unlock()

	n := fi.Size()
	if n == seedFileSize {
		seed := make([]byte, seedFileSize)
		_, err := io.ReadFull(acc.seedFile, seed)
		if err != nil || isZero(seed) {
			return ErrCorruptedSeed
		}
		acc.gen.Reseed(seed)
	} else if n != 0 {
		return ErrCorruptedSeed
	}

	seed := acc.randomDataUnlocked(seedFileSize)
	return doWriteSeed(acc.seedFile, seed)
}

// writeSeedFile writes 64 bytes of random data to the Fortuna seed
// file.  If the seed file cannot be written, a non-nil error is
// returned.  In this case, the random number generator should not be
// used until the problem is resolved.
func (acc *Accumulator) writeSeedFile() error {
	seed := acc.RandomData(seedFileSize)
	return doWriteSeed(acc.seedFile, seed)
}
