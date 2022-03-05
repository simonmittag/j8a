package j8a

import (
	"bytes"
	"fmt"
	"github.com/klauspost/compress/flate"
	"io/ioutil"
	"math/rand"
	"sync"
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
func TestDeflateEncoder(t *testing.T) {
	//run small loop to ensure pool allocation works
	for i := 0; i <= 100; i++ {
		json := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutj8a\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
		fl := *Deflate(json)

		var want = 96
		if len(fl) != want {
			t.Errorf("flate compression not working, should be compressed []byte %d for data size %d provided, but got %d", want, len(json), len(fl))
		}
	}
}

func TestDeflateThenInflateStringIntegrity(t *testing.T) {
	want := fmt.Sprintf(`{ "key":"value %v" }`, rand.Float64())
	got := *Inflate(*Deflate([]byte(want)))
	if c := bytes.Compare([]byte(want), got); c != 0 {
		t.Error("data is not equal to original after deflate, then inflate")
	} else {
		t.Logf("normal. deflate, then inflate bytes identical")
	}
}

func TestDeflateThenInflatePoolIntegrity(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i <= 100000; i++ {
		json := []byte(fmt.Sprintf(`{ "key":"value %v" }`, rand.Float64()*float64(i)))
		wg.Add(1)

		go func() {
			if c := bytes.Compare(json, *Inflate(*Deflate(json))); c != 0 {
				t.Error("data is not equal to original after deflate, then inflate")
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmahkDeflateNBytes(b *testing.B, n int) {
	b.StopTimer()
	text := []byte(randSeq(n))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Deflate(text)
	}
}

func BenchmahkInflateNBytes(b *testing.B, n int) {
	b.StopTimer()
	br := *Deflate([]byte(randSeq(n)))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Inflate(br)
	}
}

func BenchmarkDeflate128B(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<6)
}

func BenchmarkInflate128B(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<6)
}

func BenchmarkDeflate1KB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<9)
}

func BenchmarkInflate1KB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<9)
}

func BenchmarkDeflate64KB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<15)
}

func BenchmarkInflate64KB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<15)
}

func BenchmarkDeflate128KB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<16)
}

func BenchmarkInflate128KB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<16)
}

func BenchmarkDeflate1MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<19)
}

func BenchmarkInflate1MB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<19)
}

func BenchmarkDeflate2MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<20)
}

func BenchmarkInflate2MB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<20)
}

func BenchmarkDeflate4MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<21)
}

func BenchmarkInflate4MB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<21)
}

func BenchmarkDeflate8MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<22)
}

func BenchmarkInflate8MB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<22)
}

func BenchmarkDeflate16MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<23)
}

func BenchmarkInflate16MB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<23)
}

func BenchmarkDeflate32MB(b *testing.B) {
	BenchmahkDeflateNBytes(b, 2<<24)
}

func BenchmarkInflate32MB(b *testing.B) {
	BenchmahkInflateNBytes(b, 2<<24)
}
