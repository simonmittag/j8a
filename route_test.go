package jabba

import (
	"github.com/rs/zerolog"
	"net/http"
	"regexp"
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
	routeMatch("/some/", "/some/", t)
	routeMatch("/some/", "/some/more", t)
	routeMatch("/some/", "/some/more?param", t)
	routeMatch("/some/", "/some/more?param=value", t)
	routeMatch("/some/", "/some/more?param=value&param2=value2", t)
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

func requestFactory(path string) *http.Request {
	req, _ := http.NewRequest("GET", path, nil)
	req.RequestURI = path
	return req
}

func routeFactory(route string) Route {
	pR, _ := regexp.Compile("^" + route)
	r := Route{
		Path:  route,
		Regex: pR,
	}
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
		t.Errorf("url did not successfuly map, want %v, got %v, wantUrl,  url", wantUrl, gotUrl)
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
		t.Errorf("url did not successfuly map, want %v, got %v, wantUrl,  url", wantUrl, gotUrl)
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

func BenchmarkRouteMatchingRegex(b *testing.B) {
	//suppress noise
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config := new(Config).read("./jabba.json")
	config = config.compileRoutePaths().sortRoutes()

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

	config := new(Config).read("./jabba.json")
	config = config.sortRoutes()

	for i := 0; i < b.N; i++ {
		for _, route := range config.Routes {
			if ok := route.matchURI(requestFactory("/mse6")); ok {
				break
			}
		}
	}
}
