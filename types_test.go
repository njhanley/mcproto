package protocol

import (
	"bytes"
	"testing"

	"github.com/pkg/errors"
)

var (
	stringCases = []struct {
		bytes  []byte
		value  string
		length int
		err    error
	}{
		{[]byte{0x00}, "", 1, nil}, // empty string
		{[]byte{0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21}, "Hello, world!", 14, nil},    // ASCII string
		{[]byte{0x0e, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0xe4, 0xb8, 0x96, 0xe7, 0x95, 0x8c, 0x21}, "Hello, 世界!", 15, nil}, // UTF-8 string
		{[]byte{}, "", 0, errBufTooSmall},                   // empty buffer
		{[]byte{0x01}, "", 1, errBufTooSmall},               // incomplete buffer
		{[]byte{0xff, 0xff, 0x02}, "", 3, errValueTooLarge}, // string length > math.MaxInt16
	}
	positionCases = []struct {
		bytes  []byte
		value  position
		length int
		err    error
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, position{0, 0, 0}, 8, nil}, // zero value
		{[]byte{0x00, 0x00, 0x40, 0x40, 0xab, 0xff, 0xfd, 0xff}, position{257, 42, -513}, 8, nil},
		{[]byte{}, position{}, 0, errBufTooSmall},                                         // empty buffer
		{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, position{}, 7, errBufTooSmall}, // incomplete buffer
	}
)

func TestGetString(t *testing.T) {
	for _, c := range stringCases {
		s, n, err := getString(c.bytes)
		if s != c.value || n != c.length || errors.Cause(err) != c.err {
			t.Errorf("have: [% #x], want: (%q, %d, %v), got: (%q, %d, %v)", c.bytes, c.value, c.length, c.err, s, n, err)
		}
	}
}

func TestPutString(t *testing.T) {
	for _, c := range stringCases {
		if c.err != nil {
			continue // skip invalid cases
		}
		buf := make([]byte, c.length)
		n, err := putString(buf, c.value)
		if n != c.length || errors.Cause(err) != c.err || bytes.Compare(buf[:n], c.bytes) != 0 {
			t.Errorf("have: %q, want: (%d, %v, [% #x]), got: (%d, %v, [% #x])", c.value, c.length, c.err, c.bytes, n, err, buf[:n])
		}
	}
}

func TestGetPosition(t *testing.T) {
	for _, c := range positionCases {
		p, n, err := getPosition(c.bytes)
		if p != c.value || n != c.length || errors.Cause(err) != c.err {
			t.Errorf("have: %v, want: (%v, %v, %v), got: (%v, %v, %v)", c.bytes, c.value, c.length, c.err, p, n, err)
		}
	}
}

func TestPutPosition(t *testing.T) {
	for _, c := range positionCases {
		if c.err != nil {
			continue // skip invalid cases
		}
		buf := make([]byte, c.length)
		n, err := putPosition(buf, c.value)
		if n != c.length || errors.Cause(err) != c.err || bytes.Compare(buf[:n], c.bytes) != 0 {
			t.Errorf("have: %v, want: (%v, %v, %v), got: (%v, %v, %v)", c.value, c.length, c.err, c.bytes, n, err, buf[:n])
		}
	}
}
