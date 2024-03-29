package j8a

import (
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	urlverifier "github.com/davidmytton/url-verifier"
	"golang.org/x/net/idna"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Aboutj8a special Resource alias for internal endpoint
const about string = "about"

type WeightedSlugs []string
type RoutePath WeightedSlugs
type DNSNameComponents WeightedSlugs

func NewRoutePath(r Route) RoutePath {
	rps := strings.Split(r.Path, slashS)
	return RoutePath(WeightedSlugs(rps).trimEmptySlug())
}

func NewDNSNameComponents(r Route) DNSNameComponents {
	rhs := strings.Split(r.PunyHost, dot)
	for i, j := 0, len(rhs)-1; i < j; i, j = i+1, j-1 {
		rhs[i], rhs[j] = rhs[j], rhs[i]
	}
	return DNSNameComponents(WeightedSlugs(rhs).trimEmptySlug())
}

func (w WeightedSlugs) trimEmptySlug() WeightedSlugs {
	if len(w[0]) == 0 && len(w) > 1 {
		w = w[1:]
	}
	return w
}

func (w WeightedSlugs) trimNextSlug() (WeightedSlugs, error) {
	if len(w) > 1 {
		w = w[1:]
		return w, nil
	} else {
		return nil, errors.New("no more slugs")
	}
}

func (w WeightedSlugs) Less(w2 WeightedSlugs) bool {
	less := false
	if len(w[0]) > len(w2[0]) {
		less = true
	} else if len(w[0]) == len(w2[0]) {
		wn, e := w.trimNextSlug()
		w2n, e2 := w2.trimNextSlug()
		if e == nil && e2 == nil {
			less = wn.Less(w2n)
		} else if e2 != nil {
			/////aaarrggghhh this fixes an issue because we share sorting slug paths and DNS name segments with the same alg
			if w[0] == STAR {
				less = false
			} else {
				less = true
			}
		}
	}
	return less
}

type Routes []Route

func (s Routes) Len() int {
	return len(s)
}
func (s Routes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Routes) Less(i, j int) bool {
	if s[i].PunyHost == s[j].PunyHost {
		return s.PathIsLess(i, j)
	} else {
		return s.HostIsLess(i, j)
	}
}

func (s Routes) HostIsLess(i int, j int) bool {
	return WeightedSlugs(NewDNSNameComponents(s[i])).Less(WeightedSlugs(NewDNSNameComponents(s[j])))
}

func (s Routes) PathIsLess(i int, j int) bool {
	pis := NewRoutePath(s[i])
	pjs := NewRoutePath(s[j])

	less := false
	if (s[i].PathType == exact && s[j].PathType == exact) || (s[i].PathType == prefixS && s[j].PathType == prefixS) {
		less = WeightedSlugs(pis).Less(WeightedSlugs(pjs))
	} else {
		less = s[i].PathType == exact
	}
	return less
}

// Route maps a Path to an upstream resource
type Route struct {
	Host              string         //idna host pattern
	PunyHost          string         //punycode host pattern
	CompiledPunyHost  *regexp.Regexp // as regex
	Path              string
	PathType          string // exact | prefix
	CompiledPathRegex *regexp.Regexp
	Transform         string
	Resource          string
	Policy            string
	Jwt               string
}

const wildcard = "*"

func (route *Route) validHostPattern() (bool, error) {
	//first check the name is a valid idna name.
	p := idna.New(
		idna.ValidateLabels(true),
		//this has to be off it disallows * for registration
		//idna.ValidateForRegistration(),
		idna.StrictDomainName(true))
	_, err := p.ToUnicode(route.Host)
	if err != nil {
		return false, err
	} else {
		a, err := p.ToASCII(route.Host)
		if err != nil {
			return false, err
		}

		for i, j := range strings.Split(a, ".") {
			if len(j) == 0 {
				return false, errors.New("dns name segments cannot be empty string")
			}
			if j == wildcard && i != 0 {
				return false, errors.New("wildcard can only be at far left of domain name")
			}
			if strings.Contains(j, wildcard) && len(j) > 1 {
				return false, errors.New("wildcard can only be used for entire subdomains, not regex style")
			}
		}

		//now perform DNS name validation on ascii format.
		v2 := govalidator.IsDNSName(a)
		if v2 {
			return true, nil
		} else {
			//this validator does not pass wildcard names so we have to
			if strings.HasPrefix(a, wildcard) {
				return true, nil
			} else {
				return false, errors.New("not a valid DNS name after idna normalisation " + a)
			}
		}
	}
}

func (route *Route) compileHostPattern() error {
	if b, e := route.validHostPattern(); b {
		al, e1 := idna.ToASCII(route.Host)
		if e1 == nil {
			route.PunyHost = al
			re, _ := regexp.Compile(startS + al)
			route.CompiledPunyHost = re
		}
		return e1
	} else {
		return e
	}
}

func (route *Route) validPath() (bool, error) {
	const fakeHost = "http://127.0.0.1"
	defaultError := errors.New(fmt.Sprintf("route %v not a valid URL path", route.Path))

	_, err := url.ParseRequestURI(fakeHost + route.Path)
	if err != nil {
		return false, defaultError
	}
	_, err = url.Parse(fakeHost + route.Path)
	if err != nil {
		return false, defaultError
	}
	_, err = urlverifier.NewVerifier().Verify(fakeHost + route.Path)
	if err != nil {
		return false, defaultError
	}
	if len(route.Path) == 0 {
		return false, defaultError
	}
	if strings.Contains(route.Path, " ") {
		return false, errors.New(fmt.Sprintf("route %v not a valid URL path, may not contain space character", route.Path))
	}
	if strings.Index(route.Path, "/") != 0 {
		return false, errors.New(fmt.Sprintf("route %v not a valid URL path, does not start with '/'", route.Path))
	}
	if e := route.compilePath(); e != nil {
		return false, e
	}
	return true, nil
}

const startS = "^"
const dollarS = "$"
const exact = "exact"

func (route *Route) compilePath() error {
	compileMe := route.Path
	if string(compileMe[0]) != startS {
		compileMe = startS + compileMe
	}
	if strings.EqualFold(exact, route.PathType) {
		compileMe = compileMe + dollarS
	}
	var err error
	route.CompiledPathRegex, err = regexp.Compile(compileMe)
	return err
}

const slashS = "/"

func (route Route) match(request *http.Request) bool {
	if len(route.PunyHost) > 0 {
		return route.matchHostHeader(request) &&
			route.matchURIPath(request)
	} else {
		return route.matchURIPath(request)
	}
}

func (route Route) matchHostHeader(request *http.Request) bool {
	//safety measure in case we got a unicode host header
	al, _ := idna.ToASCII(request.Host)
	return route.CompiledPunyHost.MatchString(al) &&
		len(strings.Split(route.PunyHost, ".")) == len(strings.Split(request.Host, "."))
}

func (route Route) matchURIPath(request *http.Request) bool {
	match := route.CompiledPathRegex.MatchString(request.URL.Path)
	if !match &&
		route.PathType == prefixS &&
		len(request.URL.Path) > 0 &&
		string(request.URL.Path[len(request.URL.Path)-1]) != slashS {
		match = route.CompiledPathRegex.MatchString(request.URL.Path + slashS)
	}
	return match
}

// Deprecated
func (route Route) matchURI_Naive(request *http.Request) bool {
	return route.CompiledPathRegex.MatchString(request.URL.Path)
}

const upstreamResourceMapped = "upstream resource mapped"
const policyMsg = "policy"
const upResource = "upResource"
const routeMsg = "route"
const labelMsg = "label"
const defaultMsg = "default"
const routeMapped = "route mapped"
const routeNotMapped = "route not mapped"
const emptyString = ""

// maps a route to a URL. Returns the URL, the name of the mapped policy and whether mapping was successful
func (route Route) mapURL(proxy *Proxy) (*URL, string, bool) {
	var policy Policy
	var policyLabel string
	if len(route.Policy) > 0 {
		policy = Runner.Policies[route.Policy]
		policyLabel = policy.resolveLabel()
	}

	resource := Runner.Resources[route.Resource]
	if resource == nil {
		return nil, emptyString, false
	}
	//if a policy exists, we match resources with a label. TODO: this should be an interface

	if len(route.Policy) > 0 {
		for _, resourceMapping := range resource {
			for _, resourceLabel := range resourceMapping.Labels {
				if policyLabel == resourceLabel {
					infoOrTraceEv(proxy).Str(routeMsg, route.Path).
						Str(upResource, resourceMapping.URL.String()).
						Str(labelMsg, resourceLabel).
						Str(policyMsg, route.Policy).
						Str(XRequestID, proxy.XRequestID).
						Int64(dwnElpsdMicros, time.Since(proxy.Dwn.startDate).Microseconds()).
						Msg(upstreamResourceMapped)
					return &resourceMapping.URL, policyLabel, true
				}
			}
		}
	} else {
		infoOrTraceEv(proxy).
			Str(routeMsg, route.Path).
			Str(policyMsg, defaultMsg).
			Str(XRequestID, proxy.XRequestID).
			Str(upResource, resource[0].URL.String()).
			Msg(routeMapped)
		return &resource[0].URL, defaultMsg, true
	}

	infoOrTraceEv(proxy).
		Str(routeMsg, route.Path).
		Str(XRequestID, proxy.XRequestID).
		Msg(routeNotMapped)

	return nil, emptyString, false
}

func (route Route) hasJwt() bool {
	return len(route.Jwt) > 0
}

type RoutePathTypes []string

func NewRoutePathTypes() RoutePathTypes {
	return RoutePathTypes([]string{"exact", "prefix"})
}

func (r RoutePathTypes) isValid(t string) bool {
	m := false
	for _, rp := range r {
		if strings.EqualFold(t, rp) {
			m = true
		}
	}
	return m
}
