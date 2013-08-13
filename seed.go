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
	"os"

	"github.com/seehuhn/trace"
)

var (
	errReadFailed = errors.New("read error")
)

func doWriteSeed(f *os.File, seed []byte) error {
	n, err := f.Write(seed)
	if err != nil || n != len(seed) {
		f.Close()
		if err == nil {
			err = &os.PathError{Op: "write", Path: f.Name(), Err: nil}
		}
		return err
	}

	err = f.Sync()
	if err != nil {
		f.Close()
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	trace.T("fortuna/seed", trace.PrioInfo,
		"writing new seed data to %q", f.Name())
	return nil
}

// Read and update an existing Fortuna seed file.
//
// If reading the seed file fails, for example because the seed file
// does not exist or is corrupted, errReadFailed is returned.  In this
// case, the program can continue to run, and after some entropy has
// accumulated the writeSeedFile() method should be called to create a
// new seed file for future use.
//
// Any other non-nil error indicates that writing the seed file
// failed.  In this case, an error message should be shown and the
// random number generator should not be used until the problem is
// resolved.
func (acc *Accumulator) updateSeedFile(fileName string) error {
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_SYNC, 0600)
	if err != nil {
		return errReadFailed
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return errReadFailed
	} else if fi.Mode()&os.FileMode(0077) != 0 {
		trace.T("fortuna/seed", trace.PrioInfo,
			"seed file %q has insecure permissions, not used", fileName)
		f.Close()
		return errReadFailed
	}

	// Allow Read() to read one excess byte (64 + 1 = 65) to check
	// whether the input file size is correct.
	seed := make([]byte, 65)
	n, err := f.Read(seed)
	if err != nil || n != 64 || isZero(seed[:64]) {
		trace.T("fortuna/seed", trace.PrioInfo,
			"seed file %q is corrupted, not used", fileName)
		f.Close()
		return errReadFailed
	}

	trace.T("fortuna/seed", trace.PrioInfo,
		"reading seed data from %q", fileName)

	acc.genMutex.Lock()
	acc.gen.Reseed(seed[:64])

	// At this point we have successfully read the seed data.  We use
	// this data to update the state of the PRNG, and write a new seed
	// file for the next run of the program.  To prevent potential
	// attacks we need to keep the PRNG locked until the new seed file
	// is safely written to disk.
	defer acc.genMutex.Unlock()

	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		f.Close()
		return err
	}

	seed = acc.randomDataUnlocked(64)
	return doWriteSeed(f, seed)
}

// writeSeedFile writes 64 bytes of random data to a Fortuna seed file
// with name 'fileName'.  The purpose of this seed file is to provide
// a source of randomness for the initial phase of the next run of the
// same program.
//
// If the seed file cannot be written, a non-nil error is returned.
// In this case, the random number generator should not be used until
// the problem is resolved.
func (acc *Accumulator) writeSeedFile(fileName string) error {
	f, err := os.OpenFile(fileName,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC,
		os.FileMode(0600))
	if err != nil {
		return err
	}

	// The file permissions given in the open call only apply if the
	// file does not previously exist; to deal with pre-existing files
	// we also manually set the file mode here.
	err = f.Chmod(os.FileMode(0600))
	if err != nil {
		f.Close()
		return err
	}

	seed := acc.RandomData(64)
	return doWriteSeed(f, seed)
}
