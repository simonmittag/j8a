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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Version is the server version
const Server string = "Server"

var Version string = "v0.8.5"

//ID is a unique server ID
var ID string = "unknown"

//Runtime struct defines runtime environment wrapper for a config.
type Runtime struct {
	Config
	Start          time.Time
	Memory         []sample
	AcmeHandler    *AcmeHandler
	ReloadableCert *ReloadableCert
	cacheDir       string
}

type ReloadableCert struct {
	Cert *tls.Certificate
	Init bool
	mu   sync.Mutex
	//required to use runtime internally without global pointer for testing.
	runtime *Runtime
}

func (r *ReloadableCert) GetCertificateFunc(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var err error
	if r.Init == true || r.Cert == nil {
		c := []byte(r.runtime.Connection.Downstream.Tls.Cert)
		k := []byte(r.runtime.Connection.Downstream.Tls.Key)
		cert, err2 := tls.X509KeyPair(c, k)
		if err2 != nil {
			err = err2
		} else {
			r.Cert = &cert
			log.Debug().Msgf("TLS certificate init successful")
		}
		r.Init = false
	}
	return r.Cert, err
}

func (r *ReloadableCert) triggerInit() {
	r.Init = true
}

//Runner is the Live environment of the server
var Runner *Runtime

var Boot sync.WaitGroup = sync.WaitGroup{}

const tlsHandshakeError = "TLS handshake error"
const aboutPath = "/about"
const UpgradeHeader = "Upgrade"
const websocket = "websocket"
const strictTransportSecurity = "Strict-Transport-Security"
const maxAge31536000 = "max-age=31536000"

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
		validateJwt().
		compileRoutePaths().
		compileRouteTransforms().
		validateRoutes().
		addDefaultPolicy().
		setDefaultUpstreamParams().
		setDefaultDownstreamParams().
		validateHTTPConfig().
		validateAcmeConfig()

	Runner = &Runtime{
		Config:      *config,
		Start:       time.Now(),
		AcmeHandler: NewAcmeHandler(),
	}
	Runner.
		initCacheDir().
		initReloadableCert().
		initStats().
		initUserAgent().
		startListening()
}

const cacheDir = ".j8a"

func (r *Runtime) initCacheDir() *Runtime {
	home, e1 := os.UserHomeDir()
	if e1 == nil {
		myCacheDir := filepath.FromSlash(home + "/" + cacheDir)
		if _, e3 := os.Stat(myCacheDir); os.IsNotExist(e3) {
			e2 := os.Mkdir(myCacheDir, 0600)
			if e2 == nil {
				r.cacheDir = myCacheDir
				log.Debug().Msg("init cache dir in user home")
			}
		} else {
			r.cacheDir = myCacheDir
			log.Debug().Msg("found cache dir in user home")
		}
	}
	return r
}

func (r *Runtime) cacheDirIsActive() bool {
	return len(r.cacheDir) > 0
}

func (r *Runtime) initReloadableCert() *Runtime {
	r.ReloadableCert = &ReloadableCert{
		Cert:    nil,
		Init:    false,
		mu:      sync.Mutex{},
		runtime: r,
	}
	return r
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

	httpConfig := &http.Server{
		Addr:              ":" + strconv.Itoa(runtime.Connection.Downstream.Http.Port),
		ReadHeaderTimeout: readTimeoutDuration,               //downstream connection deadline
		ReadTimeout:       readTimeoutDuration,               //downstream connection deadline
		WriteTimeout:      roundTripTimeoutDurationWithGrace, //downstream connection deadline
		IdleTimeout:       idleTimeoutDuration,               //downstream connection deadline
		ErrorLog:          golog.New(&zerologAdapter{}, "", 0),
		Handler:           HandlerDelegate{},
	}

	//signal the WaitGroup that boot is over.
	Boot.Done()

	err := make(chan error)

	msg := fmt.Sprintf("j8a %s listener(s) init on", Version)
	if runtime.isHTTPOn() {
		h := msg + fmt.Sprintf(" HTTP:%d...", runtime.Connection.Downstream.Http.Port)
		go runtime.startHTTP(httpConfig, err, h)
	}
	if runtime.isTLSOn() {
		t := msg + fmt.Sprintf(" TLS:%d...", runtime.Connection.Downstream.Tls.Port)
		tlsConfig := *httpConfig
		tlsConfig.Addr = ":" + strconv.Itoa(runtime.Connection.Downstream.Tls.Port)
		go runtime.startTls(&tlsConfig, err, t)
	}

	select {
	case sig := <-err:
		log.Fatal().Err(sig).Msgf("... j8a exiting with ")
		panic(sig.Error())
	}
}

type HandlerDelegate struct{}

//TODO regex and perftest this function.
var acmeRex, _ = regexp.Compile("/.well-known/acme-challenge/")
var aboutRex, _ = regexp.Compile("^" + aboutPath + "$")

