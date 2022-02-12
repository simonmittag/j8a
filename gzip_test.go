package j8a

import (
	"bytes"
	"fmt"
	"github.com/klauspost/compress/gzip"
	"io/ioutil"
	"math/rand"
	"testing"
)

//test pool allocation of zipper
func TestGzipper(t *testing.T) {
	//run small loop to ensure pool allocation works
	for i := 0; i <= 10; i++ {
		json := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutj8a\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
		zipped := *Gzip(json)

		if c := bytes.Compare(zipped[0:2], gzipMagicBytes); c != 0 {
			t.Errorf("gzip format not properly encoded, want %v, got %v", gzipMagicBytes, zipped[0:2])
		}

		var want = [2]int{100, 120}
		if !(len(zipped) >= want[0] && len(zipped) <= want[1]) {
			t.Errorf("gzip compression not working")
		}
	}
}

func TestGzipThenUnzip(t *testing.T) {
	for i := 0; i <= 100; i++ {
		json := []byte(fmt.Sprintf(`{ "key":"value%d" }`, i))
		if c := bytes.Compare(json, *Gunzip(*Gzip(json))); c != 0 {
			t.Error("unzipped data is not equal to original")
		}
	}
}

func TestGzipCompressionRatio(t *testing.T) {
	nums := []int{1, 2, 3}
	zips := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	for _, i := range nums {
		b, _ := ioutil.ReadFile(fmt.Sprintf("./unit/example%d.json", i))
		for _, z := range zips {
			var buf bytes.Buffer
			w, _ := gzip.NewWriterLevel(&buf, z)
			w.Write(b)
			w.Flush()
			w.Close()

			r := float32(buf.Len()) / float32(len(b))

			t.Logf("json size %d, compressed size %d, gzip level %d, compression ratio %v", len(b), buf.Len(), z, r)
		}
	}

}

func BenchmahkGzipNBytes(b *testing.B, n int) {
	b.StopTimer()
	text := []byte(randSeq(n))
	//var res *[]byte
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Gzip(text)
	}
	//r := float32(binary.Size(*res))/float32(binary.Size(text))
	//b.Logf("compression ratio gzip level%d/identity: %g", gzipLevel, r)
}

func BenchmahkGunzipNBytes(b *testing.B, n int) {
	b.StopTimer()
	zipped := *Gzip([]byte(randSeq(n)))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Gunzip(zipped)
	}
}

func BenchmarkGzip128B(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<6)
}

func BenchmarkGunzip128B(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<6)
}

func BenchmarkGzip1KB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<9)
}

func BenchmarkGunzip1KB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<9)
}

func BenchmarkGzip64KB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<15)
}

func BenchmarkGunzip64KB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<15)
}

func BenchmarkGzip128KB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<16)
}

func BenchmarkGunzip128KB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<16)
}

func BenchmarkGzip1MB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<19)
}

func BenchmarkGunzip1MB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<19)
}

func BenchmarkGzip2MB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<20)
}

func BenchmarkGunzip2MB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<20)
}

func BenchmarkGzip4MB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<21)
}

func BenchmarkGunzip4MB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<21)
}

func BenchmarkGzip8MB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<22)
}

func BenchmarkGunzip8MB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<22)
}

func BenchmarkGzip16MB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<23)
}

func BenchmarkGunzip16MB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<23)
}

func BenchmarkGzip32MB(b *testing.B) {
	BenchmahkGzipNBytes(b, 2<<24)
}

func BenchmarkGunzip32MB(b *testing.B) {
	BenchmahkGunzipNBytes(b, 2<<24)
}

var letters = []rune("{}\":123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
