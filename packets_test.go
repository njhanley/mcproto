package protocol

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

var (
	packetCases = []struct {
		bytes  []byte
		value  packet
		length int
		err    error
	}{
		{[]byte{0x01, 0x00}, packet{id: 0x00, data: []byte{}}, 2, nil},
		{[]byte{0x02, 0x00, 0x00}, packet{id: 0x00, data: []byte{0x00}}, 3, nil},
		{[]byte{0x05, 0x04, 0x03, 0x02, 0x01, 0x00}, packet{id: 0x04, data: []byte{0x03, 0x02, 0x01, 0x00}}, 6, nil},
		{[]byte{0x02, 0x80, 0x01}, packet{id: 0x80, data: []byte{}}, 3, nil},
		{[]byte{0x80, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, packet{id: 0x00, data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}, 130, nil},
	}
)

func TestGetPacket(t *testing.T) {
	for _, c := range packetCases {
		p, n, err := getPacket(c.bytes)
		if !reflect.DeepEqual(p, c.value) || n != c.length || errors.Cause(err) != c.err {
			t.Errorf("have: [% #x], want: (%v, %d, %v), got: (%v, %d, %v)", c.bytes, c.value, c.length, c.err, p, n, err)
		}
	}
}

func TestPutPacket(t *testing.T) {
	for _, c := range packetCases {
		buf := make([]byte, c.length)
		n, err := putPacket(buf, c.value)
		if n != c.length || errors.Cause(err) != c.err || bytes.Compare(buf[:n], c.bytes) != 0 {
			t.Errorf("have: %v, want: (%d, %v, [% #x]), got: (%d, %v, [% #x])", c.value, c.length, c.err, c.bytes, n, err, buf[:n])
		}
	}
}
