package fortuna

import (
	"crypto/aes"
	"testing"
)

func BenchmarkAddRandomEvent(b *testing.B) {
	acc := NewAccumulator(aes.NewCipher)

	b.ResetTimer()
	pool := uint8(0)
	for i := 0; i < b.N; i++ {
		acc.AddRandomEvent(0, pool, []byte{1, 2, 3})
		pool = (pool + 1) % 32
	}
}
