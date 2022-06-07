package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cool-develope/rdiff-go/chunker"
	"github.com/cool-develope/rdiff-go/delta"
	"github.com/cool-develope/rdiff-go/signature"
)

var srcFileName = flag.String("source", "", "source file (required)")
var targetFileName = flag.String("target", "", "target file (required)")
var windowSize = flag.Int("window", 64, "average chunk size")

func main() {
	flag.Parse()

	if *srcFileName == "" {
		fatalf("source file is required")
	}

	if *targetFileName == "" {
		fatalf("target file is required")
	}

	srcFile, err := os.Open(*srcFileName)
	if err != nil {
		fatalf("unable to open source file: %v", err)
	}
	defer srcFile.Close() //nolint

	targetFile, err := os.Open(*targetFileName)
	if err != nil {
		fatalf("unable to open target file: %v", err)
	}
	defer targetFile.Close() //nolint

	cin := chunker.NewChunker(srcFile, uint32(*windowSize))
	sig, err := signature.GetSignature(cin, uint32(*windowSize))
	if err != nil {
		fatalf("signature generating failed: %v", err)
	}

	cout := chunker.NewChunker(targetFile, uint32(*windowSize))
	delta, err := delta.GetDelta(sig, cout)
	if err != nil {
		fatalf("getting delta failed: %v", err)
	}

	fmt.Printf("%+v", delta)
}

func fatalf(format string, a ...interface{}) {
	format = fmt.Sprintf("ERROR: %s\n", format)
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
