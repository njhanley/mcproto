package mcproto

import (
	"bytes"
	"math"
	"math/rand"
	"reflect"
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
		{[]byte{}, 0, 0, errBufTooSmall},                                     // empty buffer
		{[]byte{0xff}, 0, 1, errBufTooSmall},                                 // incomplete buffer
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
		{[]byte{}, 0, 0, errBufTooSmall},     // empty buffer
		{[]byte{0xff}, 0, 1, errBufTooSmall}, // incomplete buffer
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, 0, 10, errValueTooLarge}, // value overflow
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 0, 10, errValueTooLarge},       // incomplete buffer with value overflow
	}
)

func TestGetVarInt(t *testing.T) {
	for _, c := range varIntCases {
		v, n, err := GetVarInt(c.bytes)
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
		v, n, err := GetVarLong(c.bytes)
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
		n, err := PutVarInt(buf, c.value)
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
		n, err := PutVarLong(buf, c.value)
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
		n := LenVarInt(c.value)
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
		n := LenVarLong(c.value)
		if n != c.length {
			t.Errorf("have: %#v, want: %#v, got: %#v", c.value, c.length, n)
		}
	}
}

const (
	benchmarkSeed   = 1
	benchmarkDatums = 256
)

func BenchmarkGetVarInt(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([][]byte, benchmarkDatums)
	for i := range data {
		buf := make([]byte, maxIntBytes)
		v := int32(rand.Uint32())
		n, err := PutVarInt(buf, v)
		if err != nil {
			b.Fatalf("failed to create benchmark data: have: %#v, got: (%#v, %#v, %#v)",
				v,
				buf[:n], n, err)
		}
		data[i] = buf[:n]
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetVarInt(data[i%benchmarkDatums])
	}
}

func BenchmarkGetVarLong(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([][]byte, benchmarkDatums)
	for i := range data {
		buf := make([]byte, maxLongBytes)
		v := int64(rand.Uint64())
		n, err := PutVarLong(buf, v)
		if err != nil {
			b.Fatalf("failed to create benchmark data: have: %#v, got: (%#v, %#v, %#v)",
				v,
				buf[:n], n, err)
		}
		data[i] = buf[:n]
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetVarLong(data[i%benchmarkDatums])
	}
}

func BenchmarkPutVarInt(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([]int32, benchmarkDatums)
	for i := range data {
		data[i] = int32(rand.Uint32())
	}
	buf := make([]byte, maxIntBytes)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		PutVarInt(buf, data[i%benchmarkDatums])
	}
}

func BenchmarkPutVarLong(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([]int64, benchmarkDatums)
	for i := range data {
		data[i] = int64(rand.Uint64())
	}
	buf := make([]byte, maxLongBytes)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		PutVarLong(buf, data[i%benchmarkDatums])
	}
}

func BenchmarkLenVarInt(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([]int32, benchmarkDatums)
	for i := range data {
		data[i] = int32(rand.Uint32())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		LenVarInt(data[i%benchmarkDatums])
	}
}

func BenchmarkLenVarLong(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([]int64, benchmarkDatums)
	for i := range data {
		data[i] = int64(rand.Uint64())
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		LenVarLong(data[i%benchmarkDatums])
	}
}

var stringCases = []struct {
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

func TestGetString(t *testing.T) {
	for _, c := range stringCases {
		s, n, err := GetString(c.bytes)
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
		n, err := PutString(buf, c.value)
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
		buf := make([]byte, LenVarInt(int32(length))+length)
		n, _ := PutVarInt(buf, int32(length))
		for j := range buf[n:] {
			buf[n+j] = byte(' ' + rand.Intn(95)) // random ASCII printable character
		}
		data[i] = buf
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetString(data[i%benchmarkDatums])
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
		PutString(buf, data[i%benchmarkDatums])
	}
}

var positionCases = []struct {
	bytes  []byte
	value  Position
	length int
	err    error
}{
	{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, Position{0, 0, 0}, 8, nil}, // zero value
	{[]byte{0x00, 0x00, 0x40, 0x40, 0xab, 0xff, 0xfd, 0xff}, Position{257, 42, -513}, 8, nil},
	{[]byte{}, Position{}, 0, errBufTooSmall},                                         // empty buffer
	{[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, Position{}, 7, errBufTooSmall}, // incomplete buffer
}

func TestGetPosition(t *testing.T) {
	for _, c := range positionCases {
		p, n, err := GetPosition(c.bytes)
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
		n, err := PutPosition(buf, c.value)
		if n != c.length || errors.Cause(err) != c.err || bytes.Compare(buf[:n], c.bytes) != 0 {
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)", c.value, c.length, c.err, c.bytes, n, err, buf[:n])
		}
	}
}

func BenchmarkGetPosition(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([][]byte, benchmarkDatums)
	for i := range data {
		p := Position{
			x: int32(rand.Intn(1 << 26)),
			y: int16(rand.Intn(1 << 12)),
			z: int32(rand.Intn(1 << 26)),
		}
		buf := make([]byte, positionLen)
		if n, err := PutPosition(buf, p); err != nil {
			b.Fatalf("failed to create benchmark data: have: %#v, got: (%#v, %#v, %#v)",
				p,
				n, err, buf[:n])
		}
		data[i] = buf
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetPosition(data[i%benchmarkDatums])
	}
}

func BenchmarkPutPosition(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([]Position, benchmarkDatums)
	for i := range data {
		data[i].x = int32(rand.Intn(1 << 26))
		data[i].y = int16(rand.Intn(1 << 12))
		data[i].z = int32(rand.Intn(1 << 26))
	}
	buf := make([]byte, positionLen)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		PutPosition(buf, data[i%benchmarkDatums])
	}
}

