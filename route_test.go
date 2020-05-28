package jabba

import (
	"net/http"
	"regexp"
	"testing"
)

func TestRouteMatchRoot(t *testing.T) {
	doMatch(t, "/some", "/", true)
	doMatch(t, "/", "/", true)
	doMatch(t, "/some/more", "/", true)
	doMatch(t, "/some/more?param", "/", true)
	doMatch(t, "/some/more?param=value", "/", true)
	doMatch(t, "/some/more?param=value&param2=value2", "/", true)
	//path is never empty string, http server inserts "/"
}

func TestRouteMatchWithSlug(t *testing.T) {
	doMatch(t, "/some", "/so", true)
	doMatch(t, "/some/more", "/so", true)
	doMatch(t, "/some/more?param", "/so", true)
	doMatch(t, "/some/more?param=value", "/so", true)
	doMatch(t, "/some/more?param=value&param2=value2", "/so", true)
}

func TestRouteMatchWithTerminatedSlug(t *testing.T) {
	doMatch(t, "/some/", "/some/", true)
	doMatch(t, "/some/more", "/some/", true)
	doMatch(t, "/some/more?param", "/some/", true)
	doMatch(t, "/some/more?param=value", "/some/", true)
	doMatch(t, "/some/more?param=value&param2=value2", "/some/", true)
}

func doMatch(t *testing.T, path string, route string, yes bool) {
	r := routeFactory(route)
	req := requestFactory(path)
	m := false
	if yes {
		m = !r.matchURI(req)
	} else {
		m = r.matchURI(req)
	}
	if m {
		t.Errorf("route %v did not match path: %v", route, path)
	}
}

func requestFactory(path string) *http.Request {
	req, _ := http.NewRequest("GET", path, nil)
	req.RequestURI = path
	return req
}

func routeFactory(route string) (Route) {
	pR, _ := regexp.Compile("^"+route)
	r := Route{
		Path:  route,
		Regex: pR,
	}
	return r
}

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
