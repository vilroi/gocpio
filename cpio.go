package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"

	unsafe "unsafe"
)

const (
	magicval string = "07070100"
)

type CpioMember struct {
	header CpioHeader
	data   []byte
}

type CpioHeader struct {
	Magic     string
	Inode     uint64
	Mode      uint64
	Uid       uint64
	Gid       uint64
	Nlink     uint64
	Mtime     uint64
	FileSize  uint64
	DevMajor  uint64
	DevMinor  uint64
	RDevMajor uint64
	RDevMinor uint64
	NameSize  uint64
	Check     uint64
}

func (header CpioHeader) VerifyMagic() bool {
	return header.Magic == magicval
}

type RawCpioHeader struct {
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

func (rawheader RawCpioHeader) ToCpioHeader() CpioHeader {
	var cpioheader CpioHeader

	cpioheader.Magic = string(rawheader.Magic[:])
	cpioheader.Inode = binary.LittleEndian.Uint64(rawheader.Inode[:])
	cpioheader.Mode = binary.LittleEndian.Uint64(rawheader.Mode[:])
	cpioheader.Uid = binary.LittleEndian.Uint64(rawheader.Uid[:])
	cpioheader.Gid = binary.LittleEndian.Uint64(rawheader.Gid[:])
	cpioheader.Nlink = binary.LittleEndian.Uint64(rawheader.Nlink[:])
	cpioheader.Mtime = binary.LittleEndian.Uint64(rawheader.Mtime[:])
	cpioheader.FileSize = binary.LittleEndian.Uint64(rawheader.FileSize[:])
	cpioheader.DevMajor = binary.LittleEndian.Uint64(rawheader.DevMajor[:])
	cpioheader.DevMinor = binary.LittleEndian.Uint64(rawheader.DevMinor[:])
	cpioheader.RDevMajor = binary.LittleEndian.Uint64(rawheader.RDevMajor[:])
	cpioheader.RDevMinor = binary.LittleEndian.Uint64(rawheader.RDevMinor[:])
	cpioheader.NameSize = binary.LittleEndian.Uint64(rawheader.NameSize[:])
	cpioheader.Check = binary.LittleEndian.Uint64(rawheader.Check[:])

	return cpioheader
}

func (rawheader RawCpioHeader) Dump() {
	fmt.Printf("Magic: %s\n", string(rawheader.Magic[:]))
	fmt.Printf("Inode: %s\n", string(rawheader.Inode[:]))
	fmt.Printf("Mode: %s\n", string(rawheader.Mode[:]))
	fmt.Printf("Uid: %s\n", string(rawheader.Uid[:]))
	fmt.Printf("NameSize: %s\n", string(rawheader.NameSize[:]))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s cpio-file\n", os.Args[0])
		os.Exit(1)
	}

	fd, err := os.Open(os.Args[1])
	check(err)

	info, err := fd.Stat()
	check(err)

	for nread := 0; nread < int(info.Size()); {
		var raw_header RawCpioHeader
		err = binary.Read(fd, binary.LittleEndian, &raw_header)
		check(err)
		nread += int(unsafe.Sizeof(raw_header))

		header := raw_header.ToCpioHeader()
		if !header.VerifyMagic() {
			fmt.Fprintf(os.Stderr, "invalid magc number: %s\n", string(raw_header.Magic[:]))
			os.Exit(1)
		}

		//	namebuf := make([]byte, raw_header.NameSize
		break
	}

	/*
		var header CpioHeader
		binary.Read(fd, binary.LittleEndian, &header)

		if header.VerifyMagic() == false {
			fmt.Println("not valid cpio header")
		}
	*/
}

func mmap(path string) []byte {
	fd, err := syscall.Open(os.Args[1], syscall.O_RDONLY, 0)
	check(err)

	var statbuf syscall.Stat_t
	err = syscall.Fstat(fd, &statbuf)
	check(err)

	bytes, err := syscall.Mmap(fd, 0, int(statbuf.Size), syscall.PROT_READ, syscall.MAP_ANONYMOUS|syscall.MAP_PRIVATE)
	check(err)

	return bytes
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
