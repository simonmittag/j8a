package j8a

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"sync"

	"github.com/rs/zerolog/log"
)

var gzipMagicBytes = []byte{0x1f, 0x8b}
var gzipSmall = []byte{31,139,8,0,0,0,0,0,0,255,170,174,5,4,0,0,255,255,67,191,166,163,2,0,0,0}

var zipPool = sync.Pool{
	New: func() interface{} {
		var buf bytes.Buffer
		return gzip.NewWriter(&buf)
	},
}

var unzipPool = sync.Pool{
	New: func() interface{} {
		buf := bytes.NewBuffer(gzipSmall)
		r, _ := gzip.NewReader(buf)
		return r
	},
}

// Gzip a []byte
func Gzip(input []byte) []byte {
	wrt, _ := zipPool.Get().(*gzip.Writer)
	buf := &bytes.Buffer{}
	wrt.Reset(buf)

	_, _ = wrt.Write(input)
	_ = wrt.Close()
	defer zipPool.Put(wrt)

	enc := buf.Bytes()
	log.Trace().Msgf("zipped byte buffer size %d", len(enc))
	return enc
}

// Gunzip a []byte
func Gunzip(input []byte) []byte {
	rd, _ := unzipPool.Get().(*gzip.Reader)
	buf := bytes.NewBuffer(input)
	_ = rd.Reset(buf)

	dec, _ := ioutil.ReadAll(rd)
	_ = rd.Close()
	defer unzipPool.Put(rd)

	log.Trace().Msgf("unzipped byte buffer size %d", len(dec))
	return dec
}