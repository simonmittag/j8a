package j8a

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	isd "github.com/jbenet/go-is-domain"
)

func TestMain(m *testing.M) {
	testSetup()
	code := m.Run()
	os.Exit(code)
}

func testSetup() {
	os.Setenv("LOGLEVEL", "TRACE")
	os.Setenv("LOGCOLOR", "true")
}

func TestTimeZones(t *testing.T) {
	tzs := []struct {
		tz    string
		valid bool
	}{
		{
			"Australia/Sydney", true,
		},
		{
			"Africa/Johannesburg", true,
		},
		{
			"UTC", true,
		},
	}

	config := new(Config)
	for _, tz := range tzs {
		config.TimeZone = tz.tz
		config.validateTimeZone()
	}

}

// TestDefaultDownstreamReadTimeout
func TestDefaultDownstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.ReadTimeoutSeconds
	want := 5
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

// TestDefaultDownstreamIdleTimeout
func TestDefaultDownstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.IdleTimeoutSeconds
	want := 5
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

// TestDefaultDownstreamRoundtripTimeout
func TestDefaultDownstreamRoundtripTimeout(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.RoundTripTimeoutSeconds
	want := 10
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

func TestDefaultDownstreamMaxBodyBytes(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.MaxBodyBytes
	want := int64(2097152)
	if got != want {
		t.Errorf("default dwn max body bytes got %d, want %d", got, want)
	}
}

// TestDefaultUpstreamSocketTimeout
func TestDefaultUpstreamSocketTimeout(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.SocketTimeoutSeconds
	want := 3
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

// TestDefaultUpstreamSocketTimeout
func TestDefaultUpstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.ReadTimeoutSeconds
	want := 10
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

// TestDefaultUpstreamIdleTimeout
func TestDefaultUpstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

// TestDefaultUpstreamConnectionPoolSize
func TestDefaultUpstreamConnectionPoolSize(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.PoolSize
	want := 32768
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

// TestDefaultUpstreamConnectionMaxAttempts
func TestDefaultUpstreamConnectionMaxAttempts(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.MaxAttempts
	want := 1
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

// TestDefaultPolicy
func TestDefaultPolicy(t *testing.T) {
	wantLabel := "default"

	config := new(Config).addDefaultPolicy()
	def := config.Policies[wantLabel]
	gotLabel := def[0].Label
	if gotLabel != wantLabel {
		t.Errorf("default policy label got %s, want %s", gotLabel, wantLabel)
	}

	gotWeight := def[0].Weight
	wantWeight := 1.0
	if gotWeight != wantWeight {
		t.Errorf("default policy weight got %f, want %fc", gotWeight, wantWeight)
	}
}

func TestParsePolicy(t *testing.T) {
	configJson := []byte("{\"policies\": {\n\t\t\"ab\": [{\n\t\t\t\t\"label\": \"green\",\n\t\t\t\t\"weight\": 0.8\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"label\": \"blue\",\n\t\t\t\t\"weight\": 0.2\n\t\t\t}\n\t\t]\n\t}}")
	config := new(Config).parse(configJson)

	p := config.Policies

	gp0l := p["ab"][0].Label
	wp0l := "green"
	if gp0l != wp0l {
		t.Errorf("incorrectly parsed policy label, want %s, got %s", wp0l, gp0l)
	}

	gp0w := p["ab"][0].Weight
	wp0w := 0.8
	if gp0w != wp0w {
		t.Errorf("incorrectly parsed policy weight, want %f, got %f", wp0w, gp0w)
	}

	gp1l := p["ab"][1].Label
	wp1l := "blue"
	if gp1l != wp1l {
		t.Errorf("incorrectly parsed policy label, want %s, got %s", wp1l, gp1l)
	}

	gp1w := p["ab"][1].Weight
	wp1w := 0.2
	if gp1w != wp1w {
		t.Errorf("incorrectly parsed policy weight, want %f, got %f", wp1w, gp1w)
	}
}

// TestParseResource
func TestParseResource(t *testing.T) {
	configJson := []byte("{\n\t\"resources\": {\n\t\t\"customer\": [{\n\t\t\t\"labels\": [\n\t\t\t\t\"blue\"\n\t\t\t],\n\t\t\t\"url\": {\n\t\t\t\t\"scheme\": \"http\",\n\t\t\t\t\"host\": \"localhost\",\n\t\t\t\t\"port\": 8081\n\t\t\t}\n\t\t}]\n\t}\n}")
	config := new(Config).parse(configJson)
	config.reApplyResourceNames()

	customer := config.Resources["customer"]
	if customer[0].Name != "customer" {
		t.Error("resource name not re-applied after parsing server configuration. cannot identify upstream resource, mapping would fail")
	}

	if customer[0].Labels[0] != "blue" {
		t.Error("resource label not parsed, cannot perform upstream mapping")
	}

	wantURL := URL{"http", "localhost", 8081}
	gotURL := customer[0].URL
	if wantURL != gotURL {
		t.Errorf("resource url parsed incorrectly. want %s got %s", wantURL, gotURL)
	}
}

// TestParseConnection
func TestParseConnection(t *testing.T) {
	configJson := []byte("{\n\t\"connection\": {\n\t\t\"downstream\": {\n\t\t\t\"readTimeoutSeconds\": 120,\n\t\t\t\"roundTripTimeoutSeconds\": 240,\n\t\t\t\"idleTimeoutSeconds\": 30,\n\t\t\t\"port\": 8080,\n\t\t\t\"mode\": \"TLS\"\n\t\t},\n\t\t\"upstream\": {\n\t\t\t\"socketTimeoutSeconds\": 3,\n\t\t\t\"readTimeoutSeconds\": 120,\n\t\t\t\"idleTimeoutSeconds\": 120,\n\t\t\t\"maxAttempts\": 4,\n\t\t\t\"poolSize\": 1024\n\t\t}\n\t}\n}")
	config := new(Config).parse(configJson)

	for i := 0; i < 2; i++ {

		//this whole test runs twice. the first pass validates the parsing of config object. the 2nd pass
		//validates the setDefaultValues() method does not inadvertently overwrite it
		if i == 1 {
			config = config.setDefaultDownstreamParams().setDefaultUpstreamParams()
		}

		c := config.Connection
		wrts := 120
		grts := c.Downstream.ReadTimeoutSeconds
		if wrts != grts {
			t.Errorf("incorrectly parsed downstream readTimeoutSeconds, want %d, got %d", wrts, grts)
		}

		wrtts := 240
		grtts := c.Downstream.RoundTripTimeoutSeconds
		if wrtts != grtts {
			t.Errorf("incorrectly parsed downstream roundTripTimeoutSeconds, want %d, got %d", wrtts, grtts)
		}

		wits := 30
		gits := c.Downstream.IdleTimeoutSeconds
		if wits != gits {
			t.Errorf("incorrectly parsed downstream idleTimeoutSeconds, want %d, got %d", wits, gits)
		}

		wuits := 120
		guits := c.Upstream.IdleTimeoutSeconds
		if wuits != guits {
			t.Errorf("incorrectly parsed upstream idleTimeoutSeconds, want %d, got %d", wuits, guits)
		}

		wurts := 120
		gurts := c.Upstream.ReadTimeoutSeconds
		if wurts != gurts {
			t.Errorf("incorrectly parsed upstream readTimeoutSeconds, want %d, got %d", wurts, gurts)
		}

		wusts := 3
		gusts := c.Upstream.SocketTimeoutSeconds
		if wusts != gusts {
			t.Errorf("incorrectly parsed upstream socketTimeoutSeconds, want %d, got %d", wusts, gusts)
		}

		wups := 1024
		gups := c.Upstream.PoolSize
		if wups != gups {
			t.Errorf("incorrectly parsed upstream poolSize, want %d, got %d", wups, gups)
		}

		wuma := 4
		guma := c.Upstream.MaxAttempts
		if wuma != guma {
			t.Errorf("incorrectly parsed upstream maxAttempts, want %d, got %d", wuma, guma)
		}
	}
}

// TestParseRoute
func TestParseRoute(t *testing.T) {
	configJson := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutj8a\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
	config := new(Config).parse(configJson)

	customer := config.Routes[1]
	if customer.Path != "/customer" {
		t.Errorf("incorrectly parsed route path, want %s, got %s", "/customer", customer.Path)
	}

	if customer.Policy != "ab" {
		t.Errorf("incorrectly parsed route policy, want %s, got %s", "ab", customer.Policy)
	}

	if customer.Resource != "customer" {
		t.Errorf("incorrectly parsed route resource, want %s, got %s", "customer", customer.Resource)
	}
}

// TestValidateAcmeEmail
func TestValidateConfigHasHTTP(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
			},
		},
	}

	config = config.validateHTTPConfig()
}

func TestValidateConfigHasHttpAndTLS(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{
					Port: 80,
				},
				Tls: Tls{
					Port: 443,
				},
			},
		},
	}

	config = config.validateHTTPConfig()
}

