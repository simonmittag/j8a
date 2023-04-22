package j8a

import (
	"github.com/rs/zerolog"
	"net/http"
	"testing"
)

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

func doRunRouteMatchingTests(t *testing.T, tests []struct {
	n string
	r string
	t string
	u string
	v bool
}) {
	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			rr := Route{
				Path:     tt.r,
				PathType: tt.t,
			}
			//this is only to validate the test. invalid paths should not be tested and there are many.
			if v, _ := rr.validPath(); !v {
				t.Errorf("routepath %v should be valid. this is a test setup problem", rr.Path)
			}
			//test actual URL matching
			if rr.matchURI(requestFactory(tt.u)) != tt.v {
				t.Errorf("route %v type %v should match URL %v outcome %v", rr.Path, rr.PathType, tt.u, tt.v)
			}
		})
	}
}

func TestRouteMatchRoot(t *testing.T) {
	tests := []struct {
		n string
		r string
		t string
		u string
		v bool
	}{
		{n: "match root", r: "/", t: "prefix", u: "", v: false},
		{n: "match root", r: "/", t: "prefix", u: "/some", v: true},
		{n: "match root", r: "/", t: "prefix", u: "/", v: true},
		{n: "match root", r: "/", t: "prefix", u: "/some/more", v: true},
		{n: "match root", r: "/", t: "prefix", u: "/some/more?k", v: true},
		{n: "match root", r: "/", t: "prefix", u: "/some/more?k=v", v: true},
		{n: "match root", r: "/", t: "prefix", u: "/some/more?k=v&k2=v2", v: true},
	}

	doRunRouteMatchingTests(t, tests)
}

func TestRouteMatchWithSlug(t *testing.T) {
	tests := []struct {
		n string
		r string
		t string
		u string
		v bool
	}{
		{n: "match slug", r: "/so", t: "prefix", u: "so", v: false},
		{n: "match slug", r: "/so", t: "prefix", u: "/os", v: false},
		{n: "match slug", r: "/so", t: "prefix", u: "/some", v: true},
		{n: "match slug", r: "/so", t: "prefix", u: "/some/more", v: true},
		{n: "match slug", r: "/so", t: "prefix", u: "/some/more?k", v: true},
		{n: "match slug", r: "/so", t: "prefix", u: "/some/more?k=v", v: true},
		{n: "match slug", r: "/so", t: "prefix", u: "/some/more?k=v&k2=v2", v: true},
	}

	doRunRouteMatchingTests(t, tests)
}

func TestRouteMatchWithTerminatedSlug(t *testing.T) {
	tests := []struct {
		n string
		r string
		t string
		u string
		v bool
	}{
		{n: "missingPrefix", r: "/some/", t: "prefix", u: "some", v: false},
		{n: "missingPath", r: "/some/", t: "prefix", u: "", v: false},
		{n: "simple non match", r: "/some/", t: "prefix", u: "/", v: false},
		{n: "matching with missing trailing slash", r: "/some/", t: "prefix", u: "/some", v: true},
		{n: "slug not matching", r: "/some/", t: "prefix", u: "/want/some", v: false},
		{n: "slug exact match but type prefix", r: "/some/", t: "prefix", u: "/some/", v: true},
		{n: "slug prefix match type prefix", r: "/some/", t: "prefix", u: "/some/more", v: true},
		{n: "slug prefix match type prefix with params", r: "/some/", t: "prefix", u: "/some/more?param", v: true},
		{n: "slug prefix match type prefix with params", r: "/some/", t: "prefix", u: "/some/more?param=value", v: true},
		{n: "slug prefix match type prefix with params", r: "/some/", t: "prefix", u: "/some/more?param=value&param2=value2", v: true},
	}

	doRunRouteMatchingTests(t, tests)
}

func TestRouteMatchWithWildcardSlug(t *testing.T) {
	tests := []struct {
		n string
		r string
		t string
		u string
		v bool
	}{
		//valid route non matching url
		{n: "valid Route with wildcard non matching url", r: "/[a-z]*/some/", t: "prefix", u: "some", v: false},
		{n: "valid Route with wildcard non matching url", r: "/[a-z]*/some/", t: "prefix", u: "", v: false},
		{n: "valid Route with wildcard non matching url", r: "/[a-z]*/some/", t: "prefix", u: "/", v: false},
		//matches the left subpath, then the right. This is legal but iffy
		{n: "right matching", r: "/[a-z]*/some/", t: "prefix", u: "/want/some", v: true},
		{n: "right matching", r: "/[a-z]*/[a-z]*/some/", t: "prefix", u: "/really/want/some", v: true},
		{n: "right matching", r: "/[a-z]*/[a-z]*/some/", t: "prefix", u: "/really/want/some/more", v: true},
		{n: "right matching", r: "/[a-z]*/[a-z]*/some/", t: "prefix", u: "/really/want/some/more?k", v: true},
		{n: "right matching", r: "/[a-z]*/[a-z]*/some/", t: "prefix", u: "/really/want/some/more?k=v", v: true},
		{n: "right matching", r: "/[a-z]*/[a-z]*/some/", t: "prefix", u: "/really/want/some/more?k=v&k2=v2", v: true},
		//match multiple subpaths, then match right
		{n: "right matching", r: "/[a-z,/]*/some/", t: "prefix", u: "/want/want/want/want/want/some/more?param=value&param2=value2", v: true},
		{n: "right matching", r: "/[a-z,/]*/some/", t: "prefix", u: "/want/want/want/want/want/XXXX/more?param=value&param2=value2", v: false},
	}

	doRunRouteMatchingTests(t, tests)
}

