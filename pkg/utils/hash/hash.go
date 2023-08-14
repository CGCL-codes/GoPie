package hash

import (
	"crypto/md5"
)

func Hash16(s string) uint16 {
	var res uint16
	b := []byte(s)
	hashb := md5.Sum(b)
	res += uint16(hashb[0])
	res += uint16(hashb[1] << 8)
	return res
}

func Hash32(s string) uint32 {
	var res uint32
	b := []byte(s)
	hashb := md5.Sum(b)
	res += uint32(hashb[0])
	res += uint32(hashb[1] << 8)
	res += uint32(hashb[2] << 16)
	res += uint32(hashb[3] << 24)
	return res
}

func Hash64(s string) uint64 {
	var res uint64
	b := []byte(s)
	hashb := md5.Sum(b)
	res += uint64(hashb[0])
	res += uint64(hashb[1] << 8)
	res += uint64(hashb[2] << 16)
	res += uint64(hashb[3] << 24)
	res += uint64(hashb[4] << 32)
	res += uint64(hashb[5] << 40)
	res += uint64(hashb[6] << 48)
	res += uint64(hashb[7] << 56)
	return res
}
