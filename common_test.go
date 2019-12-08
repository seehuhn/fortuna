// accumulator_test.go - unit tests for common.go
// Copyright (C) 2014  Jochen Voss <voss@seehuhn.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package fortuna

import (
	"crypto/aes"
	"math"
	"testing"

	. "gopkg.in/check.v1"
)

func (s *fortunaSuite) TestUint64ToBytes(c *C) {
	testInts := []uint64{0, 1, 2, math.MaxUint64}
	for i := 0; i < 200; i++ {
		testInts = append(testInts, 10000000000*uint64(i)+100)
	}
	for _, x := range testInts {
		buf := uint64ToBytes(x)

		c.Assert((isZero(buf) && x != 0) || (!isZero(buf) && x == 0), Equals, false)

		y := bytesToUint64(buf)
		c.Assert(x, DeepEquals, y)
	}
}

func (s *fortunaSuite) TestBytesToUint64(c *C) {
	buf := make([]byte, 8)
	x := bytesToUint64(buf)
	c.Assert(x, DeepEquals, uint64(0x0))

	gen := NewGenerator(aes.NewCipher)
	gen.Seed(54321)
	for i := 0; i < 1000; i++ {
		buf = gen.PseudoRandomData(8)
		x := bytesToUint64(buf)
		buf2 := uint64ToBytes(x)
		c.Assert(buf, DeepEquals, buf2)
	}
}

func (s *fortunaSuite) TestInt64ToBytes(c *C) {
	testInts := []int64{math.MinInt64, math.MaxInt64, -1, 0, 1}
	for i := -100; i <= 100; i++ {
		testInts = append(testInts, 10000000000*int64(i)+100)
	}
	for _, x := range testInts {
		buf := int64ToBytes(x)

		c.Assert((isZero(buf) && x != 0) || (!isZero(buf) && x == 0), Equals, false)

		y := bytesToInt64(buf)
		c.Assert(x, DeepEquals, y)
	}
}

func (s *fortunaSuite) TestBytesToInt64(c *C) {
	buf := make([]byte, 8)
	x := bytesToInt64(buf)
	c.Assert(x, DeepEquals, int64(0x0))

	gen := NewGenerator(aes.NewCipher)
	gen.Seed(12345)
	for i := 0; i < 1000; i++ {
		buf = gen.PseudoRandomData(8)
		x := bytesToInt64(buf)
		buf2 := int64ToBytes(x)
		c.Assert(buf, DeepEquals, buf2)
	}
}

func (s *fortunaSuite) TestIsZero(c *C) {
	buf := make([]byte, 100)
	c.Assert(isZero(buf), Equals, true)

	buf[99] = 1
	c.Assert(isZero(buf), Equals, false)
}

func (s *fortunaSuite) TestWipe(c *C) {
	buf := []byte{1, 2, 3, 4, 5, 6, 7}
	wipe(buf)
	c.Assert(isZero(buf), Equals, true)
}

func BenchmarkBytesToUint64(b *testing.B) {
	buf := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bytesToUint64(buf)
	}
}

func BenchmarkBytesToInt64(b *testing.B) {
	buf := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bytesToInt64(buf)
	}
}
