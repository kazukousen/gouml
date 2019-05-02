package gouml

import (
	"bytes"
	"compress/zlib"
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

func compress(src string) []byte {
	buf := bytesPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bytesPool.Put(buf)
	}()

	zw, _ := zlib.NewWriterLevel(buf, zlib.BestCompression)
	zw.Write([]byte(src))
	zw.Close()
	return buf.Bytes()
}

var bytesPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

func encode64(input []byte) string {
	var buf bytes.Buffer
	length := len(input)
	// padding
	for i := 0; i < 3-length%3; i++ {
		input = append(input, byte(0))
	}

	for i := 0; i < length; i += 3 {
		cs := append3bytes(input[i], input[i+1], input[i+2])
		for _, c := range cs {
			buf.WriteByte(byte(chars[c]))
		}
	}
	return buf.String()
}

const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"

func append3bytes(b1, b2, b3 byte) []byte {
	c1 := (b1 >> 2)
	c2 := ((b1 & 0x3) << 4) | (b2 >> 4)
	c3 := ((b2 & 0xF) << 2) | (b3 >> 6)
	c4 := b3 & 0x3F
	return []byte{c1, c2, c3, c4}
}
