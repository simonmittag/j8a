package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

//Version is the server version
var Version string = "unknown"

//ID is a unique server ID
var ID string = "unknown"

//Runtime struct defines runtime environment wrapper for a config.
type Runtime struct {
	Config
}

//Runner is the Live environment of the server
var Runner *Runtime

//BootStrap starts up the server from a ServerConfig
func BootStrap() {
	config := new(Config).
		parse("./jabba.json").
		reAliasResources().
		addDefaultPolicy().
		setDefaultTimeouts()
	Runner = &Runtime{Config: *config}
	Runner.assignHandlers().
		startListening()
}

func (runtime Runtime) startListening() {
	log.Info().Msgf("Jabba listening on port %d...", runtime.Port)
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(runtime.Port),
		Handler: nil,
		ReadTimeout: time.Second * time.Duration(runtime.
			Connection.
			Server.
			ReadTimeoutSeconds),
		WriteTimeout: time.Second * time.Duration(runtime.
			Connection.
			Server.
			WriteTimeoutSeconds),
	}

	//this line blocks execution and the server stays up
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to start HTTP server on port %d, exiting...", runtime.Port)
		panic(err.Error())
	}
}

func (runtime Runtime) assignHandlers() Runtime {
	for _, route := range runtime.Routes {
		if route.Alias == AboutJabba {
			http.HandleFunc(route.Path, serverInformationHandler)
			log.Debug().Msgf("assigned internal server information handler to path %s", route.Path)
		}
	}
	http.HandleFunc("/", proxyHandler)
	log.Debug().Msgf("assigned proxy handler to path %s", "/")
	return runtime
}

func writeStandardResponseHeaders(response http.ResponseWriter, request *http.Request, statusCode int) {
	response.Header().Set("Server", "BabyJabba "+Version)
	response.Header().Set("Content-Encoding", "identity")
	response.Header().Set("Cache-control:", "no-store, no-cache, must-revalidate, proxy-revalidate")
	//for TLS response, we set HSTS header see RFC6797
	if Runner.Mode == "TLS" {
		response.Header().Set("Strict-Transport-Security", "max-age=31536000")
	}
	response.Header().Set("X-server-id", ID)
	response.Header().Set("X-xss-protection", "1;mode=block")
	response.Header().Set("X-content-type-options", "nosniff")
	response.Header().Set("X-frame-options", "sameorigin")
	//copy the X-REQUEST-ID from the request
	response.Header().Set(X_REQUEST_ID, request.Header.Get(X_REQUEST_ID))

	//status code must be last, no headers may be written after this one.
	response.WriteHeader(statusCode)
}

func sendStatusCodeAsJSON(response http.ResponseWriter, request *http.Request, statusCode int, message string) {
	if statusCode >= 299 {
		log.Warn().Int("downstreamResponseCode", statusCode).
			Str("path", request.URL.Path).
			Str(X_REQUEST_ID, request.Header.Get(X_REQUEST_ID)).
			Msgf("request not served")
	}
	writeStandardResponseHeaders(response, request, statusCode)
	response.Header().Set("Content-Type", "application/json")
	statusCodeResponse := StatusCodeResponse{
		Code:       statusCode,
		Message:    message,
		XRequestID: request.Header.Get(X_REQUEST_ID),
	}
	response.Write(statusCodeResponse.AsJSON())
}
