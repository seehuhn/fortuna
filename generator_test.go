// generator_test.go - unit tests for generator.go
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
	"crypto/aes"
	"math"
	"math/rand"
	"testing"
)

func TestOutput(t *testing.T) {
	// The reference values in this function are generated using the
	// "Python Cryptography Toolkit",
	// https://www.dlitz.net/software/pycrypto/ .

	rng := NewGenerator(aes.NewCipher)

	rng.Reseed([]byte{1, 2, 3, 4})
	out := rng.PseudoRandomData(100)
	correct := []byte{
		82, 254, 233, 139, 254, 85, 6, 222, 222, 149, 120, 35, 173, 71, 89,
		232, 51, 182, 252, 139, 153, 153, 111, 30, 16, 7, 124, 185, 159, 24,
		50, 68, 236, 107, 133, 18, 217, 219, 46, 134, 169, 156, 211, 74, 163,
		17, 100, 173, 26, 70, 246, 193, 57, 164, 167, 175, 233, 220, 160, 114,
		2, 200, 215, 80, 207, 218, 85, 58, 235, 117, 177, 223, 87, 192, 50,
		251, 61, 65, 141, 100, 59, 228, 23, 215, 58, 107, 248, 248, 103, 57,
		127, 31, 241, 91, 230, 33, 0, 164, 77, 46,
	}
	if bytes.Compare(out, correct) != 0 {
		t.Error("wrong RNG output")
	}

	out = rng.PseudoRandomData(1<<20 + 100)[1<<20:]
	correct = []byte{
		122, 164, 26, 67, 102, 65, 30, 217, 219, 113, 14, 86, 214, 146, 185,
		17, 107, 135, 183, 7, 18, 162, 126, 206, 46, 38, 54, 172, 248, 194,
		118, 84, 162, 146, 83, 156, 152, 96, 192, 15, 23, 224, 113, 76, 21,
		8, 226, 41, 161, 171, 197, 180, 138, 236, 126, 137, 101, 25, 219, 225,
		3, 189, 16, 242, 33, 91, 34, 27, 8, 171, 171, 115, 157, 109, 248,
		198, 227, 18, 204, 211, 42, 184, 92, 42, 171, 222, 198, 117, 162, 134,
		116, 109, 77, 195, 187, 139, 37, 78, 224, 63,
	}
	if bytes.Compare(out, correct) != 0 {
		t.Error("wrong RNG output")
	}

	rng.Reseed([]byte{5})
	out = rng.PseudoRandomData(100)
	correct = []byte{
		217, 168, 141, 167, 46, 9, 218, 188, 98, 124, 109, 128, 242, 22, 189,
		120, 180, 124, 15, 192, 116, 149, 211, 136, 253, 132, 60, 3, 29, 250,
		95, 66, 133, 195, 37, 78, 242, 255, 160, 209, 185, 106, 68, 105, 83,
		145, 165, 72, 179, 167, 53, 254, 183, 251, 128, 69, 78, 156, 219, 26,
		124, 202, 35, 9, 174, 167, 41, 128, 184, 25, 2, 1, 63, 142, 205,
		162, 69, 68, 207, 251, 101, 10, 29, 33, 133, 87, 189, 36, 229, 56,
		17, 100, 138, 49, 79, 239, 210, 189, 141, 46,
	}
	if bytes.Compare(out, correct) != 0 {
		t.Error("wrong RNG output")
	}
}

func TestReseed(t *testing.T) {
	rng := NewGenerator(aes.NewCipher)
	if len(rng.key) != 32 {
		t.Error("wrong key size")
	}
	if len(rng.counter) != 16 {
		t.Error("wrong key size")
	}

	rng.Reseed(nil)
	if len(rng.key) != 32 {
		t.Error("wrong key size after reseeding")
	}
	if len(rng.counter) != 16 {
		t.Error("wrong key size after reseeding")
	}
}

func TestSeed(t *testing.T) {
	rng := NewGenerator(aes.NewCipher)

	for _, seed := range []int64{0, 1, 1 << 62} {
		rng.Seed(seed)
		x := rng.PseudoRandomData(1000)
		rng.Seed(seed)
		y := rng.PseudoRandomData(1000)
		if bytes.Compare(x, y) != 0 {
			t.Error(".Seed() doesn't determine generator state")
		}
	}
}

func TestPrng(t *testing.T) {
	rng := NewGenerator(aes.NewCipher)
	rng.Seed(123)

	prng := rand.New(rng)
	n := 1000000
	pos := 0
	for i := 0; i < n; i++ {
		x := prng.NormFloat64()
		if x > 0 {
			pos += 1
		}
	}

	d := (float64(pos) - 0.5*float64(n)) / math.Sqrt(0.25*float64(n))
	if math.Abs(d) >= 4 {
		t.Error("wrong distribution")
	}
}

func BenchmarkReseed(b *testing.B) {
	rng := NewGenerator(aes.NewCipher)
	seed := []byte{1, 2, 3, 4}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rng.Reseed(seed)
	}
}

func generator(b *testing.B, n uint) {
	rng := NewGenerator(aes.NewCipher)
	rng.Seed(0)

	b.SetBytes(int64(n))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng.PseudoRandomData(n)
	}
}

func BenchmarkGenerator16(b *testing.B) { generator(b, 16) }
func BenchmarkGenerator32(b *testing.B) { generator(b, 32) }
func BenchmarkGenerator1k(b *testing.B) { generator(b, 1024) }
