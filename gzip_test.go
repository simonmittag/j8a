package jabba

import (
	"bytes"
	"testing"
)

//TestDefaultDownstreamReadTimeout
func TestGzipper(t *testing.T) {
	//run small loop to ensure pool allocation works
	for i:=0;i<=10;i++ {
		json := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutJabba\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
		zipped := Gzip(json)

		if c := bytes.Compare(zipped[0:2], gzipMagicBytes); c != 0 {
			t.Errorf("gzip format not properly encoded, want %v, got %v", gzipMagicBytes, zipped[0:2])
		}

		var want = [2]int{100, 120}
		if !(len(zipped) >= want[0] && len(zipped) <= want[1]) {
			t.Errorf("gzip compression not working")
		}
	}
}
