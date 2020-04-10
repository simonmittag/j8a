package server

import (
	"bytes"
	"compress/gzip"
)

// Gzip a []byte
func Gzip(input []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	w.Write(input)
	w.Close()
	return buf.Bytes()
}
