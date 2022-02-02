package j8a

import (
	"bytes"
	"fmt"
	"github.com/klauspost/compress/flate"
	"io/ioutil"
	"testing"
)

func TestFlateCompressionRatio(t *testing.T) {
	nums := []int{1, 2, 3}
	flates := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	for _, i := range nums {
		b, _ := ioutil.ReadFile(fmt.Sprintf("./unit/example%d.json", i))
		for _, z := range flates {
			var buf bytes.Buffer
			w, _ := flate.NewWriter(&buf, z)
			w.Write(b)
			w.Flush()
			w.Close()

			r := float32(buf.Len()) / float32(len(b))

			t.Logf("json size %d, compressed size %d, flate level %d, compression ratio %v", len(b), buf.Len(), z, r)
		}
	}

}

//test pool allocation of flate
func TestFlateEncoder(t *testing.T) {
	//run small loop to ensure pool allocation works
	for i := 0; i <= 100; i++ {
		json := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutj8a\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
		fl := *Flate(json)

		var want = 96
		if len(fl) != want {
			t.Errorf("flate compression not working, should be compressed []byte %d for data size %d provided, but got %d", want, len(json), len(fl))
		}
	}
}

func TestFlateDecoder(t *testing.T) {
	for i := 0; i <= 1000000; i++ {
		json := []byte(fmt.Sprintf(`{ "key":"value%d" }`, i))
		if c := bytes.Compare(json, *Deflate(*Flate(json))); c != 0 {
			t.Error("deflated data is not equal to original")
		}
	}
}

func BenchmahkFlateNBytes(b *testing.B, n int) {
	b.StopTimer()
	text := []byte(randSeq(n))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Flate(text)
	}
}

func BenchmahkDeflateNBytes(b *testing.B, n int) {
	b.StopTimer()
	br := *Flate([]byte(randSeq(n)))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Deflate(br)
	}
}

func BenchmarkFlate128B(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<6)
}

func BenchmarkDeflate128B(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<6)
}

func BenchmarkFlate1KB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<9)
}

func BenchmarkDeflate1KB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<9)
}

func BenchmarkFlate64KB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<15)
}

func BenchmarkDeflate64KB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<15)
}

func BenchmarkFlate128KB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<16)
}

func BenchmarkDeflate128KB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<16)
}

func BenchmarkFlate1MB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<19)
}

func BenchmarkDeflate1MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<19)
}

func BenchmarkFlate2MB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<20)
}

func BenchmarkDeflate2MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<20)
}

func BenchmarkFlate4MB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<21)
}

func BenchmarkDeflate4MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<21)
}

func BenchmarkFlate8MB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<22)
}

func BenchmarkDeflate8MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<22)
}

func BenchmarkFlate16MB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<23)
}

func BenchmarkDeflate16MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<23)
}

func BenchmarkFlate32MB(b *testing.B) {
	BenchmahkFlateNBytes(b, 2<<24)
}

func BenchmarkDeflate32MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<24)
}
