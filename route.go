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

type RoutePath []string

func NewRoutePath(r Route) RoutePath {
	rps := strings.Split(r.Path, slashS)
	return RoutePath(rps).trimEmptySlug()
}

func (rp RoutePath) trimEmptySlug() RoutePath {
	if len(rp[0]) == 0 && len(rp) > 1 {
		rp = rp[1:]
	}
	return rp
}

func (rp RoutePath) trimNextSlug() (RoutePath, error) {
	if len(rp) > 1 {
		rp = rp[1:]
		return rp, nil
	} else {
		return nil, errors.New("no more slugs")
	}
}

func (rp RoutePath) Less(rp2 RoutePath) bool {
	less := false
	if len(rp[0]) > len(rp2[0]) {
		less = true
	} else if len(rp[0]) == len(rp2[0]) {
		rpn, e := rp.trimNextSlug()
		rp2n, e2 := rp2.trimNextSlug()
		if e == nil && e2 == nil {
			less = rpn.Less(rp2n)
		} else if e2 != nil {
			less = true
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
	pis := NewRoutePath(s[i])
	pjs := NewRoutePath(s[j])

	less := false
	if (s[i].PathType == exact && s[j].PathType == exact) || (s[i].PathType == prefixS && s[j].PathType == prefixS) {
		less = pis.Less(pjs)
	} else {
		less = s[i].PathType == exact
	}
	return less
}

// Route maps a Path to an upstream resource
type Route struct {
	Host              string //idna host patterns
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
	val, err := p.ToUnicode(route.Host)
	if !(val == route.Host && err == nil) {
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

// TODO this matches a request to a route but it depends on sort order of multiple
// routes matched, it will match the first one.
const slashS = "/"

func (route Route) matchURI(request *http.Request) bool {
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
