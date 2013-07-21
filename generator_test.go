package fortuna

import (
	"crypto/aes"
	"math"
	"math/rand"
	"testing"
)

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
