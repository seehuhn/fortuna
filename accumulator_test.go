package fortuna

import (
	"bytes"
	"crypto/aes"
	"testing"
	"time"
)

func TestAccumulator(t *testing.T) {
	// The reference values in this function are generated using the
	// "Python Cryptography Toolkit",
	// https://www.dlitz.net/software/pycrypto/ .

	acc := NewAccumulator(aes.NewCipher)

	acc.AddRandomEvent(0, 0, make([]byte, 32))
	acc.AddRandomEvent(0, 0, make([]byte, 32))
	for i := 0; i < 1000; i++ {
		acc.AddRandomEvent(1, uint8(i%32), []byte{1, 2})
	}
	out := acc.RandomData(100)
	correct := []byte{
		226, 104, 210, 56, 80, 187, 224, 232, 131, 211, 35, 163, 49, 237, 24,
		137, 170, 13, 117, 170, 229, 75, 237, 29, 33, 53, 46, 187, 21, 154,
		18, 26, 157, 186, 69, 166, 241, 28, 148, 72, 62, 241, 150, 175, 15,
		70, 24, 125, 111, 133, 219, 77, 43, 112, 255, 243, 222, 152, 218, 61,
		101, 196, 45, 130, 161, 29, 73, 117, 91, 81, 24, 173, 24, 45, 48,
		90, 222, 127, 26, 195, 88, 191, 216, 22, 200, 245, 158, 162, 218, 10,
		72, 243, 193, 132, 171, 27, 179, 99, 54, 208,
	}
	if bytes.Compare(out, correct) != 0 {
		t.Error("wrong RNG output")
	}

	acc.AddRandomEvent(0, 0, make([]byte, 32))
	acc.AddRandomEvent(0, 0, make([]byte, 32))
	out = acc.RandomData(100)
	correct = []byte{
		34, 163, 146, 161, 13, 93, 118, 204, 224, 58, 215, 141, 198, 90, 38,
		26, 174, 151, 129, 91, 249, 30, 91, 23, 199, 5, 180, 150, 94, 201,
		10, 223, 129, 189, 162, 116, 22, 255, 130, 183, 50, 39, 168, 7, 98,
		138, 223, 129, 231, 222, 193, 66, 59, 187, 16, 100, 171, 169, 194, 12,
		197, 121, 10, 238, 39, 203, 43, 201, 110, 91, 56, 44, 56, 44, 246,
		38, 25, 28, 94, 93, 65, 183, 85, 46, 61, 132, 18, 96, 131, 16,
		138, 241, 1, 22, 192, 249, 66, 242, 153, 112,
	}
	if bytes.Compare(out, correct) != 0 {
		t.Error("wrong RNG output")
	}

	time.Sleep(200 * time.Millisecond)

	out = acc.RandomData(100)
	correct = []byte{
		98, 9, 233, 102, 1, 195, 243, 88, 163, 4, 58, 74, 146, 155, 152,
		92, 11, 229, 110, 108, 123, 100, 237, 1, 151, 50, 103, 163, 120, 47,
		209, 232, 249, 100, 33, 102, 126, 37, 133, 104, 57, 148, 187, 255, 186,
		232, 145, 182, 144, 141, 7, 12, 241, 184, 190, 72, 204, 123, 227, 250,
		14, 72, 4, 217, 167, 142, 222, 13, 245, 77, 224, 219, 176, 74, 20,
		13, 151, 138, 231, 135, 34, 192, 236, 5, 161, 249, 223, 212, 154, 198,
		14, 222, 197, 232, 75, 199, 134, 56, 58, 212,
	}
	if bytes.Compare(out, correct) != 0 {
		t.Error("wrong RNG output")
	}
}

func BenchmarkAddRandomEvent(b *testing.B) {
	acc := NewAccumulator(aes.NewCipher)

	b.ResetTimer()
	pool := uint8(0)
	for i := 0; i < b.N; i++ {
		acc.AddRandomEvent(0, pool, []byte{1, 2, 3})
		pool = (pool + 1) % 32
	}
}
