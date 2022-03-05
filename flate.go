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

var deflatePool = sync.Pool{
	New: func() interface{} {
		var buf bytes.Buffer
		w, _ := flate.NewWriter(&buf, flateLevel)
		return w
	},
}

var inflatePool = sync.Pool{
	New: func() interface{} {
		buf := bytes.NewBuffer(flateEmpty)
		r := flate.NewReader(buf)
		return r
	},
}

//Deflate compress a []byte
func Deflate(input []byte) *[]byte {
	wrt, _ := deflatePool.Get().(*flate.Writer)
	buf := &bytes.Buffer{}
	wrt.Reset(buf)

	_, _ = wrt.Write(input)
	_ = wrt.Close()
	defer deflatePool.Put(wrt)

	enc := buf.Bytes()
	return &enc
}

//Inflate a []byte, the inverse operation of compression with deflate.
func Inflate(input []byte) *[]byte {
	rd, _ := inflatePool.Get().(io.ReadCloser)
	buf := bytes.NewBuffer(input)
	_ = rd.(flate.Resetter).Reset(buf, nil)

	dec, _ := ioutil.ReadAll(rd)
	_ = rd.Close()
	defer inflatePool.Put(rd)

	return &dec
}
