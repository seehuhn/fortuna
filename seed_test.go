// seed_test.go - unit tests for seed.go
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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestSeedfile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	seedFileName := path.Join(tempDir, "seed")

	fmt.Println(seedFileName)

	acc, _ := NewRNG("")

	err = acc.writeSeedFile(seedFileName)
	if err != nil {
		t.Error(err)
	}

	err = acc.updateSeedFile(seedFileName)
	if err != nil {
		t.Error(err)
	}
}
