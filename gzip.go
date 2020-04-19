package jabba

import (
	"bytes"
	"compress/gzip"
	"sync"

	"github.com/rs/zerolog/log"
)

var gzipMagicBytes = []byte{0x1f, 0x8b}

var zipPool = sync.Pool{
	New: func() interface{} {
		var buf bytes.Buffer
		return gzip.NewWriter(&buf)
	},
}

// Gzip a []byte
func Gzip(input []byte) []byte {
	wrt, _ := zipPool.Get().(*gzip.Writer)
	buf := &bytes.Buffer{}
	wrt.Reset(buf)

	wrt.Write(input)
	wrt.Close()
	defer zipPool.Put(wrt)

	enc := buf.Bytes()
	log.Trace().Msgf("byte buffer size %d", len(enc))
	return enc
}
