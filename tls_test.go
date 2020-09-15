package j8a

import (
	"crypto/x509"
	"github.com/hako/durafmt"
	"testing"
	"time"
)

func TestPDurationAsString(t *testing.T) {
	pd399 := PDuration(time.Hour * 24 * 399)
	got := pd399.AsString()
	want := "1 year 4 weeks"
	if got != want {
		t.Errorf("durafmt parse error, want %s, got %s", want, got)
	}
}

func TestPDuration_AsDays(t *testing.T) {
	pd399 := PDuration(time.Hour * 24 * 399)
	got := pd399.AsDays()
	want := 399
	if got != want {
		t.Errorf("durafmt parse error, want %d, got %d", want, got)
	}
}

func TestBrowserExpiry_AsDays(t *testing.T) {
	got := new(TlsLink).browserExpiry().AsDays()
	want := 398
	if got != want {
		t.Errorf("wrong browser expiry days, want %d, got %d", want, got)
	}
}

func TestParseTlsLinks(t *testing.T) {
	tlsConfig := mockTlsConfig()
	c, _ := x509.ParseCertificate(tlsConfig.Certificates[0].Certificate[0])
	tlsLinks := parseTlsLinks([]*x509.Certificate{c})

	logCertStats(tlsLinks)

	if len(tlsLinks) != 1 {
		t.Errorf("tls links parsed incorrectly")
	} else {
		if tlsLinks[0].isCA != false {
			t.Errorf("cert should not be a CA")
		}
		if tlsLinks[0].totalValidity.AsDuration().Seconds() != time.Duration(time.Second*352257299).Seconds() {
			t.Errorf("total validity should be %s", durafmt.Parse(tlsLinks[0].totalValidity.AsDuration()))
		}
	}
}

func TestCheckCertChain(t *testing.T) {
	tlsConfig := mockTlsConfig()
	verified, err := checkCertChain(tlsConfig.Certificates[0])
	if err != nil {
		t.Errorf("certificate chain with 1 TLS cert, 1 root cert not validated, cause: %s", err)
	} else {
		t.Logf("normal. certificate chain with 1 TLS cert, 1 root cert validated, length: %d", len(verified))
	}
}

func TestTlsHealthCheck(t *testing.T) {
	//this only needs to be covered for no runtime exceptions as it logs to console. no assertions.
	tlsHealthCheck(mockTlsConfig(), false)
}
