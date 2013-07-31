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

func printTrace(t time.Time, path string, prio trace.Priority, msg string) {
	fmt.Printf("%s:%s: %s\n", t.Format("15:04:05.000"), path, msg)
}

func main() {
	trace.Register(printTrace, "", trace.PrioDebug)

	acc, err := fortuna.NewAccumulatorAES(seedFileName)
	if err != nil {
		panic("cannot initialise the RNG: " + err.Error())
	}
	defer acc.Close()

	// entropy source 1: submit some randomness from crypto/rand once a minute
	go func() {
		sink1 := acc.NewEntropyDataSink()
		for _ = range time.Tick(time.Minute) {
			buffer := make([]byte, 4)
			n, _ := rand.Read(buffer)
			sink1 <- buffer[:n]
		}
	}()

	// entropy source 2: submit time between requests
	sink2 := acc.NewEntropyTimeStampSink()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sink2 <- time.Now()

		sizeStr := r.URL.Query().Get("len")
		size, _ := strconv.ParseInt(sizeStr, 0, 32)
		if size <= 0 {
			size = 16
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))

		io.CopyN(w, acc, size)
		trace.T("main", trace.PrioInfo,
			"sent %d random bytes for %q", size, r.RequestURI)
	})

	listenAddr := ":8080"
	trace.T("main", trace.PrioInfo,
		"listening on http://localhost%s/", listenAddr)
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		trace.T("main", trace.PrioCritical, "%s", err.Error())
	}
}
