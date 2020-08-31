package jabba

import (
	"github.com/hako/durafmt"
	"testing"
	"time"
)

func TestPDurationAsString(t *testing.T) {
	pd399 := PDuration(time.Hour * 24 * 399)
	got := pd399.AsString()
	want := "1 year 4 weeks 6 days"
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
	tlsLinks := parseTlsLinks(tlsConfig.Certificates[0])
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

func TestLogCertificateStats(t *testing.T) {
	tlsConfig := mockTlsConfig()
	tlsLinks := logCertificateStats(tlsConfig.Certificates[0])
	if len(tlsLinks) != 1 {
		t.Errorf("tls links parsed incorrectly")
	}
}
