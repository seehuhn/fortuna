// accumulator_test.go - unit tests for accumulator.go
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
	"crypto/rand"
	"io"
	mrand "math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAccumulator(t *testing.T) {
	// The reference values in this function are generated using the
	// "Python Cryptography Toolkit",
	// https://www.dlitz.net/software/pycrypto/ .

	acc, _ := NewRNG("")
	acc.gen.reset()

	acc.addRandomEvent(0, 0, make([]byte, 32))
	acc.addRandomEvent(0, 0, make([]byte, 32))
	for i := uint(0); i < 1000; i++ {
		acc.addRandomEvent(1, i, []byte{1, 2})
	}
	out := acc.RandomData(100)
	correct := []byte{
		226, 104, 210, 56, 80, 187, 224, 232, 131, 211, 35, 163, 49, 237, 24,
		137, 170, 13, 117, 170, 229, 75, 237, 29, 33, 53, 46, 187, 21, 154,
		18, 26, 157, 186, 69, 166, 241, 28, 148, 72, 62, 241, 150, 175, 15,
		70, 24, 125, 111, 133, 219, 77, 43, 112, 255, 243, 222, 152, 218, 61,
		101, 196, 45, 130, 161, 29, 73, 117, 91, 81, 24, 173, 24, 45, 48,
		90, 222, 127, 26, 195, 88, 191, 216, 22, 200, 245, 158, 162, 218, 10,
		72, 243, 193, 132, 171, 27, 179, 99, 54, 208,
	}
	if !bytes.Equal(out, correct) {
		t.Error("wrong RNG output")
	}

	acc.addRandomEvent(0, 0, make([]byte, 32))
	acc.addRandomEvent(0, 0, make([]byte, 32))
	out = acc.RandomData(100)
	correct = []byte{
		34, 163, 146, 161, 13, 93, 118, 204, 224, 58, 215, 141, 198, 90, 38,
		26, 174, 151, 129, 91, 249, 30, 91, 23, 199, 5, 180, 150, 94, 201,
		10, 223, 129, 189, 162, 116, 22, 255, 130, 183, 50, 39, 168, 7, 98,
		138, 223, 129, 231, 222, 193, 66, 59, 187, 16, 100, 171, 169, 194, 12,
		197, 121, 10, 238, 39, 203, 43, 201, 110, 91, 56, 44, 56, 44, 246,
		38, 25, 28, 94, 93, 65, 183, 85, 46, 61, 132, 18, 96, 131, 16,
		138, 241, 1, 22, 192, 249, 66, 242, 153, 112,
	}
	if !bytes.Equal(out, correct) {
		t.Error("wrong RNG output")
	}

	time.Sleep(200 * time.Millisecond)

	out = acc.RandomData(100)
	correct = []byte{
		98, 9, 233, 102, 1, 195, 243, 88, 163, 4, 58, 74, 146, 155, 152,
		92, 11, 229, 110, 108, 123, 100, 237, 1, 151, 50, 103, 163, 120, 47,
		209, 232, 249, 100, 33, 102, 126, 37, 133, 104, 57, 148, 187, 255, 186,
		232, 145, 182, 144, 141, 7, 12, 241, 184, 190, 72, 204, 123, 227, 250,
		14, 72, 4, 217, 167, 142, 222, 13, 245, 77, 224, 219, 176, 74, 20,
		13, 151, 138, 231, 135, 34, 192, 236, 5, 161, 249, 223, 212, 154, 198,
		14, 222, 197, 232, 75, 199, 134, 56, 58, 212,
	}
	if !bytes.Equal(out, correct) {
		t.Error("wrong RNG output")
	}
}

func TestClose(t *testing.T) {
	tempDir := os.TempDir()
	defer os.Remove(tempDir)
	seedFileName := filepath.Join(tempDir, "seed")

	for _, name := range []string{"", seedFileName} {
		acc, err := NewRNG(name)
		if err != nil {
			t.Error(err)
		}
		acc.RandomData(1)
		acc.Close()
		caughtAccessAfterClose := func() (hasPaniced bool) {
			defer func() {
				if r := recover(); r != nil {
					hasPaniced = true
				}
			}()
			acc.RandomData(1)
			return false
		}()
		if !caughtAccessAfterClose {
			t.Error("failed to detect RNG access after close")
		}
	}
}

func TestReseedingDuringClose(t *testing.T) {
	tempDir := os.TempDir()
	defer os.Remove(tempDir)
	seedFileName := filepath.Join(tempDir, "seed")

	acc, err := NewRNG(seedFileName)
	if err != nil {
		t.Error(err)
	}

	buf := make([]byte, 32)
	sink := acc.NewEntropyDataSink()
	for i := 0; i < numPools*32/minPoolSize; i++ {
		sink <- buf
	}
	close(sink)

	acc.Close()
}

func accumulatorRead(b *testing.B, n int) {
	acc, _ := NewRNG("")
	buffer := make([]byte, n)

	b.SetBytes(int64(n))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// acc.Read is guaranteed to return the full data in one go
		// and not to return an error.
		acc.Read(buffer)
	}
}

func BenchmarkAccumulatorRead16(b *testing.B) { accumulatorRead(b, 16) }
func BenchmarkAccumulatorRead32(b *testing.B) { accumulatorRead(b, 32) }
func BenchmarkAccumulatorRead1k(b *testing.B) { accumulatorRead(b, 1024) }

func cryptoRandRead(b *testing.B, n int) {
	buffer := make([]byte, n)

	b.SetBytes(int64(n))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := io.ReadFull(rand.Reader, buffer); err != nil {
			b.Fatalf(err.Error())
		}
	}
}

func BenchmarkCryptoRandRead16(b *testing.B) { cryptoRandRead(b, 16) }
func BenchmarkCryptoRandRead32(b *testing.B) { cryptoRandRead(b, 32) }
func BenchmarkCryptoRandRead1k(b *testing.B) { cryptoRandRead(b, 1024) }

func TestRandInt63(t *testing.T) {
	acc, _ := NewRNG("")
	for i := 0; i < 100; i++ {
		r := acc.Int63()
		if r < 0 {
			t.Error("Invalid random output")
		}
	}
}

func BenchmarkFortunaInt63(b *testing.B) {
	acc, _ := NewRNG("")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = acc.Int63()
	}
}

func BenchmarkFortunaUint64(b *testing.B) {
	acc, _ := NewRNG("")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = acc.Uint64()
	}
}

func BenchmarkMathRandInt63(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mrand.Int63()
	}
}

func BenchmarkMathRandUint64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = mrand.Uint64()
	}
}

func TestRandSeed(t *testing.T) {
	acc, _ := NewRNG("")
	defer func() {
		if r := recover(); r == nil {
			t.Error("Failed to panic")
		}
	}()
	acc.Seed(0)
}

// compile-time test: Accumulator implements the rand.Source interface
var _ mrand.Source = &Accumulator{}

// compile-time test: Accumulator implements the rand.Source64 interface
var _ mrand.Source64 = &Accumulator{}
