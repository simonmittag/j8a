package j8a

import (
	"github.com/rs/zerolog"
	"golang.org/x/net/idna"
	"net/http"
	"sort"
	"testing"
)

// TODO this needs host
func requestFactory(args ...string) *http.Request {
	req, _ := http.NewRequest("GET", args[0], nil)
	if len(args) > 1 {
		req.Host = args[1]
	}
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
			if rr.match(requestFactory(tt.u)) != tt.v {
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
		{n: "match slug", r: "/so*", t: "prefix", u: "/some/more?k=v&k2=v2", v: true},
		{n: "match slug", r: "/so/*", t: "prefix", u: "/some/more?k=v&k2=v2", v: true},
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

func TestHostDNSNamePatternValid(t *testing.T) {
	for _, tt := range HostTestFactory() {
		t.Run(tt.n, func(t *testing.T) {
			r := Route{Host: tt.h}
			v, e := r.validHostPattern()
			if v != tt.v {
				al, _ := idna.ToASCII(tt.h)
				t.Errorf("u label host pattern %v, a label punycode: %v, valid: %v, expected: %v, cause: %v", tt.h, al, v, tt.v, e)
			}
		})
	}

}

func HostTestFactory() []struct {
	n string
	h string
	v bool
} {
	tests := []struct {
		n string
		h string
		v bool
	}{
		{n: "valid host", h: "blah", v: true},
		{n: "valid host unicode", h: "六书", v: true},
		{n: "valid wildcard dns umlaut", h: "*.faß.com", v: true},
		{n: "invalid wildcard pattern with double dot", h: "*..faß.com", v: false},
		{n: "invalid wildcard pattern with asterisk inside pattern", h: "a.*.faß.com", v: false},
		{n: "invalid wildcard pattern with valid regex style asterisk", h: "a*.faß.com", v: false},
		{n: "invalid wildcard pattern with invalid regex style asterisk", h: "*a.faß.com", v: false},
		{n: "valid wild dns other unicode", h: "*.😀😀😀.com", v: true},
		{n: "invalid asterisk in the middle of domain", h: "a.*.😀😀😀.com", v: false},
		{n: "valid cyrillic", h: "ԛәлп.com", v: true},
		{n: "invalid latin with stroke, case mapping is not part of IDNA 2008. We pass this anyway because go does", h: "Ⱥbby.com", v: true},
		{n: "DNS name can start with number (RFC-1123)", h: "1aaa.com", v: true},
		{n: "invalid ascii dollar sign as part of DNS name", h: "$1.a.com", v: false},
		{n: "invalid contains illegal ascii exclamation mark !", h: "!1.abc.com", v: false},
		{n: "invalid contains illegal ascii space", h: " 1.abc.com", v: false},
		{n: "invalid contains illegal ascii hash sign", h: "#1.abc.com", v: false},
		{n: "invalid contains illegal ascii percent sign", h: "%1.abc.com", v: false},
		{n: "invalid contains illegal ascii tilde", h: "^1.abc.com", v: false},
		{n: "invalid contains illegal ascii ampersand", h: "&1.abc.com", v: false},
		{n: "invalid contains illegal ascii (", h: "(1.abc.com", v: false},
		{n: "invalid contains illegal ascii )", h: ")1.abc.com", v: false},
		{n: "invalid contains illegal ascii ;", h: ";1.abc.com", v: false},
		{n: "invalid contains illegal ascii :", h: ":1.abc.com", v: false},
		{n: "invalid contains illegal ascii ,", h: ",1.abc.com", v: false},
		{n: "invalid contains illegal ascii ?", h: "?1.abc.com", v: false},
		{n: "invalid contains illegal ascii /", h: "/1.abc.com", v: false},
		{n: "invalid contains illegal ascii \\", h: "\\1.abc.com", v: false},
		{n: "invalid contains illegal ascii =", h: "=1.abc.com", v: false},
		{n: "invalid contains illegal ascii +", h: "+1.abc.com", v: false},
		{n: "invalid contains illegal ascii <", h: "<1.abc.com", v: false},
		{n: "invalid contains illegal ascii >", h: ">1.abc.com", v: false},
	}
	return tests
}

func TestHostDNSNamePatternCompiles(t *testing.T) {
	for _, tt := range HostTestFactory() {
		t.Run(tt.n, func(t *testing.T) {
			r := Route{Host: tt.h}
			e := r.compileHostPattern()
			compiled := e == nil && r.PunyHost != ""
			if compiled != tt.v {
				t.Errorf("host pattern u label %v, a label punycode: %v, compiled: %v, expected: %v, cause: %v", tt.h, r.PunyHost, compiled, tt.v, e)
			} else {
				t.Logf("host pattern u label %v, a label punycode: %v", tt.h, r.PunyHost)
			}
		})
	}

}

func TestHostDNSNamePatternMatchesHostHeader(t *testing.T) {
	tests := []struct {
		n  string
		h  string
		p  string
		hh string
		hp string
		v  bool
	}{
		{n: "fqdn identical host and pattern", h: "www.host.com", p: "/", hh: "www.host.com", hp: "/", v: true},
		{n: "fqdn identical host and pattern 2 (simple hostname)", h: "host", p: "/", hh: "host", hp: "/", v: true},
		{n: "fqdn identical host and pattern 3(simple hostname)", h: "foo", p: "/", hh: "bar", hp: "/", v: false},
		{n: "fqdn identical host and pattern 4", h: "host.tld", p: "/", hh: "host.tld", hp: "/", v: true},
		//we are not allowed to match this even if it's common. if a user wants to match host.com and www.host.com they need two entries
		{n: "fqdn common host and pattern mismatch", h: "host.com", p: "/", hh: "www.host.com", hp: "/", v: false},
		{n: "fqdn idns identical host and pattern", h: "www.host😀😀😀.com", p: "/", hh: "www.host😀😀😀.com", hp: "/", v: true},
		{n: "fqdn host and pattern non match", h: "www.foo.com", p: "/", hh: "www.bar.com", hp: "/", v: false},
		{n: "fqdn host and pattern non match 2", h: "www.foo.com", p: "/", hh: "www.fo.co", hp: "/", v: false},
		{n: "fqdn host and pattern non match 3", h: "www.foo.com", p: "/", hh: "www.foo.b.co", hp: "/", v: false},
		{n: "fqdn idns host and pattern non match", h: "www.foo.com", p: "/", hh: "www.bar😀😀😀.com", hp: "/", v: false},
		{n: "fqdn idns host and pattern non match 2", h: "www.bar😀.com", p: "/", hh: "www.bar.com", hp: "/", v: false},
		{n: "wildcard match", h: "*.bar.com", p: "/", hh: "www.bar.com", hp: "/", v: true},
		{n: "wildcard match", h: "*.baz.bar.com", p: "/", hh: "www.baz.bar.com", hp: "/", v: true},
		{n: "wildcard match", h: "*.baz.bar.com", p: "/", hh: "boo.baz.bar.com", hp: "/", v: true},
		{n: "wildcard non match", h: "*.bar.com", p: "/", hh: "www.uhoh.bar.com", hp: "/", v: false},
		{n: "idns wildcard match", h: "*.baz😀.com", p: "/", hh: "www.baz😀.com", hp: "/", v: true},
		{n: "idns wildcard match 2", h: "*.baz😀.com", p: "/", hh: "😀.baz😀.com", hp: "/", v: true},
	}

	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			r := Route{Host: tt.h, Path: tt.p}
			r.compilePath()
			r.compileHostPattern()
			got := r.match(requestFactory(tt.hp, tt.hh))
			if tt.v != got {
				t.Errorf("route host %v path %v request host %v path %v match expected %v got %v", r.Host, r.Path, tt.hh, tt.hp, tt.v, got)
			} else {
				t.Logf("route host %v path %v request host %v path %v match expected %v got %v", r.Host, r.Path, tt.hh, tt.hp, tt.v, got)
			}
		})
	}
}

