package fortuna

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/seehuhn/sha256d"
)

const (
	keySize       = 32
	countSize     = 16
	fullBlockSize = 16 * (1 << 16)
)

type Fortuna struct {
	key   []byte
	count []byte
	pools [32][]byte
}

func (f *Fortuna) inc() {
	for i := 0; i < countSize; i++ {
		f.count[i] += 1
		if f.count[i] != 0 {
			break
		}
	}
}

func (f *Fortuna) Init() {
	f.key = make([]byte, keySize)
	f.count = make([]byte, countSize)
}

func (f *Fortuna) Reseed(seed []byte) {
	hash := sha256d.New()
	hash.Write(f.key)
	hash.Write(seed)
	f.key = hash.Sum(nil)
	f.inc()
}

func (f *Fortuna) generate(size uint) []byte {
	res := make([]byte, 0, size)

	numFullBlocks := size / fullBlockSize
	remainder := size % fullBlockSize
	for i := uint(0); i < numFullBlocks; i++ {
		res = append(res, f.randomBytes(1<<20)...)
		f.key = f.randomBytes(keySize)
	}
	res = append(res, f.randomBytes(remainder)...)
	return res
}

func (f *Fortuna) randomBytes(size uint) []byte {
	if f.key == nil {
		panic("Fortuna generator is not yet seeded")
	}

	block, err := aes.NewCipher(f.key)
	if err != nil {
		panic(err)
	}
	stream := cipher.NewCTR(block, f.count)

	data := make([]byte, size)
	stream.XORKeyStream(data, data)
	return data
}
