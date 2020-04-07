package server

import (
	"github.com/rs/zerolog/log"
)

//AboutJabba special Route alias for internal endpoint
const AboutJabba string = "aboutJabba"

//Route maps a Path to an upstream resource
type Route struct {
	Path     string
	Resource string
	Policy   string
}

// maps a route to a URL. Returns the URL, the name of the mapped policy and whether mapping was successful
func (route Route) mapURL() (*URL, string, bool) {
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
					log.Trace().Msgf("route %s mapped to upstream %s for label %s", route.Path, resourceMapping.URL, resourceLabel)
					return &resourceMapping.URL, policyLabel, true
				}
			}
		}
	} else {
		log.Trace().Msgf("route %s mapped to default upstream %s", route.Path, &resource[0].URL)
		return &resource[0].URL, "default", true
	}

	log.Trace().Msgf("route %s not mapped", route.Path)
	return nil, "default", false
}
