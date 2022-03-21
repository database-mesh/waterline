package bpf

import (
	"encoding/binary"
	"unsafe"
)

func htons(num uint16) uint16 {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, num)
	return *(*uint16)(unsafe.Pointer(&b[0]))
}
