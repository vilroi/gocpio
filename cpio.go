package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"syscall"

	unsafe "unsafe"
)

const (
	magicval string = "070701"
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
	//Check     uint64
}

func (header CpioHeader) VerifyMagic() bool {
	return header.Magic == magicval
}

type RawCpioHeader struct {
	Magic     [6]byte
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
	_         [8]byte
	//Check     [8]byte
}

func (rawheader RawCpioHeader) ToCpioHeader() CpioHeader {
	var cpioheader CpioHeader

	cpioheader.Magic = string(rawheader.Magic[:])
	//cpioheader.Inode = binary.LittleEndian.Uint64(rawheader.Inode[:])
	cpioheader.Inode = byteArrayToInt(rawheader.Inode[:])
	cpioheader.Mode = byteArrayToInt(rawheader.Mode[:])
	cpioheader.Uid = byteArrayToInt(rawheader.Uid[:])
	cpioheader.Gid = byteArrayToInt(rawheader.Gid[:])
	cpioheader.Nlink = byteArrayToInt(rawheader.Nlink[:])
	cpioheader.Mtime = byteArrayToInt(rawheader.Mtime[:])
	cpioheader.FileSize = byteArrayToInt(rawheader.FileSize[:])
	cpioheader.DevMajor = byteArrayToInt(rawheader.DevMajor[:])
	cpioheader.DevMinor = byteArrayToInt(rawheader.DevMinor[:])
	cpioheader.RDevMajor = byteArrayToInt(rawheader.RDevMajor[:])
	cpioheader.RDevMinor = byteArrayToInt(rawheader.RDevMinor[:])
	cpioheader.NameSize = byteArrayToInt(rawheader.NameSize[:])
	//cpioheader.Check = byteArrayToInt(rawheader.Check[:])

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

		_, err := fd.Seek(int64(nread), os.SEEK_CUR)
		check(err)

		header := raw_header.ToCpioHeader()
		if !header.VerifyMagic() {
			fmt.Fprintf(os.Stderr, "invalid magc number: %s\n", string(raw_header.Magic[:]))
			os.Exit(1)
		}

		namebuf := make([]byte, header.NameSize)
		err = binary.Read(fd, binary.LittleEndian, namebuf)
		check(err)

		/*
			_, err = fd.Seek(int64(header.NameSize), os.SEEK_CUR)
			check(err)
			nread += int(header.NameSize)
		*/

		fmt.Printf("%+v\n", header)
		fmt.Printf("%+v\n", string(namebuf[:]))
		break
	}
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

func byteArrayToInt(bytes []byte) uint64 {
	s := string(bytes[:])
	i, err := strconv.ParseUint(s, 16, 64)
	check(err)
	return i
}
