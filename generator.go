package fortuna

import (
	"crypto/cipher"

	"github.com/seehuhn/errors"
	"github.com/seehuhn/sha256d"
)

// NewCipher represents functions which allocate a new block cipher.
// An example of a function which is of this type is aes.NewCipher.
type NewCipher func([]byte) (cipher.Block, error)

type Generator struct {
	newCipher NewCipher
	key       []byte
	counter   []byte
}

func (gen *Generator) inc() {
	// The counter is stored least-signigicant byte first.
	for i := 0; i < len(gen.counter); i++ {
		gen.counter[i]++
		if gen.counter[i] != 0 {
			break
		}
	}
}

// NewGenerator creates a new instance of the Fortuna random number
// generator.  The function newCipher should normally be aes.NewCipher
// from the crypto/aes package, but the Serpent or Twofish ciphers can
// also be used.
func NewGenerator(newCipher NewCipher) (*Generator, error) {
	gen := &Generator{
		newCipher: newCipher,
		key:       make([]byte, sha256d.Size),
	}

	// Try whether newCipher works and, at the same time, use the test
	// cipher to determine the block size.
	tmpCipher, err := newCipher(gen.key)
	if err != nil {
		return nil, errors.NewError(errors.EInvalidArgument,
			"fortuna/generator",
			"newCipher() failed with key of length %s: %s",
			len(gen.key), err.Error())
	}

	gen.counter = make([]byte, tmpCipher.BlockSize())

	return gen, nil
}

func (gen *Generator) Reseed(seed []byte) {
	hash := sha256d.New()
	hash.Write(gen.key)
	hash.Write(seed)
	gen.key = hash.Sum(nil)
	gen.inc()
}

func isZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

func (gen *Generator) generateBlocks(k int) []byte {
	if isZero(gen.counter) {
		panic("generator not yet seeded")
	}

	cipher, err := gen.newCipher(gen.key)
	if err != nil {
		panic("newCipher() failed")
	}

	counterSize := len(gen.counter)
	res := make([]byte, k*counterSize)
	for i := 0; i < k; i++ {
		cipher.Encrypt(res[i*counterSize:(i+1)*counterSize], gen.counter)
		gen.inc()
	}

	return res
}

func (gen *Generator) numBlocks(n int) int {
	k := len(gen.counter)
	return (n + k - 1) / k
}

func (gen *Generator) PseudoRandomData(n int) ([]byte, error) {
	if n < 0 || n > (1<<20) {
		return nil, errors.NewError(errors.EInvalidArgument,
			"fortuna/generator",
			"requested output size outside the valid range 0, ..., 2^20")
	}

	res := gen.generateBlocks(gen.numBlocks(n))

	keySize := len(gen.key)
	newKey := gen.generateBlocks(gen.numBlocks(keySize))
	gen.key = newKey[:keySize]

	return res[:n], nil
}

func bytesToInt64(bytes []byte) int64 {
	var res int64
	res = int64(bytes[0])
	for _, x := range bytes[1:] {
		res = res<<8 | int64(x)
	}
	return res
}

// Int63 returns a positive random integer, uniformly distributed on
// the range 0, 1, ..., 2^63-1.  This function is part of the
// rand.Source interface.
func (gen *Generator) Int63() int64 {
	bytes, _ := gen.PseudoRandomData(8)
	bytes[0] &= 0x7f
	return bytesToInt64(bytes)
}

func int64ToBytes(x int64) []byte {
	bytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		bytes[i] = byte(x & 0xff)
		x = x >> 8
	}
	return bytes
}

// Seed reseeds the generated, using the given seed value to determine
// the new generator state.  In contrast to the Reseed method, the
// Seed method discards any previously used state, thus leading to
// reproducible results.  This function is part of the rand.Source
// interface.
func (gen *Generator) Seed(seed int64) {
	bytes := int64ToBytes(seed)
	gen.key = make([]byte, len(gen.key))
	gen.Reseed(bytes)
}
