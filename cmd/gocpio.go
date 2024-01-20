package main

import (
	"fmt"
	"os"
	"github.com/vilroi/gocpio"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s cpio-file\n", os.Args[0])
		os.Exit(1)
	}

	cpio := gocpio.ParseCpio(os.Args[1])
	cpio.ListFiles()
}
