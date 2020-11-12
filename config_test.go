package j8a

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testSetup()
	code := m.Run()
	os.Exit(code)
}

func testSetup() {
	os.Setenv("TZ", "Australia/Sydney")
	os.Setenv("LOGLEVEL", "TRACE")
	os.Setenv("LOGCOLOR", "true")
}

//TestDefaultDownstreamReadTimeout
func TestDefaultDownstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.ReadTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamIdleTimeout
func TestDefaultDownstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamRoundtripTimeout
func TestDefaultDownstreamRoundtripTimeout(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.RoundTripTimeoutSeconds
	want := 240
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamIdleTimeout
func TestDefaultDownstreamTlsPort(t *testing.T) {
	config := new(Config)
	config.Connection.Downstream.Mode = "tls"
	//TODO: should i turn the entire config method set into receiver type pointer?
	config = config.setDefaultDownstreamParams()

	got := config.Connection.Downstream.Port
	want := 443
	if got != want {
		t.Errorf("default tls config got port %d, want %d", got, want)
	}
}

//TestDefaultDownstreamIdleTimeout
func TestDefaultDownstreamHttpPort(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.Port
	want := 8080
	if got != want {
		t.Errorf("default http config got port %d, want %d", got, want)
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

func TestDefaultDownstreamMode(t *testing.T) {
	config := new(Config).setDefaultDownstreamParams()
	got := config.Connection.Downstream.Mode
	want := "HTTP"
	if got != want {
		t.Errorf("default config got mode %s, want %s", got, want)
	}
}

//TestDefaultUpstreamSocketTimeout
func TestDefaultUpstreamSocketTimeout(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.SocketTimeoutSeconds
	want := 3
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamSocketTimeout
func TestDefaultUpstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.ReadTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamIdleTimeout
func TestDefaultUpstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamConnectionPoolSize
func TestDefaultUpstreamConnectionPoolSize(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.PoolSize
	want := 32768
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamConnectionMaxAttempts
func TestDefaultUpstreamConnectionMaxAttempts(t *testing.T) {
	config := new(Config).setDefaultUpstreamParams()
	got := config.Connection.Upstream.MaxAttempts
	want := 1
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultPolicy
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

//TestParseResource
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

//TestParseConnection
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

		wp := 8080
		gp := c.Downstream.Port
		if wp != gp {
			t.Errorf("incorrectly parsed downstream port, want %d, got %d", wp, gp)
		}

		wm := "TLS"
		gm := c.Downstream.Mode
		if wm != gm {
			t.Errorf("incorrectly parsed downstream mode, want %s, got %s", wm, gm)
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

//TestParseRoute
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

func TestSortRoutes(t *testing.T) {
	configJson := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutj8a\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
	config := new(Config).parse(configJson).validateRoutes()

	customer := config.Routes[0]
	if customer.Path != "/customer" {
		t.Error("incorrectly sorted routes")
	}

	about := config.Routes[1]
	if about.Path != "/about" {
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

//TestReadConfigFile
func TestReadConfigFile(t *testing.T) {
	config := new(Config).readYmlFile("./j8acfg.yml")
	if config.Routes == nil {
		t.Error("incorrectly parsed routes in config file")
	}
	if config.Policies == nil {
		t.Error("incorrectly parsed policies in config file")
	}
	if config.Connection == *new(Connection) {
		t.Error("incorrectly parsed connection in config file")
	}
	if config.Resources == nil {
		t.Error("incorrectly parsed resources in config file")
	}
}

func TestReApplyScheme(t *testing.T) {
	want := map[string]string{"http": "", "https": ""}
	config := new(Config).readYmlFile("./j8acfg.yml").reApplyResourceSchemes()

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
	if config.Connection.Downstream.Port != 8443 {
		t.Error("config not loaded from load() function")
	}
}
