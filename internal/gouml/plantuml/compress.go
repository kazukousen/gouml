package plantuml

import (
	"bytes"
	"compress/flate"
	"strings"
	"sync"
)

// Compress ...
func Compress(src string) string {
	trimmed := strings.Replace(src, "\t", "", -1)
	compressed := compress(trimmed)
	converted := encode64(compressed)
	return converted
}
func encode64(input []byte) string {
	var buf bytes.Buffer
	length := len(input)

	for i := 0; i < length; i += 3 {
		var c []byte
		if i+2 == length {
			c = append3bytes(input[i], input[i+1], 0)
		} else if i+1 == length {
			c = append3bytes(input[i], 0, 0)
		} else {
			c = append3bytes(input[i], input[i+1], input[i+2])
		}
		buf.Write(c)
	}
	return buf.String()
}

func append3bytes(b1, b2, b3 byte) []byte {
	c1 := (b1 >> 2)
	c2 := ((b1 & 0x3) << 4) | (b2 >> 4)
	c3 := ((b2 & 0xF) << 2) | (b3 >> 6)
	c4 := b3 & 0x3F
	return []byte{
		encode6bit(c1 & 0x3F),
		encode6bit(c2 & 0x3F),
		encode6bit(c3 & 0x3F),
		encode6bit(c4 & 0x3F),
	}
}

func encode6bit(b byte) byte {
	if b < 10 {
		return byte(48 + b)
	}
	b -= 10
	if b < 26 {
		return byte(65 + b)
	}
	b -= 26
	if b < 26 {
		return byte(97 + b)
	}
	b -= 26
	if b == 0 {
		return ([]byte("-"))[0]
	}
	if b == 1 {
		return ([]byte("_"))[0]
	}
	return ([]byte("?"))[0]
}

func compress(src string) []byte {
	buf := bytesPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bytesPool.Put(buf)
	}()

	// use Deflate (RFC1951)
	zw, _ := flate.NewWriter(buf, flate.BestCompression)
	zw.Write([]byte(src))
	zw.Close()
	return buf.Bytes()
}

var bytesPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}
