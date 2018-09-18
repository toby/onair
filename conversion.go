package main

import (
	"bytes"
	"encoding/binary"
)

func byteUInt64(d []byte) (uint64, error) {
	var i uint64
	buf := bytes.NewReader(d)
	err := binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		return i, err
	}
	return i, nil
}

func byteUInt32(d []byte) (uint32, error) {
	var i uint32
	buf := bytes.NewReader(d)
	err := binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		return i, err
	}
	return i, nil
}

func byteUInt8(d []byte) (uint8, error) {
	var i uint8
	buf := bytes.NewReader(d)
	err := binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		return i, err
	}
	return i, nil
}
