package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

//Version is the server version
var Version string = "unknown"

//ID is a unique server ID
var ID string = "unknown"

var Port string

func initPort() {
	Port = os.Getenv("PORT")
	if len(Port) == 0 {
		Port = "8080"
	}
}

func BootStrap() {
	parseFromFile()
	initPort()
	for _, route := range Live.Routes {
		assignHandler(route)
	}

	log.Info().Msgf("BabyJabba listening on port %s...", Port)
	err := http.ListenAndServe(":"+Port, nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to start HTTP(S) server on port %s, exiting...", Port)
		panic(err.Error())
	}
}

func assignHandler(route Route) {
	if route.Alias == AboutJabba {
		http.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			aboutString := "{\"version:\":\"" + Version + "\", \"serverID\":\"" + ID + "\"}"
			w.Write([]byte(aboutString))
		})
		log.Debug().Msgf("assigned internal server information handler to path %s", route.Path)
	} else {
		http.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			upstream := mapUpstream(route)
			w.Write([]byte(fmt.Sprintf("%v", upstream)))
		})
		log.Debug().Msgf("assigned proxy handler to path %s", route.Path)
	}

}

func mapUpstream(route Route) *Upstream {
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
	msg := fmt.Sprintf("configuration error. invalid route %v")
	log.Fatal().Msg(msg)
	panic(msg)
}
