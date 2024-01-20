package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vilroi/gocpio"
)

var listFlag bool

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	cpio := gocpio.ParseCpio(args[0])

	if listFlag {
		cpio.ListFiles()
		os.Exit(0)
	}

	if len(args[1:]) != 0 {
		for _, file := range args[1:] {
			cpio.ExtractFile(file)
		}
	}
}

func init() {
	flag.BoolVar(&listFlag, "l", false, "list: lists all files within cpio archive")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [opts] cpio-file [files]\n", os.Args[0])
		flag.PrintDefaults()
	}
}
