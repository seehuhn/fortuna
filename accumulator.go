package fortuna

import (
	"hash"
	"sync"
	"time"

	"github.com/seehuhn/sha256d"
)

const minPoolSize = 64

type Accumulator struct {
	gen         *Generator
	reseedCount int
	lastReseed  time.Time

	poolMutex    sync.Mutex
	pool         [32]hash.Hash
	poolZeroSize int
	distribute   map[uint8]uint8
}

func NewAccumulator(newCipher NewCipher) (*Accumulator, error) {
	acc := &Accumulator{}

	gen, err := NewGenerator(newCipher)
	if err != nil {
		return nil, err
	}
	acc.gen = gen

	for i := 0; i < len(acc.pool); i++ {
		acc.pool[i] = sha256d.New()
	}

	return acc, nil
}

func (acc *Accumulator) RandomData(n int) ([]byte, error) {
	acc.poolMutex.Lock()
	now := time.Now()
	if acc.poolZeroSize < minPoolSize &&
		now.Sub(acc.lastReseed) > 100*time.Millisecond {
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

	return acc.gen.PseudoRandomData(n)
}

func (acc *Accumulator) AddRandomEvent(source uint8, data []byte) {
	acc.poolMutex.Lock()
	defer acc.poolMutex.Unlock()

	poolIndex := acc.distribute[source]
	acc.distribute[source] = (acc.distribute[source] + 1) % 32
	pool := acc.pool[poolIndex]

	pool.Write([]byte{source, byte(len(data))})
	pool.Write(data)
}
