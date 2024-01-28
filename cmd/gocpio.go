package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vilroi/gocpio"
)

type CpioArgs struct {
	cpiofile string
	files    []string
}

var listFlag bool

func main() {
	args := parseArgs()
	cpio := gocpio.ParseCpio(args.cpiofile)

	if listFlag {
		cpio.ListFiles()
		os.Exit(0)
	}

	if len(args.files) != 0 {
		extractFiles(&cpio, args.files)
		os.Exit(0)
	}
	cpio.Test()

	//cpio.ExtractAllFiles()

}

func extractFiles(cpio *gocpio.Cpio, files []string) {
	for _, file := range files {
		cpio.ExtractFile(file)
	}
}

func parseArgs() CpioArgs {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	return CpioArgs{
		cpiofile: args[0],
		files:    args[1:],
	}
}

func init() {
	flag.BoolVar(&listFlag, "l", false, "list: lists all files within cpio archive")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [opts] cpio-file [files]\n", os.Args[0])
		flag.PrintDefaults()
	}
}
