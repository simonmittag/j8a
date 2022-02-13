package j8a

import (
	"bytes"
	"compress/lzw"
	"io/ioutil"
	"sync"
)

const litw int = 8
const lzwO lzw.Order = 0
var lzwEmpty = []byte{0}

var compressPool = sync.Pool{
	New: func() interface{} {
		var buf bytes.Buffer
		w := lzw.NewWriter(&buf, lzwO, litw)
		return w
	},
}

var decompressPool = sync.Pool{
	New: func() interface{} {
		buf := bytes.NewBuffer(lzwEmpty)
		r := lzw.NewReader(buf, lzwO, litw)
		return r
	},
}

//Compress a []byte using pooled lzw writer
func Compress(input []byte) *[]byte {
	wrt, _ := compressPool.Get().(*lzw.Writer)
	buf := &bytes.Buffer{}
	wrt.Reset(buf, lzwO, litw)

	_, _ = wrt.Write(input)
	_ = wrt.Close()
	defer compressPool.Put(wrt)

	enc := buf.Bytes()
	return &enc
}

//Decompress a []byte using pooled lzw reader
func Decompress(input []byte) *[]byte {
	rd, _ := decompressPool.Get().(*lzw.Reader)
	buf := bytes.NewBuffer(input)
	rd.Reset(buf, lzwO, litw)

	dec, _ := ioutil.ReadAll(rd)
	_ = rd.Close()
	defer decompressPool.Put(rd)

	return &dec
}