func TestValidateConfigHasTLS(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Tls: Tls{
					Port: 443,
				},
			},
		},
	}

	config = config.validateHTTPConfig()
}

func TestValidateConfigNoHttpAndNoTLSFails(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("config should have panicked with no tls and no http")
		} else {
			t.Logf("normal config panic without http or tls port")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{},
		},
	}

	config = config.validateHTTPConfig()
}

// TestValidateAcmeEmail
func TestValidateAcmeEmail(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
}

// TestValidateAcmeEmail
func TestValidateAcmeGracePeriod30(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:         []string{"adyntest.com"},
						Provider:        "letsencrypt",
						Email:           "noreply@example.org",
						GracePeriodDays: 30,
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()

	want := 30
	got := config.Connection.Downstream.Tls.Acme.GracePeriodDays
	if want != got {
		t.Errorf("want grace period days %v, got %v", want, got)
	}
}

func TestValidateAcmeGracePeriod15(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:         []string{"adyntest.com"},
						Provider:        "letsencrypt",
						Email:           "noreply@example.org",
						GracePeriodDays: 15,
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()

	want := 15
	got := config.Connection.Downstream.Tls.Acme.GracePeriodDays
	if want != got {
		t.Errorf("want grace period days %v, got %v", want, got)
	}
}

func TestValidateDefaultAcmeGracePeriod(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()

	want := 30
	got := config.Connection.Downstream.Tls.Acme.GracePeriodDays
	if want != got {
		t.Errorf("want grace period days %v, got %v", want, got)
	}
}

