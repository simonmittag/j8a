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
			upstream, mapped := route.mapUpstream()
			if mapped {
				handleUpstreamRequest(w, r, upstream)
			} else {
				sendStatusCodeAsJSON(w, r, 503)
				// return
			}
			break
		}
	}
	if !matched {
		sendStatusCodeAsJSON(w, r, 404)
	}
}

func writeStandardHeaders(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Server", "BabyJabba "+Version)
	w.Header().Set("Content-Encoding", "identity")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-control:", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("X-server-id", ID)
	w.Header().Set("X-xss-protection", "1;mode=block")
	w.Header().Set("X-content-type-options", "nosniff")
	w.Header().Set("X-frame-options", "sameorigin")
	w.WriteHeader(statusCode)
}

func handleUpstreamRequest(w http.ResponseWriter, r *http.Request, u *Upstream) {
	//handle request by sending upstream
	writeStandardHeaders(w, 200)
	w.Write([]byte(fmt.Sprintf("proxy request for upstream %v", *u)))
}

func sendStatusCodeAsJSON(w http.ResponseWriter, r *http.Request, statusCode int) {
	if statusCode >= 299 {
		log.Warn().Msgf("request path %v code %d", r.URL.Path, statusCode)
	}
	writeStandardHeaders(w, statusCode)
	w.Write([]byte(fmt.Sprintf("{ \"code\":\"%v\" }", strconv.Itoa(statusCode))))
}
