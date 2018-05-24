// common.go - auxiliary functions for the fortuna package
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

func bytesToUint64(bytes []byte) uint64 {
	var res uint64
	_ = bytes[7] // avoid range checks in the loop below
	res = uint64(bytes[0])
	for _, x := range bytes[1:] {
		res = res<<8 | uint64(x)
	}
	return res
}

func uint64ToBytes(x uint64) []byte {
	bytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		bytes[i] = byte(x & 0xff)
		x = x >> 8
	}
	return bytes
}

func bytesToInt64(bytes []byte) int64 {
	var res int64
	_ = bytes[7] // avoid range checks in the loop below
	res = int64(bytes[0])
	for _, x := range bytes[1:] {
		res = res<<8 | int64(x)
	}
	return res
}

func int64ToBytes(x int64) []byte {
	bytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		bytes[i] = byte(x & 0xff)
		x = x >> 8
	}
	return bytes
}

func isZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

func wipe(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
