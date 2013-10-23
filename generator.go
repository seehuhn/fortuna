// generator.go - a cryptographically strong PRNG
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
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"net"
	"os/user"
	"time"

	"github.com/seehuhn/sha256d"
	"github.com/seehuhn/trace"
)

const (
	// maxBlocks gives the maximal number of blocks to generate until
	// rekeying is required.
	maxBlocks = 1 << 16

	// keySize gives the size of the internal key in bytes
	keySize = sha256d.Size
)

// NewCipher is the type which represents the function to allocate a
// new block cipher.  A typical example of a function of this type is
// aes.NewCipher.
type NewCipher func([]byte) (cipher.Block, error)

// Generator holds the state of one instance of the Fortuna pseudo
// random number generator.  Before use, the generator must be seeded
// using the Reseed() or Seed() method.  Randomness can then be
// extracted using the PseudoRandomData() method.  The Generator class
// implements the rand.Source interface.
//
// This Generator class is not safe for use with concurrent accesss.
// If the generator is accessed from different Go-routines, the
// callers must synchronise access using sync.Mutex or similar.
type Generator struct {
	newCipher NewCipher
	key       []byte
	cipher    cipher.Block
	counter   []byte
}

func (gen *Generator) inc() {
	// The counter is stored least-significant byte first.
	for i := 0; i < len(gen.counter); i++ {
		gen.counter[i]++
		if gen.counter[i] != 0 {
			break
		}
	}
}

func (gen *Generator) setKey(key []byte) {
	if len(key) != keySize {
		panic("wrong key size")
	}
	gen.key = key
	cipher, err := gen.newCipher(gen.key)
	if err != nil {
		panic("newCipher() failed, cannot set generator key")
	}
	gen.cipher = cipher
}

// setInitialSeed sets the initial seed for the Generator.  An
// attempt is made to obtain seeds which differ between machines and
// between reboots.  To achieve this, the following information is
// incorporated into the seed: the current time of day, account
// information for the current user, and information about the
// installed network interfaces.  In addition, if available, random
// bytes from the random number generator in the crypto/rand package
// are used.
func (gen *Generator) setInitialSeed() {
	// source 1: system random number generator
	buffer := make([]byte, keySize)
	n, _ := io.ReadFull(rand.Reader, buffer)
	if n > 0 {
		trace.T("fortuna/seed", trace.PrioInfo,
			"mixing %d bytes from crypto/rand into the seed", n)
		gen.Reseed(buffer)
	}

	// source 2: current time of day
	now := time.Now()
	trace.T("fortuna/seed", trace.PrioInfo,
		"mixing the current time into the seed")
	gen.Reseed([]byte(now.String()))

	// source 3: try different files with timer information, interrupt
	// counts, etc.
	for _, fname := range []string{"/proc/timer_list", "/proc/stat"} {
		buffer, _ = ioutil.ReadFile(fname)
		if len(buffer) > 0 {
			trace.T("fortuna/seed", trace.PrioInfo,
				"mixing %d bytes from %q into the seed", len(buffer), fname)
			gen.Reseed(buffer)
		}
	}

	// source 4: user name and login details
	user, _ := user.Current()
	if user != nil {
		trace.T("fortuna/seed", trace.PrioInfo,
			"mixing information about the current user into the seed")
		gen.Reseed([]byte(user.Uid))
		gen.Reseed([]byte(user.Gid))
		gen.Reseed([]byte(user.Username))
		gen.Reseed([]byte(user.Name))
		gen.Reseed([]byte(user.HomeDir))
	}

	// source 5: network interfaces
	ifaces, _ := net.Interfaces()
	if ifaces != nil {
		trace.T("fortuna/seed", trace.PrioInfo,
			"mixing network interface information into the seed")
		for _, iface := range ifaces {
			gen.ReseedInt64(int64(iface.MTU))
			gen.Reseed([]byte(iface.Name))
			gen.Reseed(iface.HardwareAddr)
			gen.ReseedInt64(int64(iface.Flags))
		}
	}
}