func TestValidateAcmeGracePeriodFailsGreater30(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("config should have panicked with 31 days acme grace period")
		} else {
			t.Logf("normal config panic for 31 days acme grace period")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:         []string{"adyntest.com"},
						Provider:        "letsencrypt",
						Email:           "noreply@example.org",
						GracePeriodDays: 31,
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
}

// TestValidateAcmeDomainInvalidLeadingDotFails
func TestValidateAcmeDomainInvalidLeadingDotFails(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Logf("normal. config panic for invalid domain with supported provider")
		}
	}()

	acmeConfigWith(".test.com").validateAcmeConfig()
	t.Errorf("config did not panic for supported Acme provider but with missing domain")
}

// TestValidateAcmeDomainInvalidTrailingDotFails
func TestValidateAcmeDomainInvalidTrailingDotFails(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Logf("normal. config panic for invalid domain with supported provider")
		}
	}()

	acmeConfigWith("test.com.").validateAcmeConfig()
	t.Errorf("config did not panic for supported Acme provider but with missing domain")
}

// TestValidateAcmeDomainInvalidTrailingDotFails
// NOTE WE DO NOT SUPPORT WILDCART CERTS BECAUSE THEY CANNOT BE VERIFIED USING HTTP01 CHALLENGE ON LETSENCRYPT, SEE: https://letsencrypt.org/docs/faq/
func TestValidateAcmeDomainInvalidWildcartCertFails(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Logf("normal. config panic for invalid domain with supported provider")
		}
	}()

	acmeConfigWith("*.test.com").validateAcmeConfig()
	t.Errorf("config did not panic for supported Acme provider but with missing domain")
}

// TestValidateAcmeDomainValidSubdomainPasses
func TestValidateAcmeDomainValidSubdomainPasses(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("config did panic for valid subdomain and supported Acme provider")
		}
	}()

	acmeConfigWith("subdomain.test.com").validateAcmeConfig()
	t.Logf("normal. config did not panic for valid subdomain and supported Acme provider")
}

// TestValidateAcmeDomainMissingFails
func TestValidateAcmeDomainMissingFails(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Logf("normal. config panic for missing domain with supported provider")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Errorf("config did not panic for supported Acme provider but with missing domain")
}

// TestValidateAcmeProviderLetsencrypt
func TestValidateAcmeProviderLetsencrypt(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Logf("normal. no config panic for Acme provider letsencrypt")
}

// TestValidateAcmeProviderLetsencryptWithMultipleSubDomains
func TestValidateAcmeProviderLetsencryptWithMultipleSubdomains(t *testing.T) {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com", "api.adyntest.com"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Logf("normal. no config panic for Acme provider letsencrypt")
}

// TestValidateAcmeProviderLetsencryptFailsWithOneInvalidSubDomain
func TestValidateAcmeProviderLetsencryptFailsWithOneInvalidSubDomain(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("normal. config panic for illegal subdomain.")
		}
	}()
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com", "Iwannabeadomain"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Errorf("config did not panic for invalid subdomain")
}

// TestValidateAcmeProviderLetsencrypt
func TestValidateAcmeProviderPort80(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("config panic for correct config with port 80 for acme provider")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{
					Port: 80,
				},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Logf("normal. no config panic for Acme provider letsencrypt with port 80")
}

// TestValidateAcmeProviderFailsWithMissingPort80
func TestValidateAcmeProviderFailsWithMissingPort80(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("normal. config panic for missing port 80 with acme provider")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Error("no config panic for Acme provider without port 80 specified. should have panicked")
}

// TestValidateAcmeProviderFailsWithCertSpecified
func TestValidateAcmeProviderFailsWithCertSpecified(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("normal. config panic for extra cert specified")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
					Cert: "iwannabeacertwheni'mbig",
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Error("no config panic happened after superfluous cert specified but it should have")
}

// TestValidateAcmeProviderFailsWithKeySpecified
func TestValidateAcmeProviderFailsWithKeySpecified(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("normal. config panic for extra private key specified")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
					Key: "wheni'mbigIwannabeaprivatekey",
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Error("no config panic occurred after extra private key specified next to acme")
}

// TestValidateAcmeMissingProviderFails
func TestValidateAcmeMissingProviderFails(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("normal. config panic for missing ACME provider")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains: []string{"adyntest.com"},
						Email:   "noreply@example.org",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Errorf("config did not panic for missing provider")
}

// TestValidateAcmeMissingEmailFails
func TestValidateAcmeEmailFails(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("normal. config panic for missing ACME email")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Errorf("config did not panic for missing email")
}

// TestValidateAcmeInvalidEmailFails
func TestValidateAcmeInvalidEmailFails(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("normal. config panic for invalid ACME email")
		}
	}()

	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{"adyntest.com"},
						Provider: "letsencrypt",
						Email:    "invalidemail@",
					},
				},
			},
		},
	}

	config = config.validateAcmeConfig()
	t.Errorf("config did not panic for invalid email")
}

func TestSortRoutes(t *testing.T) {
	config := new(Config).readYmlFile("./j8acfg.yml").validateRoutes()

	customer := config.Routes[0]
	if customer.Path != "/path/med/1st" {
		t.Error("incorrectly sorted routes")
	}

	about := config.Routes[1]
	if about.Path != "/path/med/2" {
		t.Error("incorrectly sorted routes")
	}
}

