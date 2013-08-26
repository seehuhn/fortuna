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
	"crypto/aes"
	"hash"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/seehuhn/sha256d"
	"github.com/seehuhn/trace"
)

const (
	numPools               = 32
	minPoolSize            = 32
	minReseedInterval      = 100 * time.Millisecond
	seedFileUpdateInterval = 10 * time.Minute
)

// Accumulator holds the state of one instance of the Fortuna random
// number generator.  Randomness can be extracted using the
// RandomData() and Read() methods.  Entropy from the environment
// should be submitted regularly using channels allocated by the
// NewEntropyDataSink() or NewEntropyTimeStampSink() methods.
//
// It is safe to access an Accumulator object concurrently from
// different goroutines.
type Accumulator struct {
	seedFileName string
	stopAutoSave chan<- bool

	genMutex sync.Mutex
	gen      *Generator

	poolMutex    sync.Mutex
	reseedCount  int
	nextReseed   time.Time
	pool         [numPools]hash.Hash
	poolZeroSize int

	sourceMutex sync.Mutex
	nextSource  uint8
}

// NewRNG allocates a new instance of the Fortuna random number
// generator.
//
// The argument seedFileName gives the name of a file where a small
// amount of randomness can be stored between runs of the program; the
// program must be able to both read and write this file.  The
// contents of the seed file must be kept confidential and seed files
// must not be shared between concurrently running instances of the
// random number generator.
//
// In case the seed file does not exist or is corrupted, a new seed
// file is created.  If the seed file cannot be written, an error is
// returned.
//
// The returned random generator must be closed using the .Close()
// method after use.
func NewRNG(seedFileName string) (*Accumulator, error) {
	return NewAccumulator(aes.NewCipher, seedFileName)
}

var (
	// NewAccumulatorAES is an alias for NewRNG, provided for backward
	// compatibility.  It should not be used in new code.
	NewAccumulatorAES = NewRNG
)

// NewAccumulator allocates a new instance of the Fortuna random
// number generator.  The argument 'newCipher' allows to choose a
// block cipher like Serpent or Twofish instead of the default AES.
// NewAccumulator(aes.NewCipher, seedFileName) is the same as
// NewRNG(seedFileName).  See the documentation for NewRNG() for more
// information.
func NewAccumulator(newCipher NewCipher, seedFileName string) (*Accumulator, error) {
	acc := &Accumulator{
		seedFileName: seedFileName,
		gen:          NewGenerator(newCipher),
	}
	for i := 0; i < len(acc.pool); i++ {
		acc.pool[i] = sha256d.New()
	}

	if seedFileName != "" {
		// The initial seed of the generator depends on the current
		// time.  This protects us against old seed files being
		// restored from backups, etc.
		err := acc.updateSeedFile(seedFileName)
		if err == errReadFailed {
			err = acc.writeSeedFile(seedFileName)
		}
		if err != nil {
			return nil, err
		}

		quit := make(chan bool)
		acc.stopAutoSave = quit
		go func() {
			ticker := time.Tick(seedFileUpdateInterval)
			for {
				select {
				case <-quit:
					return
				case <-ticker:
					acc.writeSeedFile(seedFileName)
				}
			}
		}()
	}

	return acc, nil
}

func (acc *Accumulator) tryReseeding() []byte {
	now := time.Now()

	acc.poolMutex.Lock()
	defer acc.poolMutex.Unlock()

	if acc.poolZeroSize >= minPoolSize && now.After(acc.nextReseed) {
		acc.nextReseed = now.Add(minReseedInterval)
		acc.poolZeroSize = 0
		acc.reseedCount += 1

		seed := make([]byte, 0, numPools*sha256d.Size)
		pools := []string{}
		for i := uint(0); i < numPools; i++ {
			x := 1 << i
			if acc.reseedCount%x != 0 {
				break
			}
			seed = acc.pool[i].Sum(seed)
			acc.pool[i].Reset()
			pools = append(pools, strconv.Itoa(int(i)))
		}
		trace.T("fortuna/seed", trace.PrioInfo,
			"reseeding from pools %s", strings.Join(pools, " "))
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

// Read allows to extract randomness from the Accumulator using the
// io.Reader interface.  Read fills the byte slice p with random
// bytes.  The method always reads len(p) bytes and never returns an
// error.
func (acc *Accumulator) Read(p []byte) (n int, err error) {
	copy(p, acc.RandomData(uint(len(p))))
	return len(p), nil
}

// Close must be called before the program exits to ensure that the
// seed file is correctly updated.  After Close has been called the
// Accumulator must not be used any more.
func (acc *Accumulator) Close() error {
	// Reset the underlying PRNG to ensure that (1) the Accumulator
	// cannot be used any more after Close() has been called and (2)
	// information about the key is not retained in memory
	// indefinitely.
	acc.gen.reset()

	var err error
	if acc.seedFileName != "" {
		acc.stopAutoSave <- true
		err = acc.writeSeedFile(acc.seedFileName)
		acc.seedFileName= ""
	}
	return err
}
