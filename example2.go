// example2.go - a test program for the fortuna package
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

// +build ignore

package main

import (
	"crypto/aes"
	"io"
	"log"
	"os"

	"github.com/seehuhn/fortuna"
)

const (
	outputFileName = "fortuna.out"
	outputFileSize = 1024 * 1024 * 1024
)

func main() {
	acc, _ := fortuna.NewAccumulator(aes.NewCipher, "")
	acc.SetInitialSeed()

	out, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("cannot open %s: %s", outputFileName, err.Error())
	}
	defer out.Close()

	n, _ := io.CopyN(out, acc, outputFileSize)
	log.Printf("wrote %d random bytes to %s", n, outputFileName)
}