func TestTlsInsecureSkipVerify(t *testing.T) {
	//truisms in golang json string to bool parsing
	TlsInsecureSkipVerify_V(t, "y", true)
	TlsInsecureSkipVerify_V(t, "yes", true)
	TlsInsecureSkipVerify_V(t, "Y", true)
	TlsInsecureSkipVerify_V(t, "Yes", true)
	TlsInsecureSkipVerify_V(t, "n", false)
	TlsInsecureSkipVerify_V(t, "no", false)
	TlsInsecureSkipVerify_V(t, "N", false)
	TlsInsecureSkipVerify_V(t, "No", false)
	TlsInsecureSkipVerify_V(t, "True", true)
	TlsInsecureSkipVerify_V(t, "true", true)
	TlsInsecureSkipVerify_V(t, "False", false)
	TlsInsecureSkipVerify_V(t, "false", false)
}

func TlsInsecureSkipVerify_V(t *testing.T, v string, want bool) {
	config := new(Config).parse([]byte(fmt.Sprintf("---\nconnection:\n  upstream:\n    tlsInsecureSkipVerify: %s\n", v)))
	got := config.Connection.Upstream.TlsInsecureSkipVerify
	if got != want {
		t.Errorf("tlsInsecureSkipVerify got %v, want %v", got, want)
	}
}

func TestTlsInsecureSkipVerify_Optional(t *testing.T) {
	config := new(Config).parse([]byte("---\nconnection:\n  upstream:\n"))
	got := config.Connection.Upstream.TlsInsecureSkipVerify
	want := false
	if got != want {
		t.Errorf("tlsInsecureSkipVerify not specified in config, got %v, want %v", got, want)
	}
}

// TestReadConfigFile
func TestReadConfigFile(t *testing.T) {
	config := new(Config).readYmlFile("./j8acfg.yml")
	if config.Routes == nil {
		t.Error("incorrectly parsed routes in config file")
	}
	if config.Policies == nil {
		t.Error("incorrectly parsed policies in config file")
	}
	if reflect.DeepEqual(config.Connection, *new(Connection)) {
		t.Error("incorrectly parsed connection in config file")
	}
	if config.Resources == nil {
		t.Error("incorrectly parsed resources in config file")
	}
}

func TestReApplyScheme(t *testing.T) {
	want := map[string]string{"http": "", "https": ""}
	config := new(Config).readYmlFile("./j8acfg.yml").reApplyResourceURLDefaults()

	for name := range config.Resources {
		resourceMappings := config.Resources[name]
		for _, resourceMapping := range resourceMappings {
			scheme := resourceMapping.URL.Scheme
			if _, ok := want[scheme]; !ok {
				t.Errorf("incorrectly reapplied scheme, want any of %v, got %v", want, scheme)
			}
		}
	}
}

