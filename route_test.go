package jabba

import (
	"testing"
)

func TestRouteMapDefault(t *testing.T) {
	Runner = mockRuntime()
	r := Runner.Routes[1]
	gotUrl, gotLabel, got := r.mapURL()

	if got != true {
		t.Error("routes do not successfully map")
	}
	wantLabel := "default"
	if gotLabel != "default" {
		t.Errorf("label did not successfully map. want %v, got %v", wantLabel, gotLabel)
	}

	wantUrl := URL{
		Scheme: "http",
		Host:   "localhost",
		Port:   8084,
	}
	if *gotUrl != wantUrl {
		t.Errorf("url did not successfuly map, want %v, got %v, wantUrl,  url", wantUrl, gotUrl)
	}
}

func TestRouteMap(t *testing.T) {
	Runner = mockRuntime()
	r := Runner.Routes[0]
	gotUrl, gotLabel, got := r.mapURL()
	if got != true {
		t.Error("routes do not successfully map")
	}
	wantLabel := "simple"
	if gotLabel != "simple" {
		t.Errorf("label did not successfully map. want %v, got %v", wantLabel, gotLabel)
	}

	wantUrl := URL{
		Scheme: "http",
		Host:   "localhost",
		Port:   8083,
	}
	if *gotUrl != wantUrl {
		t.Errorf("url did not successfuly map, want %v, got %v, wantUrl,  url", wantUrl, gotUrl)
	}
}

func TestRouteMapIncorrectPolicyFails(t *testing.T) {
	Runner = mockRuntime()
	r := Runner.Routes[1]
	r.Policy = "i'm_not_real"
	_, _, got := r.mapURL()
	if got != false {
		t.Error("route with bad policy is not allowed to map. this is a configuration error")
	}
}
