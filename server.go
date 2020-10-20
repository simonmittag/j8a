package j8a

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
	golog "log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Version is the server version
var Version string = "v0.6.14"

//ID is a unique server ID
var ID string = "unknown"

//Runtime struct defines runtime environment wrapper for a config.
type Runtime struct {
	Config
	Start  time.Time
	Memory []sample
}

//Runner is the Live environment of the server
var Runner *Runtime

var Boot sync.WaitGroup = sync.WaitGroup{}

const tlsHandshakeError = "TLS handshake error"

type zerologAdapter struct {
	ipr iprex
}

func (zla *zerologAdapter) Write(p []byte) (n int, err error) {
	msg := string(p)
	if strings.Contains(msg, tlsHandshakeError) {
		log.Warn().
			Str("netEvt", msg).
			Str("dwnReqRemoteAddr", zla.ipr.extractAddr(msg)).
			Msg("TLS handshake error")
	} else {
		log.Trace().
			Str("netEvt", msg).
			Str("dwnReqRemoteAddr", zla.ipr.extractAddr(msg)).
			Msg("undetermined network event")
	}
	return len(p), nil
}

//BootStrap starts up the server from a ServerConfig
func BootStrap() {
	initLogger()

	config := new(Config).
		load().
		reApplyResourceSchemes().
		reApplyResourceNames().
		compileRoutePaths().
		sortRoutes().
		addDefaultPolicy().
		setDefaultUpstreamParams().
		setDefaultDownstreamParams()

	Runner = &Runtime{
		Config: *config,
		Start:  time.Now(),
	}
	Runner.initStats().
		initUserAgent().
		startListening()
}

func (runtime Runtime) startListening() {
	readTimeoutDuration := time.Second * time.Duration(runtime.Connection.Downstream.ReadTimeoutSeconds)
	roundTripTimeoutDuration := time.Second * time.Duration(runtime.Connection.Downstream.RoundTripTimeoutSeconds)
	roundTripTimeoutDurationWithGrace := roundTripTimeoutDuration + (time.Second * 1)
	idleTimeoutDuration := time.Second * time.Duration(runtime.Connection.Downstream.IdleTimeoutSeconds)

	log.Debug().
		Int64("dwnMaxBodyBytes", runtime.Connection.Downstream.MaxBodyBytes).
		Float64("dwnReadTimeoutSeconds", readTimeoutDuration.Seconds()).
		Float64("dwnRoundTripTimeoutSeconds", roundTripTimeoutDuration.Seconds()).
		Float64("dwnIdleConnTimeoutSeconds", idleTimeoutDuration.Seconds()).
		Msg("server derived downstream params")

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(runtime.Connection.Downstream.Port),
		ReadHeaderTimeout: readTimeoutDuration,
		ReadTimeout:       readTimeoutDuration,
		WriteTimeout:      roundTripTimeoutDurationWithGrace,
		IdleTimeout:       idleTimeoutDuration,
		Handler:           runtime.mapPathsToHandler(),
		ErrorLog:          golog.New(&zerologAdapter{}, "", 0),
	}

	//signal the WaitGroup that boot is over.
	Boot.Done()

	var err error

	if runtime.isTLSMode() {
		server.TLSConfig = runtime.tlsConfig()
		_, tlsErr := checkCertChain(server.TLSConfig.Certificates[0])
		if tlsErr == nil {
			go tlsHealthCheck(server.TLSConfig, true)
			log.Info().Msgf("j8a %s listening in TLS mode on port %d...", Version, runtime.Connection.Downstream.Port)
			err = server.ListenAndServeTLS("", "")
		} else {
			err = tlsErr
		}
	} else {
		log.Info().Msgf("j8a %s listening in HTTP mode on port %d...", Version, runtime.Connection.Downstream.Port)
		err = server.ListenAndServe()
	}

	if err != nil {
		log.Fatal().Err(err).Msgf("unable to start j8a on port %d, exiting...", runtime.Connection.Downstream.Port)
		panic(err.Error())
	}
}

