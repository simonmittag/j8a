package j8a

import (
	"github.com/rs/zerolog"
	"net/http"
	"testing"
)

func TestRouteMatchRoot(t *testing.T) {
	routePrefixNoMatch("/", "", t)
	routePrefixMatch("/", "/some", t)
	routePrefixMatch("/", "/", t)
	routePrefixMatch("/", "/some/more", t)
	routePrefixMatch("/", "/some/more?param", t)
	routePrefixMatch("/", "/some/more?param=value", t)
	routePrefixMatch("/", "/some/more?param=value&param2=value2", t)
	//path is never empty string, http server inserts "/"
}

func TestRouteMatchWithSlug(t *testing.T) {
	routePrefixNoMatch("/so", "so", t)
	routePrefixNoMatch("/so", "/os", t)
	routePrefixMatch("/so", "/some", t)
	routePrefixMatch("/so", "/some/more", t)
	routePrefixMatch("/so", "/some/more?param", t)
	routePrefixMatch("/so", "/some/more?param=value", t)
	routePrefixMatch("/so", "/some/more?param=value&param2=value2", t)
}

func TestRouteMatchWithTerminatedSlug(t *testing.T) {
	routePrefixNoMatch("/some/", "some", t)
	routePrefixNoMatch("/some/", "", t)
	routePrefixNoMatch("/some/", "/", t)
	routePrefixMatch("/some/", "/some", t)
	routePrefixNoMatch("/some/", "/want/some/", t)
	routePrefixMatch("/some/", "/some/", t)
	routePrefixMatch("/some/", "/some/more", t)
	routePrefixMatch("/some/", "/some/more?param", t)
	routePrefixMatch("/some/", "/some/more?param=value", t)
	routePrefixMatch("/some/", "/some/more?param=value&param2=value2", t)
}

func TestRouteMatchWithWildcardSlug(t *testing.T) {
	routePrefixNoMatch("*/some/", "some", t)
	routePrefixNoMatch("*/some/", "", t)
	routePrefixNoMatch("*/some/", "/", t)
	routePrefixMatch("*/some/", "/want/some/", t)
	routePrefixMatch("*/some/", "/really/want/some/", t)
	routePrefixMatch("*/some/", "/really/want/some/more", t)
	routePrefixMatch("*/some/", "/really/want/some/more?with=params", t)
	routePrefixMatch("*/some/", "/really/want/some/more/", t)
	//trailing slash is appended, then matches
	routePrefixMatch("*/some/", "/some", t)
	routePrefixMatch("*/some/", "/some/", t)
	routePrefixMatch("*/some/", "/some/more", t)
	routePrefixMatch("*/some/", "/some/more?param", t)
	routePrefixMatch("*/some/", "/some/more?param=value", t)
	routePrefixMatch("*/some/", "/some/more?param=value&param2=value2", t)
}

// some of these do not match because regex greedy matches to "/"
func TestRouteMatchWithAbsoluteWildcardSlug(t *testing.T) {
	routePrefixNoMatch("/*/some/", "/want/some/", t)
	routePrefixNoMatch("/*/some/", "/really/want/some/", t)
	routePrefixNoMatch("/*/some/", "/really/want/some/more", t)
	routePrefixNoMatch("/*/some/", "/really/want/some/more?with=params", t)
	routePrefixNoMatch("/*/some/", "/really/want/some/more/", t)
}

func TestRouteExactMatch(t *testing.T) {
	routeExactMatch("/some/", "/some/?k=v", t)
	routeExactMatch("/some/", "/some/", t)
	routeExactNoMatch("/some/", "/some", t)
	routeExactMatch("/some/index.html", "/some/index.html", t)
	routeExactMatch("/some/index.html", "/some/index.html?k=v", t)
	routeExactMatch("/some/index.html", "/some/index.html?k=v&k2=v2", t)
	routeExactNoMatch("/some/index.html", "/some/index", t)
	routeExactNoMatch("/some/index.html", "/some/index.", t)
	routeExactNoMatch("/some/index.html", "/some/index.htm", t)
}

func TestKubernetesIngressExamples(t *testing.T) {
	routePrefixMatch("/", "/any/thing?k=v", t)
	routeExactMatch("/foo", "/foo", t)
	routeExactNoMatch("/foo", "/bar", t)
	routeExactNoMatch("/foo", "/foo/", t)
	routeExactNoMatch("/foo/", "/foo", t)
	routePrefixMatch("/foo", "/foo", t)
	routePrefixMatch("/foo", "/foo/", t)
	//appended by matcher for prefix mode
	routePrefixMatch("/foo/", "/foo", t)
	//but can't match in exact mode
	routeExactNoMatch("/foo/", "/foo", t)
	routePrefixMatch("/foo/", "/foo/", t)

	//we match the last element of a path as substring, kube doesn't, see Kube ingress: https://kubernetes.io/docs/concepts/services-networking/ingress/
	routePrefixMatch("/aaa/bb", "/aaa/bbb", t)
	routePrefixMatch("/aaa/bbb", "/aaa/bbb", t)
	//ignores trailing slash
	routePrefixMatch("/aaa/bbb/", "/aaa/bbb", t)
	//matches trailing slash
	routePrefixMatch("/aaa/bbb", "/aaa/bbb/", t)
	routePrefixMatch("/aaa/bbb", "/aaa/bbb/ccc", t)
	//we match the last element of a path as substring, kube doesn't, see Kube ingress: https://kubernetes.io/docs/concepts/services-networking/ingress/
	routePrefixMatch("/aaa/bbb", "/aaa/bbbxyz", t)
	routePrefixMatch("/aaa/bbb", "/aaa/bbbxyz", t)
}

func routePrefixMatch(route string, path string, t *testing.T) {
	r := routeFactory(route, "prefix")
	req := requestFactory(path)
	if !r.matchURI(req) {
		t.Errorf("route %v did not match desired path: %v", route, path)
	}
}

func routePrefixNoMatch(route string, path string, t *testing.T) {
	r := routeFactory(route, "prefix")
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
	if len(args) > 1 {
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

func BenchmarkRouteMatchingRegexBusyVersion(b *testing.B) {
	//suppress noise
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config := new(Config).readYmlFile("./j8acfg.yml")
	config = config.compileRoutePaths().validateRoutes()

	for i := 0; i < b.N; i++ {
		for _, route := range config.Routes {
			if ok := route.matchURI(requestFactory("/s16")); ok {
				break
			}
		}
	}
}

func BenchmarkRouteMatchingRegexNaiveBusyVersion(b *testing.B) {
	//suppress noise
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config := new(Config).readYmlFile("./j8acfg.yml")
	config = config.compileRoutePaths().validateRoutes()

	for i := 0; i < b.N; i++ {
		for _, route := range config.Routes {
			if ok := route.matchURI_Naive(requestFactory("/s16")); ok {
				break
			}
		}
	}
}

// we can no longer run this it's illegal. Route paths must be compiled.
//func BenchmarkRouteMatchingString(b *testing.B) {
//	//suppress noise
//	zerolog.SetGlobalLevel(zerolog.InfoLevel)
//
//	config := new(Config).readYmlFile("./j8acfg.yml")
//	config = config.validateRoutes()
//
//	for i := 0; i < b.N; i++ {
//		for _, route := range config.Routes {
//			if ok := route.matchURI(requestFactory("/mse6")); ok {
//				break
//			}
//		}
//	}
//}
