package protocol

import (
	"bytes"
	"math/rand"
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
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)", c.bytes, c.value, c.length, c.err, s, n, err)
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
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)", c.value, c.length, c.err, c.bytes, n, err, buf[:n])
		}
	}
}

func TestGetPosition(t *testing.T) {
	for _, c := range positionCases {
		p, n, err := getPosition(c.bytes)
		if p != c.value || n != c.length || errors.Cause(err) != c.err {
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)", c.bytes, c.value, c.length, c.err, p, n, err)
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
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)", c.value, c.length, c.err, c.bytes, n, err, buf[:n])
		}
	}
}

const benchmarkMaxStringLen = 256

func BenchmarkGetString(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([][]byte, benchmarkDatums)
	for i := range data {
		length := rand.Intn(benchmarkMaxStringLen)
		buf := make([]byte, lenVarInt(int32(length))+length)
		n, _ := putVarInt(buf, int32(length))
		for j := range buf[n:] {
			buf[n+j] = byte(' ' + rand.Intn(95)) // random ASCII printable character
		}
		data[i] = buf
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getString(data[i%benchmarkDatums])
	}
}

func BenchmarkPutString(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([]string, benchmarkDatums)
	for i := range data {
		b := make([]byte, rand.Intn(benchmarkMaxStringLen))
		for j := range b {
			b[j] = byte(' ' + rand.Intn(95)) // random ASCII printable character
		}
		data[i] = string(b)
	}
	buf := make([]byte, maxIntBytes+benchmarkMaxStringLen)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		putString(buf, data[i%benchmarkDatums])
	}
}

func BenchmarkGetPosition(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([][]byte, benchmarkDatums)
	for i := range data {
		p := position{
			x: int32(rand.Intn(1 << 26)),
			y: int16(rand.Intn(1 << 12)),
			z: int32(rand.Intn(1 << 26)),
		}
		buf := make([]byte, positionLen)
		if n, err := putPosition(buf, p); err != nil {
			b.Fatalf("failed to create benchmark data: have: %#v, got: (%#v, %#v, %#v)",
				p,
				n, err, buf[:n])
		}
		data[i] = buf
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getPosition(data[i%benchmarkDatums])
	}
}

func BenchmarkPutPosition(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([]position, benchmarkDatums)
	for i := range data {
		data[i].x = int32(rand.Intn(1 << 26))
		data[i].y = int16(rand.Intn(1 << 12))
		data[i].z = int32(rand.Intn(1 << 26))
	}
	buf := make([]byte, positionLen)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		putPosition(buf, data[i%benchmarkDatums])
	}
}
