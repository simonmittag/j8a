package j8a

import (
	"bytes"
	"fmt"
	"testing"
)

//test pool allocation of zipper
func TestBrotliEncoder(t *testing.T) {
	//run small loop to ensure pool allocation works
	for i := 0; i <= 100; i++ {
		json := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutj8a\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
		br := *BrotliEncode(json)

		var want = 86
		if len(br) != want {
			t.Errorf("brotli compression not working, should be []byte size 86 for json provided")
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

func BenchmahkBrotliEncodeNBytes(b *testing.B, n int) {
	text := []byte(randSeq(n))
	for i := 0; i < b.N; i++ {
		BrotliEncode(text)
	}
	b.Logf("benchmark compressing %d bytes as brotli", len(text))
}

func BenchmahkBrotliDecodeNBytes(b *testing.B, n int) {
	br := *Gzip([]byte(randSeq(n)))
	for i := 0; i < b.N; i++ {
		BrotliDecode(br)
	}
	b.Logf("benchmark decompressing %d bytes as brotli", len(br))
}

func BenchmarkBrotliEncode128B(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<6)
}

func BenchmarkBrotliDecode128B(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<6)
}

func BenchmarkBrotliEncode1KB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<9)
}

func BenchmarkBrotliDecode1KB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<9)
}

func BenchmarkBrotliEncode64KB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<15)
}

func BenchmarkBrotliDecode64KB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<15)
}

func BenchmarkBrotliEncode128KB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<16)
}

func BenchmarkBrotliDecode128KB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<16)
}

func BenchmarkBrotliEncode1MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<19)
}

func BenchmarkBrotliDecode1MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<19)
}

func BenchmarkBrotliEncode2MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<20)
}

func BenchmarkBrotliDecode2MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<20)
}

func BenchmarkBrotliEncode4MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<21)
}

func BenchmarkBrotliDecode4MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<21)
}

func BenchmarkBrotliEncode8MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<22)
}

func BenchmarkBrotliDecode8MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<22)
}

func BenchmarkBrotliEncode16MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<23)
}

func BenchmarkBrotliDecode16MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<23)
}

func BenchmarkBrotliEncode32MB(b *testing.B) {
	BenchmahkBrotliEncodeNBytes(b,2<<24)
}

func BenchmarkBrotliDecode32MB(b *testing.B) {
	BenchmahkBrotliDecodeNBytes(b,2<<24)
}