func TestRouteSorting(t *testing.T) {
	// important. see sorting rules verified in sort oder below

	config := new(Config).readYmlFile("./j8acfg.yml")
	config = config.compileRoutePaths().
		compileRouteHosts().
		validateRoutes()

	//most qualified host has longest slug high up the inverted domain structure
	if config.Routes[0].Path != "/" ||
		config.Routes[0].PathType != "prefix" ||
		config.Routes[0].Host != "host123456789.com" ||

		//next longest slug has 5 chars on l2, then a subdomain
		config.Routes[1].Path != "/" ||
		config.Routes[1].PathType != "prefix" ||
		config.Routes[1].Host != "sub.host2.com" ||

		//next longest host has 5 chars on l2 but no subdomain. This order will match subdomains effectively
		config.Routes[2].Path != "/" ||
		config.Routes[2].PathType != "prefix" ||
		config.Routes[2].Host != "host3.com" ||

		//next longest host has 4 chars on l2. path is longer than [2] but this is ignored.
		config.Routes[3].Path != "/hostpath" ||
		config.Routes[3].PathType != "prefix" ||
		config.Routes[3].Host != "host.com" ||

		//no host, but longest exact path. Exact paths are matched before prefix.
		config.Routes[4].Path != "/path/med/1st" ||
		config.Routes[4].PathType != "exact" ||
		config.Routes[4].Host != "" ||

		//no host, but second longest exact path
		config.Routes[5].Path != "/path/med/2" ||
		config.Routes[5].PathType != "exact" ||
		config.Routes[5].Host != "" ||

		//no host, prefix PathType, longest first slug. Prefix matches after exact, so this is after [5]
		config.Routes[6].Path != "/longfirstslug_longfirstslug_longfirstslug_longfirstslug/short" ||
		config.Routes[6].PathType != "prefix" ||
		config.Routes[6].Host != "" ||

		//no host, prefix PathType, second longest slug.
		config.Routes[7].Path != "/badremote" ||
		config.Routes[7].PathType != "prefix" ||
		config.Routes[7].Host != "" {
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
				Path:     "/",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/",
				PathType: "prefix",
			},
		},
		{
			name: "one longer vs. one shorter wildcar routepath that isn't a slug",
			rl: Route{
				Path:     "/aabcccc*",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa*",
				PathType: "prefix",
			},
		},
		{
			name: "one longer vs. one shorter routepath that isn't a slug",
			rl: Route{
				Path:     "/aabcccc",
				PathType: "prefix",
			},
			rg: Route{
				Path:     "/aaa",
				PathType: "prefix",
			},
		},
		{
			name: "one vs. one routepath that isn't a slug",
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
			if !WeightedSlugs(NewRoutePath(tt.rl)).Less(WeightedSlugs(NewRoutePath(tt.rg))) {
				t.Errorf("routepath %v should be less than %v", tt.rl, tt.rg)
			}
		})
	}
}

