package jabba

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"time"
)

//AboutJabba special Route alias for internal endpoint
const AboutJabba string = "aboutJabba"

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
	Path     string
	Regex    *regexp.Regexp
	Resource string
	Policy   string
}

func (route Route) matchURI(request *http.Request) bool {
	matched := false
	if route.Regex != nil {
		matched = route.Regex.MatchString(request.RequestURI)
	} else {
		matched, _ = regexp.MatchString("^"+route.Path, request.RequestURI)
	}

	return matched
}

// maps a route to a URL. Returns the URL, the name of the mapped policy and whether mapping was successful
func (route Route) mapURL(proxy *Proxy) (*URL, string, bool) {
	var policy Policy
	var policyLabel string
	if len(route.Policy) > 0 {
		policy = Runner.Policies[route.Policy]
		policyLabel = policy.resolveLabel()
	}

	resource := Runner.Resources[route.Resource]
	//if a policy exists, we match resources with a label. TODO: this should be an interface
	if len(route.Policy) > 0 {
		for _, resourceMapping := range resource {
			for _, resourceLabel := range resourceMapping.Labels {
				if policyLabel == resourceLabel {
					log.Trace().
						Str("routePath", route.Path).
						Str("upstream", resourceMapping.URL.String()).
						Str("label", resourceLabel).
						Str("policy", route.Policy).
						Str(XRequestID, proxy.XRequestID).
						Int64("downstreamElapsedMillis", time.Since(proxy.Dwn.startDate).Milliseconds()).
						Msg("upstream route mapped")
					return &resourceMapping.URL, policyLabel, true
				}
			}
		}
	} else {
		log.Trace().
			Str("routePath", route.Path).
			Str("policy", "default").
			Str(XRequestID, proxy.XRequestID).
			Str("upstream", resource[0].URL.String()).
			Msg("route mapped")
		return &resource[0].URL, "default", true
	}

	log.Trace().
		Str("routePath", route.Path).
		Str(XRequestID, proxy.XRequestID).
		Msg("route not mapped")

	return nil, "", false
}
