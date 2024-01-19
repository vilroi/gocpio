package main

import (
	"encoding/binary"
	"os"
	"reflect"
	"strconv"
	unsafe "unsafe"
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
func (binreader BinaryReader) Read(buf any) int {
	err := binary.Read(binreader.fd, binreader.byteorder, buf)
	check(err)

	//nread := int64(unsafe.Sizeof(buf))
	nread := int64(sizeof(buf))
	binreader.fd.Seek(nread, os.SEEK_CUR)

	return int(nread)
}

func (binreader BinaryReader) Stat() os.FileInfo {
	info, err := binreader.fd.Stat()
	check(err)

	return info
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

func sizeof(x any) int {
	switch reflect.TypeOf(x).Kind() {
	case reflect.Slice:
		slice := reflect.ValueOf(x)
		return slice.Len()
	default:
		return int(unsafe.Sizeof(x))
	}
}
