package fortuna

import (
	"crypto/aes"
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

	acc := NewAccumulator(aes.NewCipher)
	acc.SetInitialSeed()

	err = acc.WriteSeedFile(seedFileName)
	if err != nil {
		t.Error(err)
	}

	acc.UpdateSeedFile(seedFileName)
	t.Error("XXX")
}
