package server

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

//AboutJabba special Route alias for internal endpoint
const AboutJabba string = "aboutJabba"

//Route maps a Path to an upstream resource
type Route struct {
	Path  string
	Alias string
	Label string
}

func (route Route) mapUpstream() *Upstream {
	for _, resource := range Live.Resources {
		if route.Alias == resource.Alias {
			if len(route.Label) > 0 {
				for _, label := range resource.Labels {
					if label == route.Label {
						log.Debug().Msgf("mapped route %s to upstream %s", route.Path, resource.Upstream)
						return &resource.Upstream
					}
				}
				msg := fmt.Sprintf("configuration error. invalid route %v unable to map resource", route)
				log.Fatal().Msg(msg)
				panic(msg)
			} else {
				log.Debug().Msgf("mapped route %s to upstream %s", route.Path, resource.Upstream)
				return &resource.Upstream
			}
		}
	}
	msg := fmt.Sprintf("configuration error. invalid route")
	log.Fatal().Msg(msg)
	panic(msg)
}
