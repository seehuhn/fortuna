// entropy_test.go - unit tests for entropy.go
// Copyright (C) 2013  Jochen Voss <voss@seehuhn.de>
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
	"testing"
	"time"
)

func TestPoolSelection(t *testing.T) {
	acc, _ := NewRNG("")
	sink := acc.NewEntropyDataSink()

	msg := []byte{0}
	for i := 0; i < numPools+channelBufferSize+1; i++ {
		sink <- msg
	}
	acc.poolMutex.Lock()
	size := acc.poolZeroSize
	acc.poolMutex.Unlock()

	if size != 2*(2+len(msg)) {
		t.Error("distribution of events over pools failed")
	}
}

func BenchmarkAddRandomEvent(b *testing.B) {
	acc, _ := NewRNG("")
	source := acc.allocateSource()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acc.addRandomEvent(source, uint(i), []byte{1, 2, 3, 4, 5, 6, 7, 8})
	}
}

func BenchmarkDataSink(b *testing.B) {
	acc, _ := NewRNG("")
	sink := acc.NewEntropyDataSink()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink <- []byte{1, 2, 3, 4, 5, 6, 7, 8}
	}
}

func BenchmarkTimeStampSink(b *testing.B) {
	acc, _ := NewRNG("")
	sink := acc.NewEntropyTimeStampSink()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink <- time.Now()
	}
}
