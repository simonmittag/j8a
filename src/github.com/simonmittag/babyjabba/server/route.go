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

func (route Route) mapUpstream() (*Upstream, bool) {
	var policy Policy
	var policyLabel string
	if len(route.Policy) > 0 {
		policy = Runtime.Policies[route.Policy]
		policyLabel = policy.resolveLabel()
	}

	resource := Runtime.Resources[route.Alias]
	if len(route.Policy) > 0 {
		for _, resourceMapping := range resource {
			for _, resourceLabel := range resourceMapping.Labels {
				if policyLabel == resourceLabel {
					log.Debug().Msgf("route %s mapped to upstream %s for label %s", route.Path, resourceMapping.Upstream, resourceLabel)
					return &resourceMapping.Upstream, true
				}
			}
		}
	} else {
		log.Debug().Msgf("route %s mapped to default upstream %s", route.Path, &resource[0].Upstream)
		return &resource[0].Upstream, true
	}

	log.Debug().Msgf("route %s not mapped", route.Path)
	return nil, false
}
