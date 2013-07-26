// example1.go - a test program for the fortuna package
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
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/seehuhn/fortuna"
	"github.com/seehuhn/trace"
)

const seedFileName = "seed.dat"

func myListener(t time.Time, path string, prio trace.Priority, msg string) {
	fmt.Printf("%s:%s: %s\n", t.Format("15:04:05.000"), path, msg)
}

func main() {
	trace.Register(myListener, "", trace.PrioInfo)

	acc, err := fortuna.NewAccumulator(aes.NewCipher, seedFileName)
	if err != nil {
		panic("cannot initialise the RNG: " + err.Error())
	}
	defer acc.Close()

	// entropy source 0: mix in randomness from crypto/rand every minute
	go func() {
		seq0 := uint(0)

		tick := time.Tick(time.Minute)
		for _ = range tick {
			buffer := make([]byte, 4)
			n, _ := rand.Read(buffer)
			trace.T("main/entropy", trace.PrioInfo,
				"adding %d bytes of entropy from crypto/rand", n)
			acc.AddRandomEvent(0, seq0, buffer)
			seq0 += 1
		}
	}()

	seq1 := uint(0)
	lastRequest := time.Now()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// entropy source 1: time between requests
		now := time.Now()
		dt := now.Sub(lastRequest)
		lastRequest = now
		entropy := dt.String()
		trace.T("main/entropy", trace.PrioInfo,
			"adding timer entropy %q", entropy)
		acc.AddRandomEvent(1, seq1, []byte(entropy))
		seq1 += 1

		sizeStr := r.URL.Query().Get("len")
		size, _ := strconv.ParseInt(sizeStr, 0, 32)
		if size <= 0 {
			size = 16
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))

		io.CopyN(w, acc, size)
	})
	fmt.Println("listening at http://localhost:8080/")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("error %q, aborting ...\n", err.Error())
	}
}