var packetCases = []struct {
	bytes  []byte
	value  Packet
	length int
	err    error
}{
	{[]byte{0x01, 0x00}, Packet{ID: 0x00, Data: []byte{}}, 2, nil},
	{[]byte{0x02, 0x00, 0x00}, Packet{ID: 0x00, Data: []byte{0x00}}, 3, nil},
	{[]byte{0x05, 0x04, 0x03, 0x02, 0x01, 0x00}, Packet{ID: 0x04, Data: []byte{0x03, 0x02, 0x01, 0x00}}, 6, nil},
	{[]byte{0x02, 0x80, 0x01}, Packet{ID: 0x80, Data: []byte{}}, 3, nil},
	{[]byte{0x80, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, Packet{ID: 0x00, Data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}, 130, nil},
	{[]byte{}, Packet{}, 0, errBufTooSmall},     // empty buffer
	{[]byte{0x01}, Packet{}, 1, errBufTooSmall}, // incomplete buffer
}

func TestGetPacket(t *testing.T) {
	for _, c := range packetCases {
		p, n, err := GetPacket(c.bytes)
		if !reflect.DeepEqual(p, c.value) || n != c.length || errors.Cause(err) != c.err {
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)", c.bytes, c.value, c.length, c.err, p, n, err)
		}
	}
}

func TestPutPacket(t *testing.T) {
	for _, c := range packetCases {
		if c.err != nil {
			continue // skip invalid cases
		}
		buf := make([]byte, c.length)
		n, err := PutPacket(buf, c.value)
		if n != c.length || errors.Cause(err) != c.err || bytes.Compare(buf[:n], c.bytes) != 0 {
			t.Errorf("have: %#v, want: (%#v, %#v, %#v), got: (%#v, %#v, %#v)", c.value, c.length, c.err, c.bytes, n, err, buf[:n])
		}
	}
}

const (
	benchmarkMaxPacketID      = 256
	benchmarkMaxPacketDataLen = 512
)

func BenchmarkGetPacket(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([][]byte, benchmarkDatums)
	for i := range data {
		p := Packet{
			ID:   int32(rand.Intn(benchmarkMaxPacketID)),
			Data: make([]byte, rand.Intn(benchmarkMaxPacketDataLen)),
		}
		rand.Read(p.Data)

		buf := make([]byte, 2*maxIntBytes+benchmarkMaxPacketDataLen)
		n, err := PutPacket(buf, p)
		if err != nil {
			b.Fatalf("failed to create benchmark data: have: %#v, got: (%#v, %#v, %#v)",
				p,
				n, err, buf[:n])
		}

		data[i] = buf[:n]
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetPacket(data[i%benchmarkDatums])
	}
}

func BenchmarkPutPacket(b *testing.B) {
	rand.Seed(benchmarkSeed)
	data := make([]Packet, benchmarkDatums)
	for i := range data {
		p := Packet{
			ID:   int32(rand.Intn(benchmarkMaxPacketID)),
			Data: make([]byte, rand.Intn(benchmarkMaxPacketDataLen)),
		}
		rand.Read(p.Data)
		data[i] = p
	}
	buf := make([]byte, 2*maxIntBytes+benchmarkMaxPacketDataLen)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		PutPacket(buf, data[i%benchmarkDatums])
	}
}