func (runtime Runtime) mapPathsToHandler() http.Handler {
	//TODO: do we need this handler with two handlerfuncs or can we map all requests to one handlerfunc to speed up?
	//if one handlerfunc in the system, it would need to distinguish between /about and other routes.

	handler := http.NewServeMux()
	for _, route := range runtime.Routes {
		if route.Resource == about {
			handler.Handle(route.Path, http.HandlerFunc(aboutHandler))
			log.Debug().Msgf("assigned about handler to path %s", route.Path)
		}
	}
	handler.Handle("/", http.HandlerFunc(proxyHandler))
	log.Debug().Msgf("assigned proxy handler to path %s", "/")

	return handler
}

func (runtime Runtime) initUserAgent() Runtime {
	if httpClient == nil {
		httpClient = scaffoldHTTPClient(runtime)
	}
	return runtime
}

func (runtime Runtime) initStats() Runtime {
	proc, _ := process.NewProcess(int32(os.Getpid()))
	logProcStats(proc)
	logUptime()
	return runtime
}

func (proxy *Proxy) writeStandardResponseHeaders() {
	header := proxy.Dwn.Resp.Writer.Header()

	header.Set("Server", fmt.Sprintf("j8a %s %s", Version, ID))
	//for TLS response, we set HSTS header see RFC6797
	if Runner.isTLSMode() {
		header.Set("Strict-Transport-Security", "max-age=31536000")
	}
	//copy the X-REQUEST-ID from the request
	header.Set(XRequestID, proxy.XRequestID)
}

func (runtime Runtime) tlsConfig() *tls.Config {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal().Msg("unable to parse TLS configuration, check your certificate and/or private key. j8a is exiting ...")
			os.Exit(-1)
		}
	}()

	//here we create a keypair from the PEM string in the config file
	var cert []byte = []byte(runtime.Connection.Downstream.Cert)
	var key []byte = []byte(runtime.Connection.Downstream.Key)
	chain, _ := tls.X509KeyPair(cert, key)

	var nocert error
	chain.Leaf, nocert = x509.ParseCertificate(chain.Certificate[0])
	if nocert != nil {
		panic("unable to parse malformed or missing x509 certificate.")
	}

	//now create the TLS config.
	config := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			//TLS 1.3 good ciphers
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			//TLS 1.2 good ciphers
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			//TLS 1.2 weak ciphers for IE11, Safari 6-8. We keep this for backwards compatibility with older
			//clients, it still gives us an A+ result on: https://www.ssllabs.com/ssltest/analyze.html?d=j8a.io
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		},
		Certificates: []tls.Certificate{chain},
	}

	return config
}

func sendStatusCodeAsJSON(proxy *Proxy) {
	proxy.writeStandardResponseHeaders()
	proxy.Dwn.Resp.Writer.Header().Set("Content-Type", "application/json")
	proxy.writeContentEncodingHeader()

	proxy.Dwn.Resp.Writer.WriteHeader(proxy.Dwn.Resp.StatusCode)

	statusCodeResponse := StatusCodeResponse{
		Code:    proxy.Dwn.Resp.StatusCode,
		Message: proxy.Dwn.Resp.Message,
	}

	if len(proxy.Dwn.Resp.Message) == 0 || proxy.Dwn.Resp.Message == "none" {
		statusCodeResponse.withCode(proxy.Dwn.Resp.StatusCode)
		proxy.Dwn.Resp.Message = statusCodeResponse.Message
	}

	if proxy.Dwn.Resp.SendGzip {
		proxy.Dwn.Resp.Writer.Write(*Gzip(statusCodeResponse.AsJSON()))
	} else {
		proxy.Dwn.Resp.Writer.Write(statusCodeResponse.AsJSON())
	}

	logHandledDownstreamRoundtrip(proxy)
}
