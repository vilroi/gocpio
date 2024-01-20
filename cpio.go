package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	magicval string = "070701"
)

type Cpio struct {
	members []CpioMember
}

func (cpio Cpio) listFiles() {
	for _, member := range cpio.members {
		fmt.Println(member.name)
	}
}

func (cpio *Cpio) Append(newmember CpioMember) {
	cpio.members = append(cpio.members, newmember)
}

func (cpio Cpio) extractFile(name string) {
	var subset []CpioMember

	for _, member := range cpio.members {

		if strings.Contains(member.name, name) {
			if member.name == name {
				// found! Save data to disk and return.
				member.Dump()
				return
			}
			subset = append(subset, member)
		}
	}

	if len(subset) == 0 {
		fmt.Fprintf(os.Stderr, "file '%s' not found\n", name)
		os.Exit(1)
	}

	for _, member := range subset {
		member.Dump()
	}

}

type CpioMember struct {
	header CpioHeader
	name   string
	data   []byte
}

/*
TODO: as for now, dump the file in current directory
I will later implement the creation of relative file path

Also, for now I shall ignore the file perms
*/
func (cpiomember CpioMember) Dump() {
	//path := filepath.Dir(cpiomember.name)
	filename := filepath.Base(cpiomember.name)

	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	check(err)

	_, err = fd.Write(cpiomember.data)
	check(err)
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
	Check     [8]byte
}

func (rawheader RawCpioHeader) ToCpioHeader() CpioHeader {
	var cpioheader CpioHeader

	cpioheader.Magic = string(rawheader.Magic[:])
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

	cpio := parseCpio(os.Args[1])
	cpio.listFiles()
	cpio.extractFile("munch.c")

}

func parseCpio(path string) Cpio {
	br := newBinaryReader(path, binary.LittleEndian)
	info := br.Stat()

	var cpio Cpio
	for nread := 0; nread < int(info.Size()); {
		var cpio_member CpioMember

		var raw_header RawCpioHeader
		nread += br.Read(&raw_header)

		header := raw_header.ToCpioHeader()
		if !header.VerifyMagic() {
			fmt.Fprintf(os.Stderr, "invalid magic number: %s\n", string(raw_header.Magic[:]))
			os.Exit(1)
		}
		cpio_member.header = header

		namebuf := make([]byte, header.NameSize-1)
		nread += br.Read(namebuf)
		cpio_member.name = string(namebuf[:])

		if cpio_member.name == "TRAILER!!!" {
			cpio.Append(cpio_member)
			break
		}
		br.Skip(0)

		if header.FileSize != 0 {
			cpio_member.data = make([]byte, header.FileSize)
			nread += br.Read(cpio_member.data)
		}
		br.Skip(0)

		cpio.Append(cpio_member)
	}

	return cpio
}