// NewGenerator creates a new instance of the Fortuna pseudo random
// number generator.  The function newCipher should normally be
// aes.NewCipher from the crypto/aes package, but the Serpent or
// Twofish ciphers can also be used.
//
// The initial seed is chosen based on the current time, the current
// user name, the currently installed network interfaces and
// randomness from the system random number generator.
func NewGenerator(newCipher NewCipher) *Generator {
	gen := &Generator{
		newCipher: newCipher,
	}
	gen.reset()
	gen.setInitialSeed()

	return gen
}

// reset reverts the generated to the unseeded state.  A new seed must
// be set using the .Reseed() or .Seed() methods before the generator
// can be used again.  This is mostly useful for unit testing, to
// start the PRNG from a known state.
func (gen *Generator) reset() {
	zeroKey := make([]byte, keySize)
	gen.setKey(zeroKey)
	gen.counter = make([]byte, gen.cipher.BlockSize())
}

// Reseed uses the current generator state and the given seed value to
// update the generator state.  Care is taken to make sure that
// knowledge of the new state after a reseed does not allow to
// reconstruct previous output values of the generator.
//
// This is like the ReseedInt64() method, but the seed is given as a
// byte slice instead of as an int64.
func (gen *Generator) Reseed(seed []byte) {
	hash := sha256d.New()
	hash.Write(gen.key)
	hash.Write(seed)
	gen.setKey(hash.Sum(nil))
	gen.inc()
	trace.T("fortuna/generator", trace.PrioVerbose, "seed updated")
}

// ReseedInt64 uses the current generator state and the given seed
// value to update the generator state.  Care is taken to make sure
// that knowledge of the new state after a reseed does not allow to
// reconstruct previous output values of the generator.
//
// This is like the Reseed() method, but the seed is given as an int64
// instead of as a byte slice.
func (gen *Generator) ReseedInt64(seed int64) {
	bytes := int64ToBytes(seed)
	gen.Reseed(bytes)
}

// generateBlocks appends k blocks of random bits to data and returns
// the resulting slice.  The size of a block is given by the block
// size of the underlying cipher, i.e. 16 bytes for AES.
func (gen *Generator) generateBlocks(data []byte, k uint) []byte {
	if isZero(gen.counter) {
		panic("Fortuna generator not yet seeded")
	}

	counterSize := uint(len(gen.counter))
	buf := make([]byte, counterSize)
	for i := uint(0); i < k; i++ {
		gen.cipher.Encrypt(buf, gen.counter)
		data = append(data, buf...)
		gen.inc()
	}

	return data
}

func (gen *Generator) numBlocks(n uint) uint {
	k := uint(len(gen.counter))
	return (n + k - 1) / k
}

// PseudoRandomData returns a slice of n pseudo-random bytes.  The
// result can be used as a replacement for a sequence of n uniformly
// distributed and independent bytes.
func (gen *Generator) PseudoRandomData(n uint) []byte {
	numBlocks := gen.numBlocks(n)
	res := make([]byte, 0, numBlocks*uint(len(gen.counter)))

	for numBlocks > 0 {
		count := numBlocks
		if count > maxBlocks {
			count = maxBlocks
		}
		res = gen.generateBlocks(res, count)
		numBlocks -= count

		newKey := gen.generateBlocks(nil, gen.numBlocks(keySize))
		gen.setKey(newKey[:keySize])
	}

	trace.T("fortuna/generator", trace.PrioDebug,
		"generated %d pseudo-random bytes", n)
	return res[:n]
}

// Int63 returns a positive random integer, uniformly distributed on
// the range 0, 1, ..., 2^63-1.  This function is part of the
// rand.Source interface.
func (gen *Generator) Int63() int64 {
	bytes := gen.PseudoRandomData(8)
	bytes[0] &= 0x7f
	return bytesToInt64(bytes)
}

// Seed uses the given seed value to set a new generator state.  In
// contrast to the Reseed() method, the Seed() method discards all
// previous state, thus allowing to generate reproducible output.
// This function is part of the rand.Source interface.
//
// Use of this method should be avoided in cryptographic applications,
// since reproducible output will lead to security vulnerabilities.
func (gen *Generator) Seed(seed int64) {
	gen.reset()
	gen.ReseedInt64(seed)
}
