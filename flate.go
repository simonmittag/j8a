package j8a

import (
	"bytes"
	"github.com/klauspost/compress/flate"
	"io"
	"io/ioutil"
	"sync"
)

const flateLevel int = 1
var flateEmpty = []byte{0}

var flatePool = sync.Pool{
	New: func() interface{} {
		var buf bytes.Buffer
		w, _ := flate.NewWriter(&buf, flateLevel)
		return w
	},
}

var deflatePool = sync.Pool{
	New: func() interface{} {
		buf := bytes.NewBuffer(flateEmpty)
		r := flate.NewReader(buf)
		return r
	},
}

//Flate compress a []byte
func Flate(input []byte) *[]byte {
	wrt, _ := flatePool.Get().(*flate.Writer)
	buf := &bytes.Buffer{}
	wrt.Reset(buf)

	_, _ = wrt.Write(input)
	_ = wrt.Close()
	defer flatePool.Put(wrt)

	enc := buf.Bytes()
	return &enc
}

// Deflate a []byte
func Deflate(input []byte) *[]byte {
	rd, _ := deflatePool.Get().(io.ReadCloser)
	buf := bytes.NewBuffer(input)
	_ = rd.(flate.Resetter).Reset(buf, nil)

	dec, _ := ioutil.ReadAll(rd)
	_ = rd.Close()
	defer deflatePool.Put(rd)

	return &dec
}
