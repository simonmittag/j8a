package j8a

import (
	"github.com/rs/zerolog"
	"net/http"
	"testing"
)

func TestRouteMatchRoot(t *testing.T) {
	routeNoMatch("/", "", t)
	routeMatch("/", "/some", t)
	routeMatch("/", "/", t)
	routeMatch("/", "/some/more", t)
	routeMatch("/", "/some/more?param", t)
	routeMatch("/", "/some/more?param=value", t)
	routeMatch("/", "/some/more?param=value&param2=value2", t)
	//path is never empty string, http server inserts "/"
}

func TestRouteMatchWithSlug(t *testing.T) {
	routeNoMatch("/so", "so", t)
	routeNoMatch("/so", "/os", t)
	routeMatch("/so", "/some", t)
	routeMatch("/so", "/some/more", t)
	routeMatch("/so", "/some/more?param", t)
	routeMatch("/so", "/some/more?param=value", t)
	routeMatch("/so", "/some/more?param=value&param2=value2", t)
}

func TestRouteMatchWithTerminatedSlug(t *testing.T) {
	routeNoMatch("/some/", "some", t)
	routeNoMatch("/some/", "", t)
	routeNoMatch("/some/", "/", t)
	routeNoMatch("/some/", "/some", t)
	routeNoMatch("/some/", "/want/some/", t)
	routeMatch("/some/", "/some/", t)
	routeMatch("/some/", "/some/more", t)
	routeMatch("/some/", "/some/more?param", t)
	routeMatch("/some/", "/some/more?param=value", t)
	routeMatch("/some/", "/some/more?param=value&param2=value2", t)
}

func TestRouteMatchWithWildcardSlug(t *testing.T) {
	routeNoMatch("*/some/", "some", t)
	routeNoMatch("*/some/", "", t)
	routeNoMatch("*/some/", "/", t)
	routeNoMatch("*/some/", "/some", t)
	routeMatch("*/some/", "/want/some/", t)
	routeMatch("*/some/", "/really/want/some/", t)
	routeMatch("*/some/", "/really/want/some/more", t)
	routeMatch("*/some/", "/really/want/some/more?with=params", t)
	routeMatch("*/some/", "/really/want/some/more/", t)
	routeMatch("*/some/", "/some/", t)
	routeMatch("*/some/", "/some/more", t)
	routeMatch("*/some/", "/some/more?param", t)
	routeMatch("*/some/", "/some/more?param=value", t)
	routeMatch("*/some/", "/some/more?param=value&param2=value2", t)
}

// some of these do not match because regex greedy matches to "/"
func TestRouteMatchWithAbsoluteWildcardSlug(t *testing.T) {
	routeNoMatch("/*/some/", "/want/some/", t)
	routeNoMatch("/*/some/", "/really/want/some/", t)
	routeNoMatch("/*/some/", "/really/want/some/more", t)
	routeNoMatch("/*/some/", "/really/want/some/more?with=params", t)
	routeNoMatch("/*/some/", "/really/want/some/more/", t)
}

// TODO: test trailing slashes better. Do we want matches?
// TODO: test query strings
func TestRouteExactMatch(t *testing.T) {
	routeExactMatch("/some/", "/some/", t)
	routeExactNoMatch("/some/", "/some", t)
	routeExactMatch("/some/index.html", "/some/index.html", t)
	routeExactNoMatch("/some/index.html", "/some/index", t)
	routeExactNoMatch("/some/index.html", "/some/index.", t)
	routeExactNoMatch("/some/index.html", "/some/index.htm", t)
}

func routeMatch(route string, path string, t *testing.T) {
	r := routeFactory(route)
	req := requestFactory(path)
	if !r.matchURI(req) {
		t.Errorf("route %v did not match desired path: %v", route, path)
	}
}

func routeNoMatch(route string, path string, t *testing.T) {
	r := routeFactory(route)
	req := requestFactory(path)
	if r.matchURI(req) {
		t.Errorf("route %v did match undesired path: %v", route, path)
	}
}

func routeExactMatch(route string, path string, t *testing.T) {
	r := routeFactory(route, "exact")
	req := requestFactory(path)
	if !r.matchURI(req) {
		t.Errorf("route %v did not exactly match desired path: %v", route, path)
	}
}

func routeExactNoMatch(route string, path string, t *testing.T) {
	r := routeFactory(route, "exact")
	req := requestFactory(path)
	if r.matchURI(req) {
		t.Errorf("route %v did exactly match undesired path: %v", route, path)
	}
}

// TODO this needs host
func requestFactory(path string) *http.Request {
	req, _ := http.NewRequest("GET", path, nil)
	req.RequestURI = path
	return req
}

func routeFactory(args ...string) Route {
	r := Route{
		Path: args[0],
	}
	if len(args) > 1 && "exact" == args[1] {
		r.PathType = args[1]
	}
	r.compilePath()
	return r
}

func TestRouteMapDefault(t *testing.T) {
	Runner = mockRuntime()
	r := Runner.Routes[1]
	gotUrl, gotLabel, got := r.mapURL(&Proxy{})

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
		t.Errorf("url did not successfully map, want %v, got %v, wantUrl,  url", wantUrl, gotUrl)
	}
}

func TestRouteMap(t *testing.T) {
	Runner = mockRuntime()
	r := Runner.Routes[0]
	gotUrl, gotLabel, got := r.mapURL(&Proxy{})
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
		t.Errorf("url did not successfully map, want %v, got %v, wantUrl,  url", wantUrl, gotUrl)
	}
}

func TestRouteMapIncorrectPolicyFails(t *testing.T) {
	Runner = mockRuntime()
	r := Runner.Routes[1]
	r.Policy = "i'm_not_real"
	_, _, got := r.mapURL(&Proxy{})
	if got != false {
		t.Error("route with bad policy is not allowed to map. this is a configuration error")
	}
}

func TestRoutePathTypesAreValid(t *testing.T) {
	rpt := NewRoutePathTypes()
	if !rpt.isValid("exact") {
		t.Error("exact should be valid")
	}
	if !rpt.isValid("pREFix") {
		t.Error("prefix should be valid")
	}
	if rpt.isValid("") {
		t.Error("empty string should not be valid")
	}
	if rpt.isValid("test") {
		t.Error("test should not be valid")
	}
}

func BenchmarkRouteMatchingRegex(b *testing.B) {
	//suppress noise
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config := new(Config).readYmlFile("./j8acfg.yml")
	config = config.compileRoutePaths().validateRoutes()

	for i := 0; i < b.N; i++ {
		for _, route := range config.Routes {
			if ok := route.matchURI(requestFactory("/mse6")); ok {
				break
			}
		}
	}
}

func BenchmarkRouteMatchingString(b *testing.B) {
	//suppress noise
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config := new(Config).readYmlFile("./j8acfg.yml")
	config = config.validateRoutes()

	for i := 0; i < b.N; i++ {
		for _, route := range config.Routes {
			if ok := route.matchURI(requestFactory("/mse6")); ok {
				break
			}
		}
	}
}
