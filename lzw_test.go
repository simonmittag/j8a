package j8a

import (
	"bytes"
	"compress/lzw"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestLzwCompressCompressionRatio(t *testing.T) {
	nums := []int{1, 2, 3}
	
	for _, i := range nums {
		b, _ := ioutil.ReadFile(fmt.Sprintf("./unit/example%d.json", i))
		
			var buf bytes.Buffer
			w := lzw.NewWriter(&buf, lzwO, litw)
			w.Write(b)
			w.Close()

			r := float32(buf.Len()) / float32(len(b))

			t.Logf("json size %d, compressed size %d, lzw compression ratio %v", len(b), buf.Len(), r)
	}

}

//test pool allocation of flate
func TestLzwCompress(t *testing.T) {
	//run small loop to ensure pool allocation works
	for i := 0; i <= 100; i++ {
		json := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutj8a\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
		cp := *Compress(json)

		var want = 108
		if len(cp) != want {
			t.Errorf("Lzw Compress compression not working, should be compressed []byte %d for data size %d provided, but got %d", want, len(json), len(cp))
		}
	}
}

func TestLzwDecompress(t *testing.T) {
	for i := 0; i <= 1000000; i++ {
		json := []byte(fmt.Sprintf(`{ "key":"value%d" }`, i))
		if c := bytes.Compare(json, *Decompress(*Compress(json))); c != 0 {
			t.Error("lzw decompressed data is not equal to original")
		}
	}
}

func BenchmahkLzwCompressNBytes(b *testing.B, n int) {
	b.StopTimer()
	text := []byte(randSeq(n))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Compress(text)
	}
}

func BenchmahkLzwDecompressNBytes(b *testing.B, n int) {
	b.StopTimer()
	br := *Compress([]byte(randSeq(n)))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Decompress(br)
	}
}

func BenchmarkLzwCompress128B(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<6)
}

func BenchmarkLzwDecompress128B(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<6)
}

func BenchmarkLzwCompress1KB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<9)
}

func BenchmarkLzwDecompress1KB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<9)
}

func BenchmarkLzwCompress64KB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<15)
}

func BenchmarkLzwDecompress64KB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<15)
}

func BenchmarkLzwCompress128KB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<16)
}

func BenchmarkLzwDecompress128KB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<16)
}

func BenchmarkLzwCompress1MB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<19)
}

func BenchmarkLzwDecompress1MB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<19)
}

func BenchmarkLzwCompress2MB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<20)
}

func BenchmarkLzwDecompress2MB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<20)
}

func BenchmarkLzwCompress4MB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<21)
}

func BenchmarkLzwDecompress4MB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<21)
}

func BenchmarkLzwCompress8MB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<22)
}

func BenchmarkLzwDecompress8MB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<22)
}

func BenchmarkLzwCompress16MB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<23)
}

func BenchmarkLzwDecompress16MB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<23)
}

func BenchmarkLzwCompress32MB(b *testing.B) {
	BenchmahkLzwCompressNBytes(b, 2<<24)
}

func BenchmarkLzwDecompress32MB(b *testing.B) {
	BenchmahkLzwDecompressNBytes(b, 2<<24)
}