func TestReApplyPort(t *testing.T) {
	config := new(Config).readYmlFile("./j8acfg.yml").reApplyResourceURLDefaults()
	for name := range config.Resources {
		resourceMappings := config.Resources[name]
		for _, resourceMapping := range resourceMappings {
			port := int(resourceMapping.URL.Port)
			if port == 0 {
				t.Error("incorrectly applied port, got 0")
			}
		}
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	os.Setenv("J8ACFG_YML", "---\nconnection:\n  downstream:\n    readTimeoutSeconds: 333\n    roundTripTimeoutSeconds: 20\n    idleTimeoutSeconds: 30\n    port: 8080\n    mode: HTTP\n    maxBodyBytes: 65535\n  upstream:\n    socketTimeoutSeconds: 3\n    readTimeoutSeconds: 30\n    idleTimeoutSeconds: 10\n    maxAttempts: 4\n    poolSize: 8\n    tlsInsecureSkipVerify: true\npolicies:\n  ab:\n    - label: green\n      weight: 0.8\n    - label: blue\n      weight: 0.2\nroutes:\n  - path: \"/todos\"\n    resource: jsonplaceholder\n  - path: \"/about\"\n    resource: about\n  - path: \"/mse6/some\"\n    resource: mse61\n  - path: \"/mse6/\"\n    resource: mse6\n    policy: ab\n  - path: \"/s01\"\n    resource: s01\n  - path: \"/s02\"\n    resource: s02\n  - path: \"/s03\"\n    resource: s03\n  - path: \"/s04\"\n    resource: s04\n  - path: \"/s05\"\n    resource: s05\n  - path: \"/s06\"\n    resource: s06\n  - path: \"/s07\"\n    resource: s07\n  - path: \"/s08\"\n    resource: s08\n  - path: \"/s09\"\n    resource: s09\n  - path: \"/s10\"\n    resource: s10\n  - path: \"/s11\"\n    resource: s11\n  - path: \"/s12\"\n    resource: s12\n  - path: \"/s13\"\n    resource: s13\n  - path: \"/s14\"\n    resource: s14\n  - path: \"/s15\"\n    resource: s15\n  - path: \"/s16\"\n    resource: s16\n  - path: \"/badip\"\n    resource: badip\n  - path: \"/baddns\"\n    resource: baddns\n  - path: \"/badremote\"\n    resource: badremote\n  - path: \"/badlocal\"\n    resource: badlocal\n  - path: /badssl\n    resource: badssl\nresources:\n  jsonplaceholder:\n    - url:\n        scheme: https\n        host: jsonplaceholder.typicode.com\n        port: 443\n  badssl:\n    - url:\n        scheme: https\n        host: localhost\n        port: 60101\n  badip:\n    - url:\n        scheme: http\n        host: 10.247.13.14\n        port: 29471\n  baddns:\n    - url:\n        scheme: http\n        host: kajsdkfj23848392sdfjsj332jkjkjdkshhhhimnotahost.com\n        port: 29471\n  badremote:\n    - url:\n        scheme: http\n        host: google.com\n        port: 29471\n  badlocal:\n    - url:\n        scheme: http\n        host: localhost\n        port: 29471\n  mse61:\n    - url:\n        scheme: 'http:'\n        host: localhost\n        port: 60083\n  mse6:\n    - labels:\n        - green\n      url:\n        scheme: http://\n        host: localhost\n        port: 60083\n    - labels:\n        - blue\n      url:\n        host: localhost\n        port: 60084\n  s01:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60085\n  s02:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60086\n  s03:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60087\n  s04:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60088\n  s05:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60089\n  s06:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60090\n  s07:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60091\n  s08:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60092\n  s09:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60093\n  s10:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60094\n  s11:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60095\n  s12:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60096\n  s13:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60097\n  s14:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60098\n  s15:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60099\n  s16:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60100")
	config := new(Config).readYmlEnv()
	if config.Connection.Downstream.ReadTimeoutSeconds != 333 {
		t.Error("config not loaded from ENV")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	config := new(Config).readYmlFile(DefaultConfigFile)
	if config.Connection.Downstream.ReadTimeoutSeconds != 3 {
		t.Error("config not loaded from file")
	}
}

func TestLoadConfig(t *testing.T) {
	ConfigFile = "./integration/j8a3.yml"
	config := new(Config).load()
	if config.Connection.Downstream.Tls.Port != 8443 {
		t.Error("config not loaded from load() function")
	}
}

func TestGetEnvFunction(t *testing.T) {
	// Rendering Template with placeholders
	os.Setenv("TEST_ENV_VAR", "VALUES")
	env_vars := envToMap()
	wantVar := "VALUES"
	gotvar := env_vars["TEST_ENV_VAR"]
	if wantVar != gotvar {
		t.Errorf("env var map, want %v, got %v", wantVar, gotvar)
	}

}

func TestRenderVariableTemplate(t *testing.T) {
	// Rendering Template with placeholders
	os.Setenv("PORT", "8082")
	ConfigFile = "./integration/templatej8a1.yml"
	config := new(Config).load()
	if config.Connection.Downstream.Http.Port != 8082 {
		t.Error("config not Parsed from renderTemplate() function")
	}
}
func TestRenderSecretVariableTemplate(t *testing.T) {
	// Rendering Template with placeholders
	PUBLIC_KEY := `-----BEGIN PUBLIC KEY-----
      MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuFrwC7xqek3lA7TkRMBr
      7koamTCE5DF0UxVPd0FbmloGTkkLLXW3R6fOxubi8O2PXk/tN+TfJZiOYswUE/+n
      gR7gEXLebosLtVdmbGraTGwtoGmpSe3FRr9ZmQu74pZsAzwqZVMqz6CINc7uvxTI
      Djd98ORUrnuxqgHE9Yz/uo2qvnaOgWIXKhkDkMqA8O0Fk/kaCfeeZQMN70OnCwIS
      +LPFE8uYGIdbaEIkjZfMxm/iNRENOV849vwOiOuWruCyp+YMqTVtcW49Q1mcZfyG
      T7B5GHWe7MtxqQNhf1m2Nvo1m/LvaLap/EM3684xOa6RexB1XdB8oegpMRygPx7o
      rwIDAQAB
      -----END PUBLIC KEY-----`

	os.Setenv("PUBLIC_KEY", PUBLIC_KEY)
	os.Setenv("PORT", "8080")

	ConfigFile = "./integration/templatej8a2.yml"
	config := new(Config).load().validateJwt()
	if config.Connection.Downstream.Http.Port != 8080 {
		t.Error("config not Parsed from renderTemplate() function")
	}
}
func TestLoadAndRenderConfigFromEnv(t *testing.T) {
	ConfigFile = ""
	os.Setenv("READ_TIME_OUT_SECONDS", "301")
	os.Setenv("J8ACFG_YML", "---\nconnection:\n  downstream:\n    readTimeoutSeconds: {{.READ_TIME_OUT_SECONDS}}\n    roundTripTimeoutSeconds: 20\n    idleTimeoutSeconds: 30\n    port: 8080\n    mode: HTTP\n    maxBodyBytes: 65535\n  upstream:\n    socketTimeoutSeconds: 3\n    readTimeoutSeconds: 30\n    idleTimeoutSeconds: 10\n    maxAttempts: 4\n    poolSize: 8\n    tlsInsecureSkipVerify: true\npolicies:\n  ab:\n    - label: green\n      weight: 0.8\n    - label: blue\n      weight: 0.2\nroutes:\n  - path: \"/todos\"\n    resource: jsonplaceholder\n  - path: \"/about\"\n    resource: about\n  - path: \"/mse6/some\"\n    resource: mse61\n  - path: \"/mse6/\"\n    resource: mse6\n    policy: ab\n  - path: \"/s01\"\n    resource: s01\n  - path: \"/s02\"\n    resource: s02\n  - path: \"/s03\"\n    resource: s03\n  - path: \"/s04\"\n    resource: s04\n  - path: \"/s05\"\n    resource: s05\n  - path: \"/s06\"\n    resource: s06\n  - path: \"/s07\"\n    resource: s07\n  - path: \"/s08\"\n    resource: s08\n  - path: \"/s09\"\n    resource: s09\n  - path: \"/s10\"\n    resource: s10\n  - path: \"/s11\"\n    resource: s11\n  - path: \"/s12\"\n    resource: s12\n  - path: \"/s13\"\n    resource: s13\n  - path: \"/s14\"\n    resource: s14\n  - path: \"/s15\"\n    resource: s15\n  - path: \"/s16\"\n    resource: s16\n  - path: \"/badip\"\n    resource: badip\n  - path: \"/baddns\"\n    resource: baddns\n  - path: \"/badremote\"\n    resource: badremote\n  - path: \"/badlocal\"\n    resource: badlocal\n  - path: /badssl\n    resource: badssl\nresources:\n  jsonplaceholder:\n    - url:\n        scheme: https\n        host: jsonplaceholder.typicode.com\n        port: 443\n  badssl:\n    - url:\n        scheme: https\n        host: localhost\n        port: 60101\n  badip:\n    - url:\n        scheme: http\n        host: 10.247.13.14\n        port: 29471\n  baddns:\n    - url:\n        scheme: http\n        host: kajsdkfj23848392sdfjsj332jkjkjdkshhhhimnotahost.com\n        port: 29471\n  badremote:\n    - url:\n        scheme: http\n        host: google.com\n        port: 29471\n  badlocal:\n    - url:\n        scheme: http\n        host: localhost\n        port: 29471\n  mse61:\n    - url:\n        scheme: 'http:'\n        host: localhost\n        port: 60083\n  mse6:\n    - labels:\n        - green\n      url:\n        scheme: http://\n        host: localhost\n        port: 60083\n    - labels:\n        - blue\n      url:\n        host: localhost\n        port: 60084\n  s01:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60085\n  s02:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60086\n  s03:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60087\n  s04:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60088\n  s05:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60089\n  s06:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60090\n  s07:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60091\n  s08:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60092\n  s09:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60093\n  s10:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60094\n  s11:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60095\n  s12:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60096\n  s13:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60097\n  s14:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60098\n  s15:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60099\n  s16:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60100")
	config := new(Config).load()
	if config.Connection.Downstream.ReadTimeoutSeconds != 301 {
		t.Error("config not Rendered from ENV")
	}
	os.Unsetenv("READ_TIME_OUT_SECONDS")
}
func TestLoadAndRenderCertificateFromEnv(t *testing.T) {
	os.Setenv("PORT", "9443")
	PRIVATE_KEY := "-----BEGIN PRIVATE KEY-----\n        MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCwJND2stNBiMiU\n        YXsQ6tMm7HwTnXhNSgC5DTGjU05Kyym2OYPIaFvwrv4D9OB1T/GU8xwl/bac0NDY\n        ymqmAcaUNy5e3dKZWtyxN5vLU9rzWDEqnbdrnUHzECgegfBdVJZ2IxaTcf8mO/92\n        1gvUvXzSd57Bxa8rboA1TCXNYboFbVBVRdc11Gh2bFDV6DZL+owEBIWPpCMKxZYm\n        IR+bPg9Cmw/yrkioJxfzkFSoY73X0mUT95cQCCT1i6HrX/nSff0H+w8LSZTmzq8f\n        RI8IG/PWDxKJdryOPeMJfFr/Rmxi0BWo97kGhE+8rarWxspslWvsSmaI50rEo8ei\n        dEdqzB4bAgMBAAECggEAc9nDFn7HM3MjeXQj3RyVhCRF9yC63xqtHwjufN1twQOe\n        i5uIcWcyETsHFtMYThAmdDDxcotMcBdnRS7cthK06QbiGMMMoJCCVoyciz674xE+\n        RSk2WjE0DwmxWV9dGAVqcIjjcFap2hvcCez+Gw4F6ueCIzBB5e7npCZRNqPwFWCY\n        /og06ypz/4LHXFNatvRJC3qhWFwFo1bVC1ycZmc5RQ4IHeQHzi6oCJhSRdCbM7Q9\n        7fQhmjtcw0pxvJTVV+XP7tTf9iDDwgi/Le2iEqNQ1D6c4+nYAGYj2D3919oUYnyv\n        tnznZ2GTibIyP6kl4L79ChRz0JGBzraKH7aJh9H1AQKBgQDS0xd8RNLJ5hkwPitt\n        x/4RNlobGGZqqvQiCKkaDvnc1E1eKR3rVlxH17/ccA/qs0vcoFDPbGyRa/OgD7p5\n        Ro3R+EPDFoFq1KMP4SRoGcDgrNKHQ3o06sUngmtUUP28G+DUQm46xW7cnQrvvNiK\n        f9MMfeNH+tAPtssZh0HNKJa9fwKBgQDV40peDqh9b4Ag+mfiHIJGuwTN8LMvKRqr\n        N5dVirl2BDYMwF6JflIwIwjBZq7ah3NsT5/Yd+nuY+ux/pO4iU5jMbTtoAOv2dc7\n        VKpqNTaQdJhta5OlOdBSP5iXj4siVCMIFL1jz8JtWuXX4hUtbliG6ICZTH2/5ivG\n        faPiOhAlZQKBgQCp941jnojiRSPhhP22UBpA/jS+y3kmXhTcq2bJn3FJ289ULon0\n        hXd4ZDRGIAJ1EYADqyv7TkppI0MStBt+UqdbtG/NBIPqAOxFjRmw47JgcHR6oKgR\n        qYSxSbAGFhW6Zi9ocPY1Y57xNZrvlKxvXIZl98gY69h6EsDDIAyoviRpOQKBgC0g\n        Rjlv+EZ2tt6+VhqTjzzjClF03ikuD+1dzjUDDrwCiXDJSWjS2P5E9fzv8CY0+7o3\n        Vm8yZY2hUUH9hycg+QPeoeCcqQp5+HoRE99SmM+DegFj+AOdHgGsX0Jiy6UTgUyc\n        K5UaaVfvHJ0emv85z72u4ir1w3YwVr4LFf+N5ogtAoGAHODQpVC7sg+nlbeSKsPf\n        RbULfOG4YD5pHszNM+nCjNWs00ofJoZOFA64qXwTIc4Vrh8JLiwAkXiTGYM2guv5\n        Qnp+HbFi/tAc+rQu4SGBaVIglnIj7jFNdgJOb68Vw/L9v2jW1Y8VoAC0eCRWpHud\n        GsMkN4GFOfQKoBI/aCXn4DM=\n        -----END PRIVATE KEY-----"
	os.Setenv("J8A_PRIVATE_KEY", PRIVATE_KEY)
	ConfigFile = "./integration/templatej8a3.yml"
	config := new(Config).load()

	if config.Connection.Downstream.Tls.Port != 9443 {
		t.Error("config not Rendered from ENV")
	}
}

func TestFailureOnEnvironmentVariableNotPresentWhenSettingConfigThroughEnvrionmentVariable(t *testing.T) {
	ConfigFile = ""
	os.Unsetenv("READ_TIME_OUT_SECONDS_2")
	os.Setenv("J8ACFG_YML", "---\nconnection:\n  downstream:\n    readTimeoutSeconds: {{.READ_TIME_OUT_SECONDS_2}}\n    roundTripTimeoutSeconds: 20\n    idleTimeoutSeconds: 30\n    port: 8080\n    mode: HTTP\n    maxBodyBytes: 65535\n  upstream:\n    socketTimeoutSeconds: 3\n    readTimeoutSeconds: 30\n    idleTimeoutSeconds: 10\n    maxAttempts: 4\n    poolSize: 8\n    tlsInsecureSkipVerify: true\npolicies:\n  ab:\n    - label: green\n      weight: 0.8\n    - label: blue\n      weight: 0.2\nroutes:\n  - path: \"/todos\"\n    resource: jsonplaceholder\n  - path: \"/about\"\n    resource: about\n  - path: \"/mse6/some\"\n    resource: mse61\n  - path: \"/mse6/\"\n    resource: mse6\n    policy: ab\n  - path: \"/s01\"\n    resource: s01\n  - path: \"/s02\"\n    resource: s02\n  - path: \"/s03\"\n    resource: s03\n  - path: \"/s04\"\n    resource: s04\n  - path: \"/s05\"\n    resource: s05\n  - path: \"/s06\"\n    resource: s06\n  - path: \"/s07\"\n    resource: s07\n  - path: \"/s08\"\n    resource: s08\n  - path: \"/s09\"\n    resource: s09\n  - path: \"/s10\"\n    resource: s10\n  - path: \"/s11\"\n    resource: s11\n  - path: \"/s12\"\n    resource: s12\n  - path: \"/s13\"\n    resource: s13\n  - path: \"/s14\"\n    resource: s14\n  - path: \"/s15\"\n    resource: s15\n  - path: \"/s16\"\n    resource: s16\n  - path: \"/badip\"\n    resource: badip\n  - path: \"/baddns\"\n    resource: baddns\n  - path: \"/badremote\"\n    resource: badremote\n  - path: \"/badlocal\"\n    resource: badlocal\n  - path: /badssl\n    resource: badssl\nresources:\n  jsonplaceholder:\n    - url:\n        scheme: https\n        host: jsonplaceholder.typicode.com\n        port: 443\n  badssl:\n    - url:\n        scheme: https\n        host: localhost\n        port: 60101\n  badip:\n    - url:\n        scheme: http\n        host: 10.247.13.14\n        port: 29471\n  baddns:\n    - url:\n        scheme: http\n        host: kajsdkfj23848392sdfjsj332jkjkjdkshhhhimnotahost.com\n        port: 29471\n  badremote:\n    - url:\n        scheme: http\n        host: google.com\n        port: 29471\n  badlocal:\n    - url:\n        scheme: http\n        host: localhost\n        port: 29471\n  mse61:\n    - url:\n        scheme: 'http:'\n        host: localhost\n        port: 60083\n  mse6:\n    - labels:\n        - green\n      url:\n        scheme: http://\n        host: localhost\n        port: 60083\n    - labels:\n        - blue\n      url:\n        host: localhost\n        port: 60084\n  s01:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60085\n  s02:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60086\n  s03:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60087\n  s04:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60088\n  s05:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60089\n  s06:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60090\n  s07:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60091\n  s08:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60092\n  s09:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60093\n  s10:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60094\n  s11:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60095\n  s12:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60096\n  s13:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60097\n  s14:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60098\n  s15:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60099\n  s16:\n    - url:\n        scheme: http\n        host: localhost\n        port: 60100")

	shouldPanic(t, new(Config).load)
}
func TestFailureOnEnvironmentVarialbeNotPresentWhenSettingConfigThroughConfigFile(t *testing.T) {
	os.Unsetenv("PORT")

	ConfigFile = "./integration/templatej8a1.yml"
	shouldPanic(t, new(Config).load)
}

// Helps with detection of Panic in the execution
func shouldPanic(t *testing.T, f func() *Config) {
	t.Helper()
	defer func() { _ = recover() }()
	f()
	t.Errorf("Should Have Panicked")
}

func acmeConfigWith(domain string) *Config {
	config := &Config{
		Connection: Connection{
			Downstream: Downstream{
				Http: Http{Port: 80},
				Tls: Tls{
					Acme: Acme{
						Domains:  []string{domain},
						Provider: "letsencrypt",
						Email:    "noreply@example.org",
					},
				},
			},
		},
	}
	return config
}

func TestFqdnValidate(t *testing.T) {
	//should pass
	if !isd.IsDomain("adyntest.com") {
		t.Error("adyntest.com should have fqdn validated")
	}
	if !isd.IsDomain("we.money") {
		t.Error("we.moneyh should have fqdn validated")
	}
	if !isd.IsDomain("911.com.au") {
		t.Error("911.com.au should have fqdn validated")
	}
	if !isd.IsDomain("mittag.biz") {
		t.Error("mittag.biz should have fqdn validated")
	}
	if !isd.IsDomain("foo.studio") {
		t.Error("foo.studio should have fqdn validated")
	}
	if !isd.IsDomain("foo.life") {
		t.Error("foo.life should have fqdn validated")
	}
	if !isd.IsDomain("foo.shop") {
		t.Error("foo.shop should have fqdn validated")
	}
	if !isd.IsDomain("foo.health") {
		t.Error("foo.health should have fqdn validated")
	}
	if !isd.IsDomain("foo.de") {
		t.Error("foo.de should have fqdn validated")
	}
	if !isd.IsDomain("api.foo.de") {
		t.Error("api.foo.de should have fqdn validated")
	}
	if !isd.IsDomain("x.y.z.api.foo.de") {
		t.Error("x.y.z.api.foo.de should have fqdn validated")
	}
	if !isd.IsDomain("foo.co.uk") {
		t.Error("foo.co.uk should have fqdn validated")
	}
	if !isd.IsDomain("foo.tattoo") {
		t.Error("foo.tattoo should have fqdn validated")
	}
	if !isd.IsDomain("foo.design") {
		t.Error("foo.design should have fqdn validated")
	}
	if !isd.IsDomain("foo.sydney") {
		t.Error("foo.sydney should have fqdn validated")
	}
	if !isd.IsDomain("foo.melbourne") {
		t.Error("foo.melbourne should have fqdn validated")
	}

	//must fail
	if isd.IsDomain("-we.money") {
		t.Error("-we.money should not have fqdn validated")
	}
	if isd.IsDomain("_we.money") {
		t.Error("_we.money should not have fqdn validated")
	}
	if isd.IsDomain("foo.baz") {
		t.Error("foo.baz should not have fqdn validated")
	}
	if isd.IsDomain("foo.zydney") {
		t.Error("foo.zydney should not have fqdn validated")
	}
	if isd.IsDomain("foo.nelbourne") {
		t.Error("foo.nelbourne should not have fqdn validated")
	}
}

func TestParseDisableXRequestInfoTrue(t *testing.T) {
	configJson := []byte("{\"disableXRequestInfo\": true}")
	config := new(Config).parse(configJson)

	got := config.DisableXRequestInfo
	want := true

	if got != want {
		t.Error("X-Request-Info should be disabled")
	}
}

func TestParseDisableXRequestInfoFalse(t *testing.T) {
	config := new(Config).parse([]byte(""))

	got := config.DisableXRequestInfo
	want := false

	if got != want {
		t.Error("X-Request-Info should be enabled")
	}
}

func TestConfigValidationPanicsForInvalidRouteType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("config should have panicked")
		}
	}()

	config := new(Config).readYmlFile("./j8acfg.yml")
	//thou shall not pass!
	config.Routes[0].PathType = "blah"

	config = config.validateRoutes()
}

func TestConfigValidationPanicsForEmptyResource(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("config should have panicked")
		}
	}()

	config := new(Config).readYmlFile("./j8acfg.yml")
	//thou shall not pass!
	config.Resources["mse6"] = nil

	config = config.validateResources()
}

func TestConfigValidationPanicsForBadResource(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("config should have panicked")
		}
	}()

	config := new(Config).readYmlFile("./j8acfg.yml")
	//thou shall not pass!
	config.Routes[0].Resource = "asdfadsf"

	config = config.validateRoutes()
}
