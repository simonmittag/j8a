package jabba

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testSetup()
	code := m.Run()
	//teardown()
	os.Exit(code)
}

func testSetup() {
	os.Setenv("TZ", "Australia/Sydney")
	os.Setenv("LOGLEVEL", "TRACE")
	os.Setenv("LOGCOLOR", "true")
	setupJabba()
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
	configJson := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutJabba\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
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
	configJson := []byte("{\"routes\": [{\n\t\t\t\"path\": \"/about\",\n\t\t\t\"resource\": \"aboutJabba\"\n\t\t},\n\t\t{\n\t\t\t\"path\": \"/customer\",\n\t\t\t\"resource\": \"customer\",\n\t\t\t\"policy\": \"ab\"\n\t\t}\n\t]}")
	config := new(Config).parse(configJson).sortRoutes()

	customer := config.Routes[0]
	if customer.Path != "/customer" {
		t.Error("incorrectly sorted routes")
	}

	about := config.Routes[1]
	if about.Path != "/about" {
		t.Error("incorrectly sorted routes")
	}
}

func BenchmarkRouteMatchingRegex(b *testing.B) {
	config := new(Config).read("./jabba.json")
	config = config.compileRoutePaths().sortRoutes()

	for _, route := range config.Routes {
		if ok := route.matchURI(requestFactory("/mse6")); ok {
			break
		}
	}
}

func BenchmarkRouteMatchingString(b *testing.B) {
	config := new(Config).read("./jabba.json")
	config = config.sortRoutes()

	for _, route := range config.Routes {
		if ok := route.matchURI(requestFactory("/mse6")); ok {
			break
		}
	}
}

//TestReadConfigFile
func TestReadConfigFile(t *testing.T) {
	config := new(Config).read("./jabba.json")
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
	config := new(Config).read("./jabba.json").reApplyResourceSchemes()

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
