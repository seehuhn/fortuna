package fortuna

import (
	"hash"
	"sync"
	"time"

	"github.com/seehuhn/sha256d"
)

const minPoolSize = 64

type Accumulator struct {
	genMutex sync.Mutex
	gen      *Generator

	poolMutex    sync.Mutex
	reseedCount  int
	lastReseed   time.Time
	pool         [32]hash.Hash
	poolZeroSize int
}

func NewAccumulator(newCipher NewCipher) *Accumulator {
	acc := &Accumulator{
		gen: NewGenerator(newCipher),
	}
	for i := 0; i < len(acc.pool); i++ {
		acc.pool[i] = sha256d.New()
	}
	return acc
}

func (acc *Accumulator) RandomData(n uint) []byte {
	acc.poolMutex.Lock()
	now := time.Now()
	if acc.poolZeroSize < minPoolSize &&
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

		acc.gen.Reseed(seed)
	}
	acc.poolMutex.Unlock()

	acc.genMutex.Lock()
	defer acc.genMutex.Unlock()
	return acc.gen.PseudoRandomData(n)
}

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