func (hd HandlerDelegate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if Runner.AcmeHandler.isActive() &&
		acmeRex.MatchString(r.RequestURI) {
		acmeHandler(w, r)
	} else if Runner.isHTTPOn() &&
		Runner.Connection.Downstream.Http.Redirecttls &&
		r.TLS == nil {
		redirectHandler(w, r)
	} else if r.ProtoMajor == 1 && r.Header.Get(UpgradeHeader) == websocket {
		websocketHandler(w, r)
		//TODO: this does not resolve whether about was actually configured in routes.
	} else if aboutRex.MatchString(r.RequestURI) {
		aboutHandler(w, r)
	} else {
		httpHandler(w, r)
	}
}

func (runtime *Runtime) startTls(server *http.Server, err chan<- error, msg string) {
	p := runtime.Connection.Downstream.Tls.Acme.Provider
	if len(p) > 0 {

		acmeErr := runtime.fetchAcmeCertAndKey(acmeProviders[p])
		if acmeErr != nil {
			err <- acmeErr
		}
	}
	server.TLSConfig = runtime.tlsConfig()

	_, tlsErr := checkCertChain(*runtime.ReloadableCert.Cert)
	if tlsErr == nil {
		go runtime.tlsHealthCheck(true)
		log.Info().Msg(msg)
		err <- server.ListenAndServeTLS("", "")
	} else {
		err <- tlsErr
	}
}

func (runtime Runtime) startHTTP(server *http.Server, err chan<- error, msg string) {
	server.Addr = ":" + strconv.Itoa(runtime.Connection.Downstream.Http.Port)
	log.Info().Msg(msg)
	err <- server.ListenAndServe()
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

	header.Set(Server, serverVersion())
	//for TLS response, we set HSTS header see RFC6797
	if Runner.isTLSOn() {
		header.Set(strictTransportSecurity, maxAge31536000)
	}
	//copy the X-REQUEST-ID from the request
	header.Set(XRequestID, proxy.XRequestID)
}

const j8a string = "j8a"

func serverVersion() string {
	return fmt.Sprintf("%s %s %s", j8a, Version, ID)
}

func (runtime *Runtime) tlsConfig() *tls.Config {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal().Msg("unable to parse TLS configuration, check your certificate and/or private key. j8a is exiting ...")
			os.Exit(-1)
		}
	}()

	//here we create a keypair from the runtime params. They may have originated from the config file or ACME
	//in both instances the certificate now sits as reloadable in GetCertificateFunc which also uses Runner.
	var cert []byte = []byte(runtime.Connection.Downstream.Tls.Cert)
	var key []byte = []byte(runtime.Connection.Downstream.Tls.Key)
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
		},
		GetCertificate: runtime.ReloadableCert.GetCertificateFunc,
	}

	//to init the cert.
	runtime.ReloadableCert.GetCertificateFunc(nil)

	return config
}

const contentType = "Content-Type"
const applicationJSON = "application/json"
const none = "none"

const connectionS = "Connection"
const closeS = "close"
const HTTP11 = "1.1"
const clientError = 400
const downstreamConnClose = "downstream connection close triggered for >=400 response code"

func sendStatusCodeAsJSON(proxy *Proxy) {
	statusCodeResponse := StatusCodeResponse{
		Code:    proxy.Dwn.Resp.StatusCode,
		Message: proxy.Dwn.Resp.Message,
	}

	if len(proxy.Dwn.Resp.Message) == 0 || proxy.Dwn.Resp.Message == none {
		statusCodeResponse.withCode(proxy.Dwn.Resp.StatusCode)
		proxy.Dwn.Resp.Message = statusCodeResponse.Message
	}

	if proxy.Dwn.Resp.SendGzip {
		proxy.Dwn.Resp.Body = Gzip(statusCodeResponse.AsJSON())
	} else {
		b := []byte(statusCodeResponse.AsJSON())
		proxy.Dwn.Resp.Body = &b
	}

	proxy.writeStandardResponseHeaders()

	if proxy.Dwn.Resp.StatusCode >= clientError {
		//for http1.1 we send a connection:close. Go HTTP/2 server removes this header which is illegal in HTTP/2.
		//but magically maps this to a GOAWAY frame for HTTP/2, see: https://go-review.googlesource.com/c/net/+/121415/
		proxy.Dwn.Resp.Writer.Header().Set(connectionS, closeS)
	}

	proxy.Dwn.Resp.Writer.Header().Set(contentType, applicationJSON)
	proxy.setContentLengthHeader()
	proxy.writeContentEncodingHeader()
	proxy.Dwn.Resp.Writer.WriteHeader(proxy.Dwn.Resp.StatusCode)

	proxy.Dwn.Resp.Writer.Write(*proxy.Dwn.Resp.Body)

	logHandledDownstreamRoundtrip(proxy)
}
