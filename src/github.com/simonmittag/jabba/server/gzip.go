package server

import (
	"bytes"
	"compress/gzip"
	"sync"

	"github.com/rs/zerolog/log"
)

type zipper struct {
	buf *bytes.Buffer
	wrt *gzip.Writer
}

var zipperPool = sync.Pool{
	New: func() interface{} {
		var buf bytes.Buffer
		return &zipper{
			buf: &buf,
			wrt: gzip.NewWriter(&buf),
		}
	},
}

// Gzip a []byte
func Gzip(input []byte) []byte {
	zipper, _ := zipperPool.Get().(*zipper)
	// zipper.buf.Reset()
	// zipper.wrt.Reset(ioutil.Discard)

	zipper.wrt.Write(input)
	zipper.wrt.Close()
	defer zipperPool.Put(zipper)

	enc := zipper.buf.Bytes()
	log.Trace().Msgf("byte buffer size %d", len(enc))
	return enc
}
