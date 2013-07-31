package fortuna

import (
	"testing"
	"time"
)

func TestPoolSelection(t *testing.T) {
	acc, _ := NewAccumulatorAES("")
	acc.SetInitialSeed()
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
	acc, _ := NewAccumulatorAES("")
	acc.SetInitialSeed()
	source := acc.allocateSource()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		acc.addRandomEvent(source, uint(i), []byte{1, 2, 3, 4, 5, 6, 7, 8})
	}
}

func BenchmarkDataSink(b *testing.B) {
	acc, _ := NewAccumulatorAES("")
	acc.SetInitialSeed()
	sink := acc.NewEntropyDataSink()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink <- []byte{1, 2, 3, 4, 5, 6, 7, 8}
	}
}

func BenchmarkTimeStampSink(b *testing.B) {
	acc, _ := NewAccumulatorAES("")
	acc.SetInitialSeed()
	sink := acc.NewEntropyTimeStampSink()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink <- time.Now()
	}
}
