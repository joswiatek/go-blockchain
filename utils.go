package main

import "encoding/binary"

func IntToHex(n int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(n))
	return b
}
