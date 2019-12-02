// entropy.go - collect environmental entropy into pools
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
	"crypto/sha256"
	"time"
)

const channelBufferSize = 4

// addRandomEvent should be called periodically to add entropy to the
// state of the random number generator.  Different sources of
// randomness should use different values for the 'source' argument.
// The .allocateSource() method can be used to allocate source
// numbers.
//
// The value 'seq' is used to spread out entropy over the available
// entropy pools; for each entropy source, sequence values 0, 1, 2,
// ... should be passed in.  Finally, the argument 'data' gives the
// randomness to add to the pool.  'data' should be at most 32 bytes
// long; longer values should be hashed by the caller and the hash be
// submitted instead.
func (acc *Accumulator) addRandomEvent(source uint8, seq uint, data []byte) {
	pool := seq % numPools
	acc.poolMutex.Lock()
	defer acc.poolMutex.Unlock()

	poolHash := acc.pool[pool]
	_, err := poolHash.Write([]byte{source, byte(len(data))})
	if err != nil {
		return
	}

	_, err = poolHash.Write(data)
	if err != nil {
		return
	}

	if pool == 0 {
		acc.poolZeroSize += 2 + len(data)
	}
}

// allocateSource allocates a new source index for an entropy source.
func (acc *Accumulator) allocateSource() uint8 {
	acc.sourceMutex.Lock()
	defer acc.sourceMutex.Unlock()
	source := acc.nextSource
	acc.nextSource++
	return source
}

// NewEntropyDataSink returns a channel through which data can be
// submitted to the Accumulator's entropy pools.  Data should be
// written to the returned channel periodically to add entropy to the
// state of the random number generator.  The written data should be
// derived from quantities which change between calls and which cannot
// be (completely) known to an attacker.  Typical sources of
// randomness include noise from a microphone/camera, CPU cycle
// counters, or the number of processes running on the system.
//
// If the data written to the channel is longer than 32 bytes, the
// data is hashed internally and the hash is submitted to the entropy
// pools instead of the data itself.
//
// The channel can be closed by the caller to indicate that no more
// entropy will be sent via this channel.
func (acc *Accumulator) NewEntropyDataSink() chan<- []byte {
	source := acc.allocateSource()

	c := make(chan []byte, channelBufferSize)

	acc.sources.Add(1)
	go func() {
		defer acc.sources.Done()
		seq := uint(0)

	loop:
		for {
			select {
			case data, ok := <-c:
				if !ok {
					break loop
				}

				if len(data) > 32 {
					hash := sha256.New()
					_, err := hash.Write(data)
					if err != nil {
						return
					}

					data = hash.Sum(nil)
				}

				acc.addRandomEvent(source, seq, data)
				seq++
			case <-acc.stopSources:
				break loop
			}
		}
	}()

	return c
}

// NewEntropyTimeStampSink returns a channel through which timing data
// can be submitted to the Accumulator's entropy pools.  The current
// time should be written to the returned channel regularly to add
// entropy to the state of the random number generator.  The submitted
// times should be chosen such that they cannot be (completely) known
// to an attacker.  Typical sources of randomness include the arrival
// times of network packets or the times of key-presses by the user.
//
// The channel can be closed by the caller to indicate that no more
// entropy will be sent via this channel.
func (acc *Accumulator) NewEntropyTimeStampSink() chan<- time.Time {
	source := acc.allocateSource()

	c := make(chan time.Time, channelBufferSize)

	acc.sources.Add(1)
	go func() {
		defer acc.sources.Done()
		seq := uint(0)
		lastRequest := time.Now()

	loop:
		for {
			select {
			case now, ok := <-c:
				if !ok {
					break loop
				}

				dt := now.Sub(lastRequest)
				lastRequest = now

				acc.addRandomEvent(source, seq, int64ToBytes(int64(dt)))
				seq++
			case <-acc.stopSources:
				break loop
			}
		}
	}()

	return c
}
