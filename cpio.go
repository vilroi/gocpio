package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	magicval string = "07070100"
)

type CpioHeader struct {
	Magic     [8]byte
	Inode     [8]byte
	Mode      [8]byte
	Uid       [8]byte
	Gid       [8]byte
	Nlink     [8]byte
	Mtime     [8]byte
	FileSize  [8]byte
	DevMajor  [8]byte
	DevMinor  [8]byte
	RDevMajor [8]byte
	RDevMinor [8]byte
	NameSize  [8]byte
	Check     [8]byte
}

func (header CpioHeader) Dump() {
	fmt.Printf("Magic: %s\n", string(header.Magic[:]))
	fmt.Printf("Inode: %s\n", string(header.Inode[:]))
	fmt.Printf("Mode: %s\n", string(header.Mode[:]))
	fmt.Printf("Uid: %s\n", string(header.Uid[:]))
	fmt.Printf("NameSize: %s\n", string(header.NameSize[:]))
}

func (header CpioHeader) VerifyMagic() bool {
	return string(header.Magic[:]) == magicval
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s cpio-file\n", os.Args[0])
		os.Exit(1)
	}

	fd, err := os.Open(os.Args[1])
	check(err)

	var header CpioHeader
	binary.Read(fd, binary.LittleEndian, &header)

	if header.VerifyMagic() == false {
		fmt.Println("not valid cpio header")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
