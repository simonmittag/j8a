package j8a

import (
	"bytes"
	"github.com/andybalholm/brotli"
	"io/ioutil"
	"sync"
)

const brotliLevel int = 1
var brotliEmpty = []byte{0}

var brotliEncPool = sync.Pool{
	New: func() interface{} {
		var buf bytes.Buffer
		return brotli.NewWriterLevel(&buf, brotliLevel)
	},
}

var brotliDecPool = sync.Pool{
	New: func() interface{} {
		buf := bytes.NewBuffer(brotliEmpty)
		r := brotli.NewReader(buf)
		return r
	},
}

// BrotliEncode encodes to brotli from byte array.
func BrotliEncode(input []byte) *[]byte {
	wrt, _ := brotliEncPool.Get().(*brotli.Writer)
	buf := &bytes.Buffer{}
	wrt.Reset(buf)

	_, _ = wrt.Write(input)
	_ = wrt.Close()
	defer brotliEncPool.Put(wrt)

	enc := buf.Bytes()
	return &enc
}

// BrotliDecode decodes a []byte from Brotli binary format
func BrotliDecode(input []byte) *[]byte {
	rd, _ := brotliDecPool.Get().(*brotli.Reader)
	buf := bytes.NewBuffer(input)
	_ = rd.Reset(buf)

	dec, _ := ioutil.ReadAll(rd)
	_ = rd.Reset(buf)
	defer brotliDecPool.Put(rd)

	return &dec
}
