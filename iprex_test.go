package j8a

import (
	"fmt"
	"math/rand"
	"testing"
)

var ipv4s = []string{"127.0.0.1",
	"10.1.1.1",
	"192.168.0.0",
	"254.1.1.1"}

var ipv4scenarios = []string{"i'm an ip address %s",
	"ip %s with some other text",
	"ip%swith some other text",
	"http://%s",
	"http://%s:80",
	"https://%s",
	"https://%s:443",
	"%s.",
	",%s.",
	"%s,"}

var ipv6s = []string{"::1",
	"2406:2000:e4:1605::9001",
	"1200:0000:AB00:1234:0000:2552:7777:1313",
	"21DA:D3:0:2F3B:2AA:FF:FE28:9C5A",
	"2600::",
	"2001:4860:4860::8888",
	"::8844",
	"2606:2800:220:1:248:1893:25c8:1946",
	"2001:db8:3333:4444:CCCC:DDDD:EEEE:FFFF",
	"2001:0db8:0001:0000:0000:0ab9:C0A8:0102",
	"2001:db8:1c0:2:21::",
	"1::1234:5678",
	//embedded
	"::ffff:10.1.1.1",
	"::0:0:0:0:FFFF:10.1.1.1",
	"::0:0:0:FFFF:10.1.1.1",
	"::0:0:FFFF:10.1.1.1",
	"::0:ffff:10.1.1.1",
	"1:0:0:0:0:FFFF:10.1.1.1",
	"2001:db8:c000:221::",
	"2001:db8:1c0:2:21::",
	"2001:db8:122:c000:2:2100::",
	"2001:db8:122:3c0:0:221::",
	"2001:db8:122:344:c0:2:2100::",
	"2001:db8:122:344::192.0.2.33",
}

var ipv6scenarios = []string{"i'm an ip address %s",
	"ip %s with some other text",
	"ip%swith some other text",
	"http://[%s]",
	"https://[%s]",
	"https://[%s]:443",
	"%s.",
	"[%s]",
	"[%s]:80"}

func BenchmarkIprexUncompiledIPv6Extract(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipr := iprex{}
		ipr.extractAddr(fmt.Sprintf(ipv6scenarios[rand.Intn(len(ipv6scenarios))], ipv6s[rand.Intn(len(ipv6s))]))
	}
}

func BenchmarkIprexUnCompiledIPv4Extract(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ipr := iprex{}
		ipr.extractAddr(fmt.Sprintf(ipv4scenarios[rand.Intn(len(ipv4scenarios))], ipv4s[rand.Intn(len(ipv4s))]))
	}
}

func BenchmarkIprexCompiledIPv6Extract(b *testing.B) {
	ipr := iprex{}
	for i := 0; i < b.N; i++ {
		ipr.extractAddr(fmt.Sprintf(ipv6scenarios[rand.Intn(len(ipv6scenarios))], ipv6s[rand.Intn(len(ipv6s))]))
	}
}

func BenchmarkIprexCompiledIPv4Extract(b *testing.B) {
	ipr := iprex{}
	for i := 0; i < b.N; i++ {
		ipr.extractAddr(fmt.Sprintf(ipv4scenarios[rand.Intn(len(ipv4scenarios))], ipv4s[rand.Intn(len(ipv4s))]))
	}
}

func TestIprexIPv4Extract(t *testing.T) {
	ipr := iprex{}
	for _, want := range ipv4s {
		for _, sc := range ipv4scenarios {
			got := ipr.extractAddr(fmt.Sprintf(sc, want))
			if want != got {
				t.Errorf("extracted address does not match, want %s, got %s", want, got)
			}
		}
	}
}

func TestIprexIPv6Extract(t *testing.T) {
	ipr := iprex{}
	for _, want := range ipv6s {
		for _, sc := range ipv6scenarios {
			got := ipr.extractAddr(fmt.Sprintf(sc, want))
			if want != got {
				t.Errorf("extracted address does not match, want %s, got %s", want, got)
			}
		}
	}
}
