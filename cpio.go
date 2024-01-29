package gocpio

import (
	"encoding/binary"
	"fmt"
	"io"
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

	filemode := cpiomember.getMode()
	fd, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, filemode)
	if err != nil {
		createFilePath(filepath)
		fd, err = os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, filemode)
		check(err)
	}
	defer fd.Close()

	_, err = fd.Write(cpiomember.data)
	check(err)
}

func (cpiomember CpioMember) isTrailer() bool {
	return cpiomember.name == trailer
}

func (cpiomember CpioMember) getMode() os.FileMode {
	return os.FileMode(cpiomember.header.Mode & 0777)
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

// TODO: Code duplicate, and also very ugly Refactor later.
func ParseCpio(path string) Cpio {
	br := newBinaryReader(path, binary.LittleEndian)
	var cpio Cpio
	for {
		cpio_member, err := nextCpioEntry(&br)
		if err == io.EOF {
			break
		}
		check(err)

		if cpio_member.isTrailer() {
			break
		}

		// there are multile cpio entries that share the same data
		// there will be multiple headers, but only one copy of the data
		if cpio_member.header.FileSize == 0 && !cpio_member.IsDir() {
			tmp_cpio_slice := []CpioMember{cpio_member}
			for {
				cpio_member, err := nextCpioEntry(&br)
				if err == io.EOF {
					break
				}
				check(err)

				tmp_cpio_slice = append(tmp_cpio_slice, cpio_member)
				if cpio_member.header.FileSize != 0 {
					break
				}
			}

			last := tmp_cpio_slice[len(tmp_cpio_slice)-1]
			for _, ent := range tmp_cpio_slice {
				if &ent == &last {
					cpio.members = append(cpio.members, last)
					break
				}

				ent.data = last.data
				cpio.members = append(cpio.members, ent)
			}
			continue
		}
		cpio.members = append(cpio.members, cpio_member)
	}

	return cpio
}

// consume the next entry in the cpio file
// this will move the offset within the BinaryReader forward as a side effect
func nextCpioEntry(br *BinaryReader) (CpioMember, error) {
	// extract and parse header
	var cpio_member CpioMember

	var raw_header RawCpioHeader
	if _, err := br.Read(&raw_header); err != nil {
		return CpioMember{}, err
	}

	if !raw_header.verifyMagic() {
		fmt.Fprintf(os.Stderr, "invalid file format or magic number: %s\n", string(raw_header.Magic[:]))
		os.Exit(1)
	}

	header := raw_header.ToCpioHeader()
	cpio_member.header = header

	namebuf := make([]byte, header.NameSize-1)
	if _, err := br.Read(namebuf); err != nil {
		return CpioMember{}, err
	}
	cpio_member.name = string(namebuf[:])

	br.Skip(0)

	// extract data
	if header.FileSize != 0 {
		cpio_member.data = make([]byte, header.FileSize)
		if _, err := br.Read(cpio_member.data); err != nil {
			return CpioMember{}, err
		}
	}
	br.Skip(0)

	return cpio_member, nil
}
