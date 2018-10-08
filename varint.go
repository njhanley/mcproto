package protocol

import (
	"math/bits"

	"github.com/pkg/errors"
)

const (
	cbits = 7          // bits per chunk
	msb   = 1 << cbits // most significant bit
	cmask = msb - 1    // chunk mask

	maxIntBytes  = 5
	maxLongBytes = 10
)

func getVarN(buf []byte, maxBytes int) (v uint64, n int, err error) {
	for i := 0; i < len(buf) && i < maxBytes; i++ {
		v |= uint64(buf[i]&cmask) << uint(i*cbits)
		if buf[i]&msb == 0 {
			return v, i + 1, nil
		}
	}
	if len(buf) < maxBytes {
		return 0, maxBytes, errors.WithStack(errBufTooSmall)
	}
	return 0, maxBytes, errors.WithStack(errValueTooLarge)
}

func getVarInt(buf []byte) (v int32, n int, err error) {
	_v, n, err := getVarN(buf, maxIntBytes)
	return int32(_v), n, err
}

func getVarLong(buf []byte) (v int64, n int, err error) {
	_v, n, err := getVarN(buf, maxLongBytes)
	return int64(_v), n, err
}

func putVarN(buf []byte, v uint64, maxBytes int) (int, error) {
	for i := 0; i < len(buf) && i < maxBytes; i++ {
		if v&^cmask == 0 {
			buf[i] = byte(v)
			return i + 1, nil
		}
		buf[i] = byte(v | msb)
		v >>= cbits
	}
	if len(buf) < maxBytes {
		return maxBytes, errors.WithStack(errBufTooSmall)
	}
	return maxBytes, errors.WithStack(errValueTooLarge)
}

func putVarInt(buf []byte, v int32) (int, error) {
	// convert to uint32 before uint64 to avoid sign extension
	return putVarN(buf, uint64(uint32(v)), maxIntBytes)
}

func putVarLong(buf []byte, v int64) (int, error) {
	return putVarN(buf, uint64(v), maxLongBytes)
}

func lenVarN(v uint64) int {
	return 1 + ((bits.Len64(v) - 1) / cbits)
}

func lenVarInt(v int32) int {
	// convert to uint32 before uint64 to avoid sign extension
	return lenVarN(uint64(uint32(v)))
}

func lenVarLong(v int64) int {
	return lenVarN(uint64(v))
}
