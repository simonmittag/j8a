package jabba

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

//Version is the server version
var Version string = "v0.3.6"

//ID is a unique server ID
var ID string = "unknown"

//Runtime struct defines runtime environment wrapper for a config.
type Runtime struct {
	Config
}

//Runner is the Live environment of the server
var Runner *Runtime

var Boot sync.WaitGroup = sync.WaitGroup{}

//BootStrap starts up the server from a ServerConfig
func BootStrap() {
	initLogger()

	config := new(Config).
		read("./jabba.json").
		reApplyResourceNames().
		addDefaultPolicy().
		setDefaultValues()

	Runner = &Runtime{Config: *config}
	Runner.initStats().
		initUserAgent().
		assignHandlers().
		startListening()
}

func (runtime Runtime) startListening() {
	readTimeoutDuration := time.Second * time.Duration(runtime.Connection.Downstream.ReadTimeoutSeconds)
	writeTimeoutDuration := time.Second * time.Duration(runtime.Connection.Downstream.RoundTripTimeoutSeconds)
	idleTimeoutDuration := time.Second * time.Duration(runtime.Connection.Downstream.IdleTimeoutSeconds)

	log.Debug().
		Float64("downstreamReadTimeoutSeconds", readTimeoutDuration.Seconds()).
		Float64("downstreamWriteTimeoutSeconds", writeTimeoutDuration.Seconds()).
		Float64("downstreamIdleConnTimeoutSeconds", idleTimeoutDuration.Seconds()).
		Msg("server derived downstream params")
	log.Info().Msgf("Jabba %s listening on port %d...", Version, runtime.Connection.Downstream.Port)

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(runtime.Connection.Downstream.Port),
		Handler:      nil,
		ReadTimeout:  readTimeoutDuration,
		WriteTimeout: writeTimeoutDuration,
		IdleTimeout:  idleTimeoutDuration,
	}

	//signal the WaitGroup that boot is over.
	Boot.Done()

	//this line blocks execution and the server stays up
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to start HTTP server on port %d, exiting...", runtime.Connection.Downstream.Port)
		panic(err.Error())
	}
}

func (runtime Runtime) assignHandlers() Runtime {
	for _, route := range runtime.Routes {
		if route.Resource == AboutJabba {
			http.HandleFunc(route.Path, aboutHandler)
			log.Debug().Msgf("assigned about handler to path %s", route.Path)
		}
	}
	http.HandleFunc("/", proxyHandler)
	log.Debug().Msgf("assigned proxy handler to path %s", "/")
	return runtime
}

func (runtime Runtime) initUserAgent() Runtime {
	if httpClient == nil {
		httpClient = scaffoldHTTPClient(runtime)
	}
	return runtime
}

func (runtime Runtime) initStats() Runtime {
	go stats(os.Getpid())
	return runtime
}

//TODO: this really needs to be configurable. it adds a lot of options to every response otherewise.
func writeStandardResponseHeaders(proxy *Proxy) {
	header := proxy.Dwn.Resp.Writer.Header()

	header.Set("Server", fmt.Sprintf("Jabba %s %s", Version, ID))
	header.Set("Cache-control:", "no-store, no-cache, must-revalidate, proxy-revalidate")
	//for TLS response, we set HSTS header see RFC6797
	if Runner.Connection.Downstream.Mode == "TLS" {
		header.Set("Strict-Transport-Security", "max-age=31536000")
	}
	header.Set("X-xss-protection", "1;mode=block")
	header.Set("X-content-type-options", "nosniff")
	header.Set("X-frame-options", "sameorigin")
	//copy the X-REQUEST-ID from the request
	header.Set(XRequestID, proxy.XRequestID)
}

func sendStatusCodeAsJSON(proxy *Proxy) {

	if proxy.Dwn.Resp.StatusCode > 399 {
		log.Warn().Int("downstreamResponseCode", proxy.Dwn.Resp.StatusCode).
			Str("downstreamResponseMessage", proxy.Dwn.Resp.Message).
			Str("path", proxy.Dwn.Req.URL.Path).
			Str(XRequestID, proxy.XRequestID).
			Str("method", proxy.Dwn.Method).
			Msgf("request not served")
	}
	writeStandardResponseHeaders(proxy)
	proxy.Dwn.Resp.Writer.Header().Set("Content-Type", "application/json")

	proxy.Dwn.Resp.Writer.WriteHeader(proxy.Dwn.Resp.StatusCode)

	statusCodeResponse := StatusCodeResponse{
		Code:       proxy.Dwn.Resp.StatusCode,
		Message:    proxy.Dwn.Resp.Message,
		XRequestID: proxy.XRequestID,
	}
	if proxy.Dwn.Resp.SendGzip {
		proxy.Dwn.Resp.Writer.Write(Gzip(statusCodeResponse.AsJSON()))
	} else {
		proxy.Dwn.Resp.Writer.Write(statusCodeResponse.AsJSON())
	}
}
