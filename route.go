package j8a

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"time"
)

//Aboutj8a special Resource alias for internal endpoint
const about string = "about"

type Routes []Route

func (s Routes) Len() int {
	return len(s)
}
func (s Routes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Routes) Less(i, j int) bool {
	return len(s[i].Path) > len(s[j].Path)
}

//Route maps a Path to an upstream resource
type Route struct {
	Path      string
	PathRegex *regexp.Regexp
	Transform string
	Resource  string
	Policy    string
	Jwt       string
}

func (route Route) matchURI(request *http.Request) bool {
	matched := false
	if route.PathRegex != nil {
		matched = route.PathRegex.MatchString(request.RequestURI)
	} else {
		matched, _ = regexp.MatchString("^"+route.Path, request.RequestURI)
	}

	return matched
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
		return nil, "", false
	}
	//if a policy exists, we match resources with a label. TODO: this should be an interface

	if len(route.Policy) > 0 {
		for _, resourceMapping := range resource {
			for _, resourceLabel := range resourceMapping.Labels {
				if policyLabel == resourceLabel {
					var ev *zerolog.Event
					if proxy.XRequestInfo {
						ev = log.Info()
					} else {
						ev = log.Trace()
					}

					ev.Str(routeMsg, route.Path).
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
		log.Trace().
			Str(routeMsg, route.Path).
			Str(policyMsg, defaultMsg).
			Str(XRequestID, proxy.XRequestID).
			Str(upResource, resource[0].URL.String()).
			Msg(routeMapped)
		return &resource[0].URL, defaultMsg, true
	}

	log.Trace().
		Str(routeMsg, route.Path).
		Str(XRequestID, proxy.XRequestID).
		Msg(routeNotMapped)

	return nil, emptyString, false
}

func (route Route) hasJwt() bool {
	return len(route.Jwt) > 0
}
