package j8a

import (
	"bytes"
	"fmt"
	"github.com/andybalholm/brotli"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
)

func TestBrotliCompressionRatio(t *testing.T) {
	nums := []int{1, 2, 3}
	brotlis := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

	for _, i := range nums {
		b, _ := ioutil.ReadFile(fmt.Sprintf("./unit/example%d.json", i))
		for _, z := range brotlis {
			var buf bytes.Buffer
			w := brotli.NewWriterLevel(&buf, z)
			w.Write(b)
			w.Flush()
			w.Close()

			r := float32(buf.Len()) / float32(len(b))

			t.Logf("json size %d, compressed size %d, brotli level %d, compression ratio %v", len(b), buf.Len(), z, r)
		}
	}

}

//test pool allocation of zipper
func TestBrotliEncoder(t *testing.T) {
	//run small loop to ensure pool allocation works
	for i := 0; i <= 100; i++ {
		json := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutj8a\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
		br := *BrotliEncode(json)

		var want = 111
		if len(br) != want {
			t.Errorf("brotli compression not working, should be compressed []byte %d for data size %d provided, but got %d", want, len(json), len(br))
		}
	}
}

func TestBrotliDecoder(t *testing.T) {
	for i := 0; i <= 100; i++ {
		json := []byte(fmt.Sprintf(`{ "key":"value%d" }`, i))
		if c := bytes.Compare(json, *BrotliDecode(*BrotliEncode(json))); c != 0 {
			t.Error("brotli data is not equal to original")
		}
	}
}

func TestBrotliEncodeThenBrotliDecodePoolIntegrity(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i <= 100000; i++ {
		json := []byte(fmt.Sprintf(`{ "key":"value %v" }`, rand.Float64()*float64(i)))
		wg.Add(1)

		go func() {
			if c := bytes.Compare(json, *BrotliDecode(*BrotliEncode(json))); c != 0 {
				t.Error("brotli decoded data is not equal to original")
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmahkBrotliEncodeNBytes(b *testing.B, n int) {
	b.StopTimer()
	text := []byte(randSeq(n))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		BrotliEncode(text)
	}
}

func BenchmahkBrotliDecodeNBytes(b *testing.B, n int) {
	b.StopTimer()
	br := *BrotliEncode([]byte(randSeq(n)))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		BrotliDecode(br)
	}
}

func BenchmarkBrotliEncode128B(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<6)
}

func BenchmarkBrotliDecode128B(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<6)
}

func BenchmarkBrotliEncode1KB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<9)
}

func BenchmarkBrotliDecode1KB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<9)
}

func BenchmarkBrotliEncode64KB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<15)
}

func BenchmarkBrotliDecode64KB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<15)
}

func BenchmarkBrotliEncode128KB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<16)
}

func BenchmarkBrotliDecode128KB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<16)
}

func BenchmarkBrotliEncode1MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<19)
}

func BenchmarkBrotliDecode1MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<19)
}

func BenchmarkBrotliEncode2MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<20)
}

func BenchmarkBrotliDecode2MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<20)
}

func BenchmarkBrotliEncode4MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<21)
}

func BenchmarkBrotliDecode4MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<21)
}

func BenchmarkBrotliEncode8MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<22)
}

func BenchmarkBrotliDecode8MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<22)
}

func BenchmarkBrotliEncode16MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<23)
}

func BenchmarkBrotliDecode16MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<23)
}

func BenchmarkBrotliEncode32MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b, 2<<24)
}

func BenchmarkBrotliDecode32MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b, 2<<24)
}