func TestRouteHostPathLess(t *testing.T) {
	var tests = []struct {
		name string
		rl   Route
		rg   Route
	}{
		{
			name: "longer host should be first to allow overriding wildcards with explicit host names",
			rl: Route{
				Host:     "foo.bar.com",
				Path:     "/mse6jwtjwksbadrotate2",
				PathType: "prefix",
			},
			rg: Route{
				Host:     "*.bar.com",
				Path:     "/mse6jwtjwksbadrotatepath",
				PathType: "prefix",
			},
		},
		{
			name: "longer host should be first even if DNS components are different.",
			rl: Route{
				Host:     "foo.baz.com",
				Path:     "/a",
				PathType: "prefix",
			},
			rg: Route{
				Host:     "*.bar.com",
				Path:     "/a",
				PathType: "prefix",
			},
		},
		{
			name: "longer path should be first for same host name",
			rl: Route{
				Host:     "foo.bar.com",
				Path:     "/mse6jwtjwksbadrotate2",
				PathType: "prefix",
			},
			rg: Route{
				Host:     "foo.bar.com",
				Path:     "/",
				PathType: "prefix",
			},
		},
		{
			name: "fqdn host should be first compared to wildcard",
			rl: Route{
				Host:     "f.bar.com",
				Path:     "/m",
				PathType: "prefix",
			},
			rg: Route{
				Host:     "*.bar.com",
				Path:     "/m",
				PathType: "prefix",
			},
		},
		{
			name: "same host, exact path should with over prefix",
			rl: Route{
				Host:     "f.bar.com",
				Path:     "/a.html",
				PathType: "exact",
			},
			rg: Route{
				Host:     "f.bar.com",
				Path:     "/abcdefghijklm",
				PathType: "prefix",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//we need to do this here because config validate does not run and routes are not properly initialised.
			tt.rl.compilePath()
			tt.rl.compileHostPattern()
			tt.rg.compilePath()
			tt.rg.compileHostPattern()
			rs := Routes{tt.rl, tt.rg}
			sort.Sort(rs)
			if rs[0] != tt.rl {
				t.Errorf("route %v should be less than %v", tt.rl, tt.rg)
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
			if ok := route.match(requestFactory("/mse6")); ok {
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
			if ok := route.match(requestFactory("/s16")); ok {
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
