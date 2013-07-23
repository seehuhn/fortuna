// accumulator.go - an entropy accumulator for Fortuna
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
	"hash"
	"sync"
	"time"

	"github.com/seehuhn/sha256d"
)

const (
	minPoolSize           = 48
	seedFileWriteInterval = 10 * time.Minute
)

// Accumulator holds the state of one instance of the Fortuna random
// number generator.  Randomness can be extracted using the
// RandomData() method, entropy from the environment should be
// submitted regularly using the AddRandomEvent() method.  It is safe
// to access an Accumulator object concurrently from different
// go-routines.
type Accumulator struct {
	genMutex sync.Mutex
	gen      *Generator

	poolMutex    sync.Mutex
	reseedCount  int
	lastReseed   time.Time
	pool         [32]hash.Hash
	poolZeroSize int
}

// NewAccumulator creates a new instance of the Fortuna random number
// generator.  The function newCipher should normally be aes.NewCipher
// from the crypto/aes package, but the Serpent or Twofish ciphers can
// also be used.  The argument seedFileName gives the name of a file
// where a small amount of randomness can be stored between runs of
// the program; the program must be able to both read and write this
// file.
//
// In case the seed file does not exist or is corrupted, a new seed
// file is created.  If the seed file cannot be written, and error is
// returned.  NewAccumulator() starts a background go-routine which
// updates the seed file every 10 minutes during the run of the
// program.
func NewAccumulator(newCipher NewCipher,
	seedFileName string) (*Accumulator, error) {
	acc := &Accumulator{
		gen: NewGenerator(newCipher),
	}
	for i := 0; i < len(acc.pool); i++ {
		acc.pool[i] = sha256d.New()
	}

	if seedFileName != "" {
		// We use SetInitialSeed() to protect against missing seed
		// files.  Since the data from the seed file is mixed into the
		// data from SetInitialSeed(), and since the latter depends on
		// the current time, this also protects against old seed files
		// being restored from backups, etc.
		acc.SetInitialSeed()

		err := acc.updateSeedFile(seedFileName)
		if err == errReadFailed {
			err = acc.WriteSeedFile(seedFileName)
		}
		if err != nil {
			return nil, err
		}
		go func() {
			tick := time.Tick(seedFileWriteInterval)
			for _ = range tick {
				acc.WriteSeedFile(seedFileName)
			}
		}()
	}

	return acc, nil
}

func (acc *Accumulator) tryReseeding() []byte {
	acc.poolMutex.Lock()
	defer acc.poolMutex.Unlock()

	now := time.Now()
	if acc.poolZeroSize >= minPoolSize &&
		now.Sub(acc.lastReseed) > 100*time.Millisecond {
		acc.lastReseed = now
		acc.poolZeroSize = 0
		acc.reseedCount += 1

		seed := []byte{}
		for i := uint(0); i < 32; i++ {
			x := 1 << i
			if acc.reseedCount%x == 0 {
				seed = acc.pool[i].Sum(seed)
				acc.pool[i].Reset()
			} else {
				break
			}
		}
		return seed
	}
	return nil
}

// RandomData returns a slice of n random bytes.  The result can be
// used as a replacement for a sequence of uniformly distributed and
// independent bytes, and will be difficult to guess for an attacker.
func (acc *Accumulator) RandomData(n uint) []byte {
	seed := acc.tryReseeding()
	acc.genMutex.Lock()
	defer acc.genMutex.Unlock()
	if seed != nil {
		acc.gen.Reseed(seed)
	}
	return acc.gen.PseudoRandomData(n)
}

func (acc *Accumulator) randomDataUnlocked(n uint) []byte {
	seed := acc.tryReseeding()
	if seed != nil {
		acc.gen.Reseed(seed)
	}
	return acc.gen.PseudoRandomData(n)
}

// AddRandomEvent should be called periodically to add entropy to the
// state of the random number generator.  The random data provided
// should be derived from quantities which change between calls and
// which cannnot be (completely) known by an attacker.  Typical
// sources of randomness include the times between the arrival of
// network packets, the time between key-presses by the user, noise
// from an external microphone, etc.
//
// Different sources of randomness should use different values for the
// 'source' argument.  There are 32 internal pools for storing of
// randomness, numbered 0, 1, ..., 31; the pool the randomness from
// the current call is destined for is given by the 'pool' argument.
// Callers must distribute the randomness from each source uniformly
// over the pools in a round-robin fashion.  Finally, the argument
// 'data' gives the randomness to add to the pool.  'data' should be
// at most 32 bytes long; longer values should be hashed by the caller
// and the hash be submitted instead.
func (acc *Accumulator) AddRandomEvent(source uint8, pool uint8, data []byte) {
	acc.poolMutex.Lock()
	defer acc.poolMutex.Unlock()

	poolHash := acc.pool[pool]
	poolHash.Write([]byte{source, byte(len(data))})
	poolHash.Write(data)
	if pool == 0 {
		acc.poolZeroSize += 2 + len(data)
	}
}