func TestRouteExactMatch(t *testing.T) {
	tests := []struct {
		n string
		r string
		t string
		u string
		v bool
	}{
		//valid route non matching url
		{n: "match exact with params", r: "/some/", t: "exact", u: "/some/?k=v", v: true},
		{n: "exact pathtype does not append trailing slash", r: "/some/", t: "exact", u: "/some", v: false},
		{n: "match file", r: "/some/index.html", t: "exact", u: "/some/index.html", v: true},
		{n: "match file with params", r: "/some/index.html", t: "exact", u: "/some/index.html?k=v", v: true},
		{n: "match file with params", r: "/some/index.html", t: "exact", u: "/some/index.html?k=v&k2=v2", v: true},
		{n: "do not match file unless exact", r: "/some/index.html", t: "exact", u: "/some/index", v: false},
		{n: "do not match file unless exact", r: "/some/index.html", t: "exact", u: "/some/index.", v: false},
		{n: "do not match file unless exact", r: "/some/index.html", t: "exact", u: "/some/index.htm", v: false},
	}

	doRunRouteMatchingTests(t, tests)
}

func TestRouteUnicodeMatch(t *testing.T) {
	tests := []struct {
		n string
		r string
		t string
		u string
		v bool
	}{
		//valid route non matching url
		{n: "match unicode chars type exact", r: "/foo指", t: "exact", u: "/foo指", v: true},
		{n: "match unicode chars type exact with params", r: "/foo/指", t: "exact", u: "/foo/指?k=v", v: true},
		{n: "match unicode subpaths exact with params", r: "/指/指", t: "exact", u: "/指/指?k=v", v: true},
		{n: "match unicode subpaths prefix", r: "/指/指", t: "prefix", u: "/指/指aaa?k=v", v: true},
		{n: "match unicode subpaths prefix", r: "/指/指", t: "prefix", u: "/指/指😀?k=v", v: true},
		{n: "match unicode subpaths with regex compilation", r: "/指*/指", t: "prefix", u: "/指指指指/指aaa", v: true},
		{n: "match unicode subpaths with regex compilation and params", r: "/指*/指", t: "prefix", u: "/指指指指/指aaa?k=v", v: true},
		{n: "match emoji as prefix", r: "/😀/😀", t: "prefix", u: "/😀/😀aaa", v: true},
		{n: "match emoji as prefix with regex compilation", r: "/[😀]*/😀", t: "prefix", u: "/😀😀😀😀😀😀/😀aaa", v: true},
		{n: "match emoji as prefix with regex compilation", r: "/😀*/😀", t: "prefix", u: "/😀/😀aaa", v: true},
		{n: "match emoji as prefix with regex compilation", r: "/[😀]*/😀", t: "prefix", u: "/a/😀aaa", v: false},
	}

	doRunRouteMatchingTests(t, tests)
}

