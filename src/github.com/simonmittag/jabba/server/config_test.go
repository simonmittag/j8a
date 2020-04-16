package server

import (
	"testing"
)

//TestDefaultDownstreamReadTimeout
func TestDefaultDownstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.ReadTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamIdleTimeout
func TestDefaultDownstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultDownstreamRoundtripTimeout
func TestDefaultDownstreamRoundtripTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Downstream.RoundTripTimeoutSeconds
	want := 240
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamSocketTimeout
func TestDefaultUpstreamSocketTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.SocketTimeoutSeconds
	want := 3
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamSocketTimeout
func TestDefaultUpstreamReadTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.ReadTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamIdleTimeout
func TestDefaultUpstreamIdleTimeout(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.IdleTimeoutSeconds
	want := 120
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamConnectionPoolSize
func TestDefaultUpstreamConnectionPoolSize(t *testing.T) {
	config := new(Config).setDefaultValues()
	got := config.Connection.Upstream.PoolSize
	want := 32768
	if got != want {
		t.Errorf("default config got %d, want %d", got, want)
	}
}

//TestDefaultUpstreamConnectionMaxAttempts
func TestDefaultUpstreamConnectionMaxAttempts(t *testing.T) {
	config := new(Config).setDefaultValues()
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
