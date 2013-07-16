package fortuna

import (
	"github.com/seehuhn/sha256d"
)

const (
	fortunaKeySize = 32
)

type Fortuna struct {
	k           []byte
	cLow, cHigh uint64
	pools       [32][]byte
}

func (f *Fortuna) inc() {
	f.cLow += 1
	if f.cLow == 0 {
		f.cHigh += 1
	}
}

func (f *Fortuna) Init() {
	f.k = make([]byte, fortunaKeySize)
	f.cLow = 0
	f.cHigh = 0
}

func (f *Fortuna) Reseed(seed []byte) {
	hash := sha256d.New()
	hash.Write(f.k)
	hash.Write(seed)
	f.k = hash.Sum(nil)
	f.inc()
}

func (f *Fortuna) generateBlocks(count uint) {
	if f.k == nil {
		panic("Fortuna generator is not yet seeded")
	}
	// ...
}
