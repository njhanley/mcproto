package protocol

import (
	"math"

	"github.com/pkg/errors"
)

type packet struct {
	id   int32
	data []byte
}

func getPacket(buf []byte) (p packet, n int, err error) {
	length, m, err := getVarInt(buf)
	if n += m; err != nil {
		return p, n, err
	}

	id, m, err := getVarInt(buf[n:])
	if n += m; err != nil {
		return p, n, err
	}

	l := int(length) - m
	if len(buf) < n+l {
		return p, n, errors.WithStack(errBufTooSmall)
	}
	data := make([]byte, l)
	n += copy(data, buf[n:n+l])

	return packet{id, data}, n, nil
}

func putPacket(buf []byte, p packet) (n int, err error) {
	length := lenVarInt(p.id) + len(p.data)
	if length > math.MaxInt32 {
		return n, errors.New("packet is too large")
	}

	m, err := putVarInt(buf, int32(length))
	if n += m; err != nil {
		return n, err
	}

	m, err = putVarInt(buf[n:], p.id)
	if n += m; err != nil {
		return n, err
	}

	m = copy(buf[n:], p.data)
	if n += m; m < len(p.data) {
		return n, errors.WithStack(errBufTooSmall)
	}

	return n, nil
}
