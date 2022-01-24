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
