package fortuna

import (
	"testing"
)

func TestReseed(t *testing.T) {
	rng := &Fortuna{}
	rng.Init()
	rng.Reseed(nil)
	if len(rng.key) != keySize {
		t.Error("wrong key size")
	}
}
