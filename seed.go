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
	"crypto/rand"
	"errors"
	"net"
	"os"
	"os/user"
	"time"

	"github.com/seehuhn/trace"
)

var (
	errReadFailed = errors.New("read error")
)

// SetInitialSeed sets an initial seed for the Accumulator.  An
// attempt is made to obtain seeds which differ between machines and
// between reboots.  To achieve this, the following information is
// incorporated into the seed: the current time of day, account
// information for the current user, and information about the
// installed network interfaces.  In addition, if available, random
// bytes from the random number generator in the crypto/rand package
// are used.
func (acc *Accumulator) SetInitialSeed() {
	acc.genMutex.Lock()
	defer acc.genMutex.Unlock()
	gen := acc.gen

	// source 1: system random number generator
	buffer := make([]byte, len(gen.key))
	n, _ := rand.Read(buffer)
	if n > 0 {
		trace.T("fortuna/seed", trace.PrioInfo,
			"using crypto/rand for seed data")
		gen.Reseed(buffer)
	}

	// source 2: current time of day
	now := time.Now()
	trace.T("fortuna/seed", trace.PrioInfo,
		"using the current time for seed data")
	gen.Reseed([]byte(now.String()))

	// source 3: user name and login details
	user, _ := user.Current()
	if user != nil {
		trace.T("fortuna/seed", trace.PrioInfo,
			"using information about the current user for seed data")
		gen.Reseed([]byte(user.Uid))
		gen.Reseed([]byte(user.Gid))
		gen.Reseed([]byte(user.Username))
		gen.Reseed([]byte(user.Name))
		gen.Reseed([]byte(user.HomeDir))
	}

	// source 4: network interfaces
	ifaces, _ := net.Interfaces()
	if ifaces != nil {
		trace.T("fortuna/seed", trace.PrioInfo,
			"using network interface information for seed data")
		for _, iface := range ifaces {
			gen.ReseedInt64(int64(iface.MTU))
			gen.Reseed([]byte(iface.Name))
			gen.Reseed(iface.HardwareAddr)
			gen.ReseedInt64(int64(iface.Flags))
		}
	}
}

func writeSeed(f *os.File, seed []byte) error {
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
		"new seed data written to %q", f.Name())
	return nil
}

// Read and update an existing Fortuna seed file.
//
// If reading the seed file fails, for example because the seed file
// does not exist or is corrupted, errReadFailed is returned.  In this
// case, the program can continue to run, and after some entropy has
// accumulated the WriteSeedFile() method should be called to create a
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
	return writeSeed(f, seed)
}

// WriteSeedFile writes a new Fortuna seed file with name 'fileName'.
// The purpose of this seed file is to provide a source of randomness
// for the initial phase of the next run of the same program.  The
// contents of the seed file must be kept confidential.
//
// The function WriteSeedFile should be called on program shutdown and
// also periodically while the program runs.  Ferguson and Schneier
// (Practical Cryptography, Wiley, 2003) recommend to write a new seed
// file "every 10 minutes or so".
//
// If the seed file cannot be written, a non-nil error is returned.
// In this case, an error message should be shown and the random
// number generator should not be used until the problem is resolved.
func (acc *Accumulator) WriteSeedFile(fileName string) error {
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
	return writeSeed(f, seed)
}
