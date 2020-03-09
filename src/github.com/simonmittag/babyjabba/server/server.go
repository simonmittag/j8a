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
	mapAboutJabba("/chuba")
	for _, route := range Live.Routes {
		upstream := mapResource(route)
		log.Debug().Msgf("mapped route %s to upstream %s", route.Path, upstream)
	}

	log.Info().Msgf("BabyJabba listening on port %s...", Port)
	err := http.ListenAndServe(":"+Port, nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to start HTTP(S) server on port %s, exiting...", Port)
		panic(err.Error())
	}
}

func mapAboutJabba(path string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		aboutString := "{\"version:\":\"" + Version + "\", \"serverID\":\"" + ID + "\"}"
		w.Write([]byte(aboutString))
	})
	log.Debug().Msgf("mapped route %s", path)
}

func mapResource(route Route) *Upstream {
	for _, resource := range Live.Resources {
		if route.Alias == resource.Alias {
			if len(route.Label) > 0 {
				for _, label := range resource.Labels {
					if label == route.Label {
						return &resource.Upstream
					}
				}
				msg := fmt.Sprintf("configuration error. invalid route %v unable to map resource", route)
				log.Fatal().Msg(msg)
				panic(msg)
			} else {
				return &resource.Upstream
			}
		}
	}
	msg := fmt.Sprintf("configuration error. invalid route %v")
	log.Fatal().Msg(msg)
	panic(msg)
}
