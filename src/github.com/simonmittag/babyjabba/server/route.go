package server

import (
	"fmt"

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

func (route Route) mapUpstream() *Upstream {
	var policy Policy
	var policyLabel string
	if len(route.Policy) > 0 {
		policy = Runtime.Policies[route.Policy]
		policyLabel = policy.resolveLabel()
	}
	for _, resource := range Runtime.Resources {
		if route.Alias == resource.Alias {
			if len(route.Policy) > 0 {
				for _, resourceLabel := range resource.Labels {
					if policyLabel == resourceLabel {
						log.Debug().Msgf("route %s mapped to upstream %s", route.Path, resource.Upstream)
						return &resource.Upstream
					}
				}
			} else {
				log.Debug().Msgf("route %s mapped to upstream %s", route.Path, resource.Upstream)
				return &resource.Upstream
			}
		}
	}
	msg := fmt.Sprintf("route %v invalid", route)
	log.Warn().Msg(msg)
	panic(msg)
}
