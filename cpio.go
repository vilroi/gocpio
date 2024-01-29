package gocpio

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

const (
	magicval string = "070701"
	trailer  string = "TRAILER!!!"
)

const (
	FILETYPE_MASK    uint64 = 0170000
	FILETYPE_SOCKET  uint64 = 0140000
	FILETYPE_SYMLINK uint64 = 0120000
	FILETYPE_REGULAR uint64 = 0100000
	FILETYPE_BLK     uint64 = 0060000
	FILETYPE_DIR     uint64 = 0040000
	FILETYPE_CHAR    uint64 = 0020000
	FILETYPE_FIFO    uint64 = 0010000
)

const (
	FILEPERM_SUID   uint64 = 0004000
	FILEPERM_SGID   uint64 = 0002000
	FILEPERM_STICKY uint64 = 0001000
	FILEPERM_PERM   uint64 = 0000777
)

type Cpio struct {
	members []CpioMember
}

func (cpio Cpio) ListFiles() {
	for _, member := range cpio.members {
		if member.name == trailer {
			break
		}
		fmt.Println(member.name)
	}
}

func (cpio *Cpio) Append(newmember CpioMember) {
	cpio.members = append(cpio.members, newmember)
}

func (cpio Cpio) ExtractFile(name string) {
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

func (cpio Cpio) ExtractAllFiles() {
	for _, member := range cpio.members {
		member.Dump()
	}
}

func (cpio Cpio) Test() {
	for _, member := range cpio.members {
		if member.IsRegular() {
			fmt.Println(member.name)
		}
	}
}

type CpioMember struct {
	header CpioHeader
	name   string
	data   []byte
}

func (cpiomember CpioMember) Dump() {
	filepath := cpiomember.name
	if filepath == "." {
		return
	}

	filepath = "./" + filepath

	if cpiomember.IsDir() {
		err := os.MkdirAll(filepath, 0755)
		check(err)
		return
	}

	if cpiomember.IsSymlink() {
		target := string(cpiomember.data[:])

		err := os.Symlink(target, cpiomember.name)
		check(err)
		return
	}

	fd, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
	check(err)

	_, err = fd.Write(cpiomember.data)
	check(err)
}

func (cpiomember CpioMember) IsSocket() bool {
	return FILETYPE_SOCKET == (cpiomember.header.Mode & FILETYPE_MASK)
}

func (cpiomember CpioMember) IsSymlink() bool {
	return FILETYPE_SYMLINK == (cpiomember.header.Mode & FILETYPE_MASK)
}

func (cpiomember CpioMember) IsRegular() bool {
	return FILETYPE_REGULAR == (cpiomember.header.Mode & FILETYPE_MASK)
}

func (cpiomember CpioMember) IsBlock() bool {
	return FILETYPE_BLK == (cpiomember.header.Mode & FILETYPE_MASK)
}

func (cpiomember CpioMember) IsDir() bool {
	return FILETYPE_DIR == (cpiomember.header.Mode & FILETYPE_MASK)
}

func (cpiomember CpioMember) IsChar() bool {
	return FILETYPE_CHAR == (cpiomember.header.Mode & FILETYPE_MASK)
}

func (cpiomember CpioMember) IsFifo() bool {
	return FILETYPE_FIFO == (cpiomember.header.Mode & FILETYPE_MASK)
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

func (rawheader RawCpioHeader) verifyMagic() bool {
	return string(rawheader.Magic[:]) == magicval
}

func ParseCpio(path string) Cpio {
	br := newBinaryReader(path, binary.LittleEndian)
	info := br.Stat()

	var cpio Cpio
	for nread := 0; nread < int(info.Size()); {
		var cpio_member CpioMember

		var raw_header RawCpioHeader
		nread += br.Read(&raw_header)
		if !raw_header.verifyMagic() {
			fmt.Fprintf(os.Stderr, "invalid file format or magic number: %s\n", string(raw_header.Magic[:]))
			os.Exit(1)
		}

		header := raw_header.ToCpioHeader()
		cpio_member.header = header

		namebuf := make([]byte, header.NameSize-1)
		nread += br.Read(namebuf)
		cpio_member.name = string(namebuf[:])

		// EOF
		if cpio_member.name == trailer {
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
