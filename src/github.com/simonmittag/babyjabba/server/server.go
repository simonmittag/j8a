package server

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
)

//Version is the server version
var Version string = "unknown"

//ID is a unique server ID
var ID string = "unknown"

//Server struct defines runis the runtime environment for a config.
type Server struct {
	Config
}

//Runtime has access server config
var Runtime *Server

//BootStrap starts up the server from a ServerConfig
func BootStrap() {
	config := new(Config).
		parse("./babyjabba.json").
		reAliasResources().
		addDefaultPolicy()
	Runtime = &Server{Config: *config}
	Runtime.assignHandlers().
		startListening()
}

func (server Server) startListening() {
	log.Info().Msgf("BabyJabba listening on port %d...", server.Port)
	err := http.ListenAndServe(":"+strconv.Itoa(server.Port), nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to start HTTP(S) server on port %d, exiting...", server.Port)
		panic(err.Error())
	}
}

func (server Server) assignHandlers() Server {
	for _, route := range server.Routes {
		if route.Alias == AboutJabba {
			http.HandleFunc(route.Path, serverInformationHandler)
			log.Debug().Msgf("assigned internal server information handler to path %s", route.Path)
		}
	}
	http.HandleFunc("/", proxyHandler)
	log.Debug().Msgf("assigned proxy handler to path %s", "/")
	return server
}

func serverInformationHandler(w http.ResponseWriter, r *http.Request) {
	aboutString := "{\"version:\":\"" + Version + "\", \"serverID\":\"" + ID + "\"}"
	w.Write([]byte(aboutString))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	matched := false
	for _, route := range Runtime.Routes {
		if matched, _ = regexp.MatchString(route.Path, r.RequestURI); matched {
			upstream := route.mapUpstream()
			log.Info().Msgf("mapping %s to upstream %s", r.RequestURI, *upstream)
			w.Write([]byte(fmt.Sprintf("proxy request for upstream %v", *upstream)))
			break
		}
	}
	if !matched {
		w.Write([]byte(fmt.Sprintf("%v", "unmatched request")))
	}
}
