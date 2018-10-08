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
		{[]byte{0x00}, "", 1, nil},
		{[]byte{0x01, 0x30}, "0", 2, nil},
		{[]byte{0x0d, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21}, "hello, world!", 14, nil},
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
		buf := make([]byte, c.length)
		n, err := putString(buf, c.value)
		if n != c.length || errors.Cause(err) != c.err || bytes.Compare(buf[:n], c.bytes) != 0 {
			t.Errorf("have: %q, want: (%d, %v, [% #x]), got: (%d, %v, [% #x])", c.value, c.length, c.err, c.bytes, n, err, buf[:n])
		}
	}
}
