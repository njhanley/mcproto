package protocol

import (
	"bytes"
	"math"
	"math/rand"
	"testing"

	"github.com/pkg/errors"
)

var (
	varIntCases = []struct {
		bytes  []byte
		value  int32
		length int
		err    error
	}{
		// valid
		{[]byte{0x00}, 0, 1, nil},                                     // zero
		{[]byte{0x01}, 1, 1, nil},                                     // one
		{[]byte{0x7f}, 127, 1, nil},                                   // largest value of length 1
		{[]byte{0x80, 0x01}, 128, 2, nil},                             // smallest value of length 2
		{[]byte{0xff, 0xff, 0xff, 0xff, 0x07}, math.MaxInt32, 5, nil}, // largest positive value
		{[]byte{0xff, 0xff, 0xff, 0xff, 0x0f}, -1, 5, nil},            // negative one
		{[]byte{0x80, 0x80, 0x80, 0x80, 0x08}, math.MinInt32, 5, nil}, // largest negative value
		// invalid
		{[]byte{}, 0, 5, errBufTooSmall},                                     // empty buffer
		{[]byte{0xff}, 0, 5, errBufTooSmall},                                 // incomplete buffer
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0x0f}, 0, 5, errValueTooLarge}, // value overflow
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff}, 0, 5, errValueTooLarge},       // incomplete buffer with value overflow
	}
	varLongCases = []struct {
		bytes  []byte
		value  int64
		length int
		err    error
	}{
		// valid
		{[]byte{0x00}, 0, 1, nil},         // zero
		{[]byte{0x01}, 1, 1, nil},         // one
		{[]byte{0x7f}, 127, 1, nil},       // largest value of length 1
		{[]byte{0x80, 0x01}, 128, 2, nil}, // smallest value of length 2
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, math.MaxInt64, 9, nil},        // largest positive value
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, -1, 10, nil},            // negative one
		{[]byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, math.MinInt64, 10, nil}, // largest negative number
		// invalid
		{[]byte{}, 0, 10, errBufTooSmall},     // empty buffer
		{[]byte{0xff}, 0, 10, errBufTooSmall}, // incomplete buffer
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, 0, 10, errValueTooLarge}, // value overflow
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, 10, errValueTooLarge},       // incomplete buffer with value overflow
	}
)

func TestGetVarInt(t *testing.T) {
	for _, c := range varIntCases {
		v, n, err := getVarInt(c.bytes)
		if v != c.value || n != c.length || errors.Cause(err) != c.err {
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)",
				c.bytes,
				c.value, c.length, c.err,
				v, n, err)
		}
	}
}

func TestGetVarLong(t *testing.T) {
	for _, c := range varLongCases {
		v, n, err := getVarLong(c.bytes)
		if v != c.value || n != c.length || errors.Cause(err) != c.err {
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)",
				c.bytes,
				c.value, c.length, c.err,
				v, n, err)
		}
	}
}

func TestPutVarInt(t *testing.T) {
	buf := make([]byte, maxIntBytes)
	for _, c := range varIntCases {
		if c.err != nil {
			continue // skip invalid cases
		}
		n, err := putVarInt(buf, c.value)
		if n != c.length || errors.Cause(err) != c.err || bytes.Compare(buf[:n], c.bytes) != 0 {
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)",
				c.value,
				c.bytes, c.length, c.err,
				buf[:n], n, err)
		}
	}
}

func TestPutVarLong(t *testing.T) {
	buf := make([]byte, maxLongBytes)
	for _, c := range varLongCases {
		if c.err != nil {
			continue // skip invalid cases
		}
		n, err := putVarLong(buf, c.value)
		if n != c.length || errors.Cause(err) != c.err || bytes.Compare(buf[:n], c.bytes) != 0 {
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)",
				c.value,
				c.bytes, c.length, c.err,
				buf[:n], n, err)
		}
	}
}

func TestLenVarInt(t *testing.T) {
	for _, c := range varIntCases {
		if c.err != nil {
			continue // skip invalid cases
		}
		n := lenVarInt(c.value)
		if n != c.length {
			t.Errorf("have: %#v, want: %#v, got: %#v", c.value, c.length, n)
		}
	}
}

func TestLenVarLong(t *testing.T) {
	for _, c := range varLongCases {
		if c.err != nil {
			continue // skip invalid cases
		}
		n := lenVarLong(c.value)
		if n != c.length {
			t.Errorf("have: %#v, want: %#v, got: %#v", c.value, c.length, n)
		}
	}
}

const (
	seed   = 1
	datums = 256
)

func BenchmarkGetVarInt(b *testing.B) {
	rand.Seed(seed)
	data := make([][]byte, datums)
	for i := range data {
		buf := make([]byte, maxIntBytes)
		v := int32(rand.Uint32())
		n, err := putVarInt(buf, v)
		if err != nil {
			b.Fatalf("failed to create benchmark data: have: %#v, got: (%#v, %#v, %#v)",
				v,
				buf[:n], n, err)
		}
		data[i] = buf[:n]
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getVarInt(data[i%datums])
	}
}

func BenchmarkGetVarLong(b *testing.B) {
	rand.Seed(seed)
	data := make([][]byte, datums)
	for i := range data {
		buf := make([]byte, maxLongBytes)
		v := int64(rand.Uint64())
		n, err := putVarLong(buf, v)
		if err != nil {
			b.Fatalf("failed to create benchmark data: have: %#v, got: (%#v, %#v, %#v)",
				v,
				buf[:n], n, err)
		}
		data[i] = buf[:n]
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getVarLong(data[i%datums])
	}
}

func BenchmarkPutVarInt(b *testing.B) {
	rand.Seed(seed)
	data := make([]int32, datums)
	for i := range data {
		data[i] = int32(rand.Uint32())
	}
	buf := make([]byte, maxIntBytes)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		putVarInt(buf, data[i%datums])
	}
}

func BenchmarkPutVarLong(b *testing.B) {
	rand.Seed(seed)
	data := make([]int64, datums)
	for i := range data {
		data[i] = int64(rand.Uint64())
	}
	buf := make([]byte, maxLongBytes)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		putVarLong(buf, data[i%datums])
	}
}

func BenchmarkLenVarInt(b *testing.B) {
	rand.Seed(seed)
	data := make([]int32, datums)
	for i := range data {
		data[i] = int32(rand.Uint32())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lenVarInt(data[i%datums])
	}
}

func BenchmarkLenVarLong(b *testing.B) {
	rand.Seed(seed)
	data := make([]int64, datums)
	for i := range data {
		data[i] = int64(rand.Uint64())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lenVarLong(data[i%datums])
	}
}
