package gocpio

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

type BinaryReader struct {
	fd        *os.File
	byteorder binary.ByteOrder
}

func newBinaryReader(path string, byteorder binary.ByteOrder) BinaryReader {
	fd, err := os.Open(path)
	check(err)

	return BinaryReader{
		fd:        fd,
		byteorder: byteorder,
	}
}

/*
@buf: the caller is responsible for passing a reference to the buffer
*/
func (binreader *BinaryReader) Read(buf any) int {
	err := binary.Read(binreader.fd, binreader.byteorder, buf)
	check(err)

	nread := int64(sizeof(buf))
	check(err)

	return int(nread)
}

func (binreader BinaryReader) Stat() os.FileInfo {
	info, err := binreader.fd.Stat()
	check(err)

	return info
}

func (binreader *BinaryReader) Skip(c byte) {
	tmp := make([]byte, 1)
	for {
		binreader.Read(tmp)
		if tmp[0] != 0 {
			binreader.SeekCur(-1)
			break
		}
	}
}

func (binreader *BinaryReader) SeekCur(n int64) {
	_, err := binreader.fd.Seek(n, os.SEEK_CUR)
	check(err)
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

func sizeof(data any) int {
	val := reflect.ValueOf(data)
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		return val.Len()
	case reflect.Pointer:
		return int(reflect.Indirect(val).Type().Size())
	default:
		fmt.Println("type not implemented:", reflect.TypeOf(data).Kind())
		os.Exit(1)
	}

	return 0
}
