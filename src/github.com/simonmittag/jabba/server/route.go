package server

import (
	"github.com/rs/zerolog/log"
)

//AboutJabba special Route alias for internal endpoint
const AboutJabba string = "aboutJabba"

//Route maps a Path to an upstream resource
type Route struct {
	Path   string
	Alias  string
	Policy string
}

func (route Route) mapUpstream() (*Upstream, string, bool) {
	var policy Policy
	var policyLabel string
	if len(route.Policy) > 0 {
		policy = Runner.Policies[route.Policy]
		policyLabel = policy.resolveLabel()
	}

	resource := Runner.Resources[route.Alias]
	//if a policy exists, we match resources with a label. TODO: this should be an interface
	if len(route.Policy) > 0 {
		for _, resourceMapping := range resource {
			for _, resourceLabel := range resourceMapping.Labels {
				if policyLabel == resourceLabel {
					log.Trace().Msgf("route %s mapped to upstream %s for label %s", route.Path, resourceMapping.Upstream, resourceLabel)
					return &resourceMapping.Upstream, policyLabel, true
				}
			}
		}
	} else {
		log.Trace().Msgf("route %s mapped to default upstream %s", route.Path, &resource[0].Upstream)
		return &resource[0].Upstream, "default", true
	}

	log.Trace().Msgf("route %s not mapped", route.Path)
	return nil, "default", false
}
