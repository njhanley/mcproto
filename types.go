package protocol

import (
	"encoding/binary"
	"math"

	"github.com/pkg/errors"
)

var (
	errBufTooSmall   = errors.New("buf too small")
	errValueTooLarge = errors.New("value too large")
	errStringTooLong = errors.New("string too long")
)

func getString(buf []byte) (s string, n int, err error) {
	length, m, err := getVarInt(buf)
	if n += m; err != nil {
		return s, n, err
	}
	if length > math.MaxInt16 {
		return s, n, errors.WithStack(errStringTooLong)
	}

	if len(buf) < n+int(length) {
		return s, n, errors.WithStack(errBufTooSmall)
	}
	s = string(buf[n : n+int(length)])
	n += int(length)

	return s, n, nil
}

func putString(buf []byte, s string) (n int, err error) {
	length := len(s)
	if length > math.MaxInt16 {
		return n, errors.WithStack(errStringTooLong)
	}

	m, err := putVarInt(buf, int32(length))
	if n += m; err != nil {
		return n, err
	}

	m = copy(buf[n:], s)
	if n += m; m < len(s) {
		return n, errors.WithStack(errBufTooSmall)
	}

	return n, nil
}

type position struct {
	x int32 // size=26, offset=38
	y int16 // size=12, offset=26
	z int32 // size=26, offset=0
}

//func getField(n uint64, size, offset uint) int64 {
//	return int64(n<<(64-(size+offset))) >> (64 - size)
//}

//func putField(n int64, size, offset uint) uint64 {
//	return uint64(n) << (64 - size) >> (64 - (size + offset))
//}

func getPosition(buf []byte) (p position, n int, err error) {
	if len(buf) < 8 {
		return p, 0, errors.WithStack(errBufTooSmall)
	}
	v := binary.BigEndian.Uint64(buf)
	p.x = int32(int64(v) >> 38)
	p.y = int16(int64(v<<26) >> 52)
	p.z = int32(int64(v<<38) >> 38)
	return p, 8, nil
}

func putPosition(buf []byte, p position) (int, error) {
	if len(buf) < 8 {
		return 0, errors.WithStack(errBufTooSmall)
	}
	var v uint64
	v |= uint64(p.x) << 38
	v |= uint64(p.y) << 52 >> 26
	v |= uint64(p.z) << 38 >> 38
	return 8, nil
}
