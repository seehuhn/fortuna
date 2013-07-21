package fortuna

import (
	"bytes"
	"crypto/aes"
	"math"
	"math/rand"
	"testing"
)

func TestOutput(t *testing.T) {
	rng, err := NewGenerator(aes.NewCipher)
	if err != nil {
		t.Fatal(err.Error())
	}

	rng.Reseed([]byte{1, 2, 3, 4})
	out, _ := rng.PseudoRandomData(100)
	// The reference values in the variable 'correct' are generated
	// using the "Python Cryptography Toolkit" from
	// https://www.dlitz.net/software/pycrypto/ .
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

	rng.Reseed([]byte{5})
	out, _ = rng.PseudoRandomData(100)
	correct = []byte{
		201, 0, 38, 91, 149, 135, 150, 103, 206, 172, 243, 146, 22, 218, 114,
		200, 52, 54, 26, 45, 169, 60, 123, 161, 30, 131, 6, 142, 2, 41,
		32, 223, 118, 229, 56, 15, 111, 109, 200, 140, 251, 236, 59, 125, 130,
		133, 93, 141, 180, 137, 63, 253, 101, 15, 57, 240, 220, 130, 222, 44,
		237, 160, 125, 201, 224, 63, 229, 34, 143, 133, 24, 7, 189, 93, 91,
		57, 96, 100, 202, 1, 16, 127, 180, 117, 155, 27, 156, 34, 77, 229,
		157, 137, 63, 123, 196, 182, 231, 16, 219, 177,
	}
	if bytes.Compare(out, correct) != 0 {
		t.Error("wrong RNG output")
	}
}

func TestReseed(t *testing.T) {
	rng, err := NewGenerator(aes.NewCipher)
	if err != nil {
		t.Fatal(err.Error())
	}
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

func TestPrng(t *testing.T) {
	rng, err := NewGenerator(aes.NewCipher)
	if err != nil {
		t.Fatal(err.Error())
	}
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
