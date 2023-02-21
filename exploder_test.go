package exploder_test

import (
	"log"
	"os"

	"github.com/pschou/go-exploder"
)

func ExampleExplode() {
	fh, err := os.Open("testdata.zip") // Open a file
	if err != nil {
		log.Fatal(err)
	}
	stat, err := fh.Stat() // Stat the file to get the size
	if err != nil {
		log.Fatal(err)
	}

	outputPath := "output/"

	err = exploder.Explode(outputPath, fh, stat.Size(), -1)
	// Output:
}
