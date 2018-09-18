package main

import (
	"bytes"
	"encoding/binary"
)

func byteUInt64(b []byte) (uint64, error) {
	var i uint64
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		return i, err
	}
	return i, nil
}

func byteUInt32(b []byte) (uint32, error) {
	var i uint32
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		return i, err
	}
	return i, nil
}

func byteUInt8(b []byte) (uint8, error) {
	var i uint8
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &i)
	if err != nil {
		return i, err
	}
	return i, nil
}
