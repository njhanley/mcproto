package protocol

import (
	"encoding/binary"
	"math"
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

var (
	errBufTooSmall   = errors.New("buf too small")
	errValueTooLarge = errors.New("value too large")
)

func getVarN(buf []byte, maxBytes int) (v uint64, n int, err error) {
	for i := 0; i < len(buf) && i < maxBytes; i++ {
		v |= uint64(buf[i]&cmask) << uint(i*cbits)
		if buf[i]&msb == 0 {
			return v, i + 1, nil
		}
	}
	if len(buf) < maxBytes {
		return 0, len(buf), errors.WithStack(errBufTooSmall)
	}
	return 0, maxBytes, errors.WithStack(errValueTooLarge)
}

func GetVarInt(buf []byte) (v int32, n int, err error) {
	_v, n, err := getVarN(buf, maxIntBytes)
	return int32(_v), n, err
}

func GetVarLong(buf []byte) (v int64, n int, err error) {
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
		return len(buf), errors.WithStack(errBufTooSmall)
	}
	return maxBytes, errors.WithStack(errValueTooLarge)
}

func PutVarInt(buf []byte, v int32) (int, error) {
	// convert to uint32 before uint64 to avoid sign extension
	return putVarN(buf, uint64(uint32(v)), maxIntBytes)
}

func PutVarLong(buf []byte, v int64) (int, error) {
	return putVarN(buf, uint64(v), maxLongBytes)
}

func lenVarN(v uint64) int {
	return 1 + ((bits.Len64(v) - 1) / cbits)
}

func LenVarInt(v int32) int {
	// convert to uint32 before uint64 to avoid sign extension
	return lenVarN(uint64(uint32(v)))
}

func LenVarLong(v int64) int {
	return lenVarN(uint64(v))
}

func GetString(buf []byte) (s string, n int, err error) {
	length, m, err := GetVarInt(buf)
	if n += m; err != nil {
		return s, n, err
	}
	if length > math.MaxInt16 {
		return s, n, errors.WithStack(errValueTooLarge)
	}

	if len(buf) < n+int(length) {
		return s, n, errors.WithStack(errBufTooSmall)
	}
	s = string(buf[n : n+int(length)])
	n += int(length)

	return s, n, nil
}

func PutString(buf []byte, s string) (n int, err error) {
	length := len(s)
	if length > math.MaxInt16 {
		return n, errors.WithStack(errValueTooLarge)
	}

	m, err := PutVarInt(buf, int32(length))
	if n += m; err != nil {
		return n, err
	}

	m = copy(buf[n:], s)
	if n += m; m < len(s) {
		return n, errors.WithStack(errBufTooSmall)
	}

	return n, nil
}

const positionLen = 8

type Position struct {
	x int32 // size=26, offset=38
	y int16 // size=12, offset=26
	z int32 // size=26, offset=0
}

func GetField(n uint64, size, offset uint) int64 {
	return int64(n<<(64-(size+offset))) >> (64 - size)
}

func PutField(n int64, size, offset uint) uint64 {
	return uint64(n) << (64 - size) >> (64 - (size + offset))
}

func GetPosition(buf []byte) (p Position, n int, err error) {
	if len(buf) < positionLen {
		return p, len(buf), errors.WithStack(errBufTooSmall)
	}
	v := binary.BigEndian.Uint64(buf)
	p.x = int32(GetField(v, 26, 38))
	p.y = int16(GetField(v, 12, 26))
	p.z = int32(GetField(v, 26, 0))
	return p, positionLen, nil
}

func PutPosition(buf []byte, p Position) (int, error) {
	if len(buf) < positionLen {
		return len(buf), errors.WithStack(errBufTooSmall)
	}
	var v uint64
	v |= PutField(int64(p.x), 26, 38)
	v |= PutField(int64(p.y), 12, 26)
	v |= PutField(int64(p.z), 26, 0)
	binary.BigEndian.PutUint64(buf, v)
	return positionLen, nil
}

type Packet struct {
	ID   int32
	Data []byte
}

func GetPacket(buf []byte) (p Packet, n int, err error) {
	length, m, err := GetVarInt(buf)
	if n += m; err != nil {
		return p, n, err
	}

	id, m, err := GetVarInt(buf[n:])
	if n += m; err != nil {
		return p, n, err
	}

	l := int(length) - m
	if len(buf) < n+l {
		return p, n, errors.WithStack(errBufTooSmall)
	}
	data := make([]byte, l)
	n += copy(data, buf[n:n+l])

	return Packet{id, data}, n, nil
}

func PutPacket(buf []byte, p Packet) (n int, err error) {
	length := LenVarInt(p.ID) + len(p.Data)
	if length > math.MaxInt32 {
		return n, errors.WithStack(errValueTooLarge)
	}

	m, err := PutVarInt(buf, int32(length))
	if n += m; err != nil {
		return n, err
	}

	m, err = PutVarInt(buf[n:], p.ID)
	if n += m; err != nil {
		return n, err
	}

	m = copy(buf[n:], p.Data)
	if n += m; m < len(p.Data) {
		return n, errors.WithStack(errBufTooSmall)
	}

	return n, nil
}