func TestKubernetesIngressExamples(t *testing.T) {
	tests := []struct {
		n string
		r string
		t string
		u string
		v bool
	}{
		//valid route non matching url
		{n: "kubernetes ingress examples", r: "/", t: "prefix", u: "/any/thing?k=v", v: true},
		{n: "kubernetes ingress examples", r: "/foo", t: "exact", u: "/foo", v: true},
		{n: "kubernetes ingress examples", r: "/foo", t: "exact", u: "/bar", v: false},
		{n: "kubernetes ingress examples", r: "/foo", t: "exact", u: "/foo/", v: false},
		{n: "kubernetes ingress examples", r: "/foo/", t: "exact", u: "/foo", v: false},
		{n: "kubernetes ingress examples", r: "/foo", t: "prefix", u: "/foo", v: true},
		{n: "kubernetes ingress examples", r: "/foo", t: "prefix", u: "/foo/", v: true},
		//appended by matcher in prefix mode but not in exact
		{n: "kubernetes ingress examples", r: "/foo/", t: "prefix", u: "/foo", v: true},
		{n: "kubernetes ingress examples", r: "/foo/", t: "exact", u: "/foo", v: false},
		{n: "kubernetes ingress examples", r: "/foo/", t: "exact", u: "/foo/", v: true},
		//we match the last element of a path as substring, kube doesn't, see Kube ingress: https://kubernetes.io/docs/concepts/services-networking/ingress/
		{n: "kubernetes ingress examples", r: "/aaa/bb", t: "prefix", u: "/aaa/bbb", v: true},
		{n: "kubernetes ingress examples", r: "/aaa/bbb", t: "exact", u: "/aaa/bbb", v: true},
		//ignores trailing slash
		{n: "kubernetes ingress examples", r: "/aaa/bbb/", t: "prefix", u: "/aaa/bbb", v: true},
		//matches trailing slash
		{n: "kubernetes ingress examples", r: "/aaa/bbb", t: "prefix", u: "/aaa/bbb/", v: true},
		{n: "kubernetes ingress examples", r: "/aaa/bbb", t: "prefix", u: "/aaa/bbb/ccc", v: true},
		//we match the last element of a path as substring, kube doesn't, see Kube ingress: https://kubernetes.io/docs/concepts/services-networking/ingress/
		{n: "kubernetes ingress examples", r: "/aaa/bbb", t: "prefix", u: "/aaa/bbbxyz", v: true},
	}

	doRunRouteMatchingTests(t, tests)
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

func TestRoutePathsValid(t *testing.T) {
	mkPrfx := func(p string) Route {
		return Route{
			Path:     p,
			PathType: "prefix",
		}
	}
	var tests = []struct {
		n string
		r Route
		v bool
	}{
		{n: "simple", r: mkPrfx("/a"), v: true},
		{n: "simple", r: mkPrfx("/a/b"), v: true},
		{n: "double slash", r: mkPrfx("//a/b"), v: true},
		{n: "regex", r: mkPrfx("/a/b/*"), v: true},
		{n: "regex", r: mkPrfx("/a/b/a*"), v: true},
		{n: "emoji", r: mkPrfx("/😈"), v: true},
		{n: "unicode", r: mkPrfx("/指"), v: true},
		{n: "regex", r: mkPrfx("/a/b/*"), v: true},

		{n: "regex", r: mkPrfx("/a/b/**"), v: false},
		{n: "no space", r: mkPrfx("/a a"), v: false},
		{n: "no space", r: mkPrfx("/a a"), v: false},
		{n: "no space", r: mkPrfx(" /a"), v: false},
		{n: "no slash prefix", r: mkPrfx("a/a"), v: false},
	}

	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			if v, _ := tt.r.validPath(); v != tt.v {
				t.Errorf("routepath %v should be %v", tt.r.Path, tt.v)
			}
		})
	}
}

func TestRouteSorting(t *testing.T) {
	//the sort order needs to be exact over prefix, then by longest path.

	config := new(Config).readYmlFile("./j8acfg.yml")
	config = config.compileRoutePaths().validateRoutes()

	if config.Routes[0].Path != "/path/med/1st" ||
		config.Routes[1].Path != "/path/med/2" ||
		config.Routes[2].Path != "/longfirstslug_longfirstslug_longfirstslug_longfirstslug/short" ||
		config.Routes[3].Path != "/badremote" {
		t.Errorf("sort order wrong %v", config.Routes)
	}
}

func TestRoutePathLess(t *testing.T) {
	var tests = []struct {
		name string
		rl   Route
		rg   Route
	}{
		{
			name: "empty paths for nil check",
			rl: Route{
				Path:     "",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "",
				PathType: "prefix",
			},
		},
		{
			name: "one longer vs. one shorter wildcar routepath that isn't a slug",
			rl: Route{
				Path:     "aabcccc*",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "aaa*",
				PathType: "prefix",
			},
		},
		{
			name: "one longer vs. one shorter routepath that isn't a slug",
			rl: Route{
				Path:     "aabcccc",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "aaa",
				PathType: "prefix",
			},
		},
		{
			name: "one vs. one routepath that isn't a slug",
			rl: Route{
				Path:     "aab",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "aaa",
				PathType: "prefix",
			},
		},
		{
			name: "one vs. one slugs",
			rl: Route{
				Path:     "/aab",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa",
				PathType: "prefix",
			},
		},
		{
			name: "one vs. one slugs",
			rl: Route{
				Path:     "/aab",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa",
				PathType: "prefix",
			},
		},
		{
			name: "two vs. one slugs",
			rl: Route{
				Path:     "/aaa/bbb",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa",
				PathType: "prefix",
			},
		},
		{
			name: "longer subpath slug at the end",
			rl: Route{
				Path:     "/aaa/bbb/cc",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa/bbb/c",
				PathType: "prefix",
			},
		},
		{
			name: "longer subpath slug higher up the chain",
			rl: Route{
				Path:     "/aaa/bbbbbbbbbbbbbbbbbbb/c",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa/bbb/cc/d/e/f",
				PathType: "prefix",
			},
		},
		{
			name: "longer subpath slug top of the chain",
			rl: Route{
				Path:     "/aaaaaa/bbb/c",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa/bbb/c/d/e/f",
				PathType: "prefix",
			},
		},
		{
			name: "more subpath slugs at the end of the chain for same root",
			rl: Route{
				Path:     "/aaa/bbb/ccc/ddd/",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa/bbb/ccc/ddd",
				PathType: "prefix",
			},
		},
		{
			name: "more subpath slugs at the end of the chain for same root",
			rl: Route{
				Path:     "/aaa/bbb/ccc/ddd/eee/fff/ggg",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa/bbb/ccc/ddd",
				PathType: "prefix",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !NewRoutePath(tt.rl).Less(NewRoutePath(tt.rg)) {
				t.Errorf("routepath %v should be less than %v", tt.rl, tt.rg)
			}
		})
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
