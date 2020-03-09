package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

//Version is the server version
var Version string = "unknown"

//ID is a unique server ID
var ID string = "unknown"

//BootStrap starts up the server from a ServerConfig
func BootStrap() {
	config := parse("./babyjabba.json")
	for _, route := range config.Routes {
		assignHandler(route)
	}
	startListening(config)
}

func startListening(config *Config) {
	log.Info().Msgf("BabyJabba listening on port %d...", config.Port)
	err := http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to start HTTP(S) server on port %d, exiting...", config.Port)
		panic(err.Error())
	}
}

func assignHandler(route Route) {
	if route.Alias == AboutJabba {
		http.HandleFunc(route.Path, serverInformationHandler)
		log.Debug().Msgf("assigned internal server information handler to path %s", route.Path)
	} else {
		http.HandleFunc(route.Path, proxyHandler)
		log.Debug().Msgf("assigned proxy handler to path %s", route.Path)
	}

}

func serverInformationHandler(w http.ResponseWriter, r *http.Request) {
	aboutString := "{\"version:\":\"" + Version + "\", \"serverID\":\"" + ID + "\"}"
	w.Write([]byte(aboutString))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: find route from request, then map to upstream
	w.Write([]byte(fmt.Sprintf("%v", "Hello World")))
}
