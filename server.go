package j8a

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/process"
	golog "log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//Version is the server version
const Server string = "Server"

var Version string = "v0.9.3"

//ID is a unique server ID
var ID string = "unknown"

//Runtime struct defines runtime environment wrapper for a config.
type Runtime struct {
	Config
	Start             time.Time
	Memory            []sample
	AcmeHandler       *AcmeHandler
	ReloadableCert    *ReloadableCert
	cacheDir          string
	ConnectionWatcher ConnectionWatcher
}

type ReloadableCert struct {
	Cert *tls.Certificate
	mu   sync.Mutex
	Init bool
	//required to use runtime internally without global pointer for testing.
	runtime *Runtime
}

func (r *ReloadableCert) GetCertificateFunc(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return r.Cert, nil
}

func (r *ReloadableCert) triggerInit() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Init = true

	var cert tls.Certificate
	var err error

	c := []byte(r.runtime.Connection.Downstream.Tls.Cert)
	k := []byte(r.runtime.Connection.Downstream.Tls.Key)

	cert, err = tls.X509KeyPair(c, k)
	if err == nil {
		r.Cert = &cert
		r.Cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	}
	if err == nil {
		log.Debug().Msgf("TLS certificate #%v initialized", formatSerial(cert.Leaf.SerialNumber))
	}
	r.Init = false
	return err
}

type ConnectionWatcher struct {
	dwnOpenConns    int64
	dwnMaxOpenConns int64
	upOpenConns     int64
	upMaxOpenConns  int64
}

// OnStateChange records open connections in response to connection
// state changes. Set net/http Server.ConnState to this method
// as value.
func (cw *ConnectionWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		cw.AddDwn(1)
	case http.StateHijacked, http.StateClosed:
		cw.AddDwn(-1)
	}
	cw.UpdateMaxDwn(cw.DwnCount())
}

// Count returns the number of connections at the time
// the call.
func (cw *ConnectionWatcher) DwnCount() uint64 {
	return uint64(atomic.LoadInt64(&cw.dwnOpenConns))
}

func (cw *ConnectionWatcher) DwnMaxCount() uint64 {
	return uint64(atomic.LoadInt64(&cw.dwnMaxOpenConns))
}

// Add adds c to the number of active connections.
func (cw *ConnectionWatcher) AddDwn(c int64) {
	atomic.AddInt64(&cw.dwnOpenConns, c)
}

// Sets the maximum number of active connections observed
func (cw *ConnectionWatcher) UpdateMaxDwn(c uint64) {
	if c > cw.DwnMaxCount() {
		atomic.StoreInt64(&cw.dwnMaxOpenConns, int64(c))
	}
}

func (cw *ConnectionWatcher) UpCount() uint64 {
	return uint64(atomic.LoadInt64(&cw.upOpenConns))
}

func (cw *ConnectionWatcher) UpMaxCount() uint64 {
	return uint64(atomic.LoadInt64(&cw.upMaxOpenConns))
}

func (cw *ConnectionWatcher) AddUp(c int64) {
	atomic.AddInt64(&cw.upOpenConns, c)
}

func (cw *ConnectionWatcher) SetUp(c uint64) {
	atomic.StoreInt64(&cw.upOpenConns, int64(c))
}

func (cw *ConnectionWatcher) UpdateMaxUp(c uint64) {
	if c > cw.UpMaxCount() {
		atomic.StoreInt64(&cw.upMaxOpenConns, int64(c))
	}
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
		reApplyResourceURLDefaults().
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
		Config:            *config,
		Start:             time.Now(),
		AcmeHandler:       NewAcmeHandler(),
		ConnectionWatcher: ConnectionWatcher{dwnOpenConns: 0},
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
			e2 := os.Mkdir(myCacheDir, acmeRwx)
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

func (rt *Runtime) startListening() {
	readTimeoutDuration := time.Second * time.Duration(rt.Connection.Downstream.ReadTimeoutSeconds)
	roundTripTimeoutDuration := time.Second * time.Duration(rt.Connection.Downstream.RoundTripTimeoutSeconds)
	roundTripTimeoutDurationWithGrace := roundTripTimeoutDuration + (time.Second * 1)
	idleTimeoutDuration := time.Second * time.Duration(rt.Connection.Downstream.IdleTimeoutSeconds)

	log.Debug().
		Int64("dwnMaxBodyBytes", rt.Connection.Downstream.MaxBodyBytes).
		Float64("dwnReadTimeoutSeconds", readTimeoutDuration.Seconds()).
		Float64("dwnRoundTripTimeoutSeconds", roundTripTimeoutDuration.Seconds()).
		Float64("dwnIdleConnTimeoutSeconds", idleTimeoutDuration.Seconds()).
		Msg("server derived downstream params")

	httpConfig := &http.Server{
		Addr:              ":" + strconv.Itoa(rt.Connection.Downstream.Http.Port),
		ReadHeaderTimeout: readTimeoutDuration,               //downstream connection deadline
		ReadTimeout:       readTimeoutDuration,               //downstream connection deadline
		WriteTimeout:      roundTripTimeoutDurationWithGrace, //downstream connection deadline
		IdleTimeout:       idleTimeoutDuration,               //downstream connection deadline
		ErrorLog:          golog.New(&zerologAdapter{}, "", 0),
		Handler:           HandlerDelegate{},
		ConnState:         rt.ConnectionWatcher.OnStateChange,
	}

	//signal the WaitGroup that boot is over.
	Boot.Done()

	err := make(chan error)

	msg := fmt.Sprintf("j8a %s listener(s) init on", Version)
	if rt.isHTTPOn() {
		h := msg + fmt.Sprintf(" HTTP:%d...", rt.Connection.Downstream.Http.Port)
		go rt.startHTTP(httpConfig, err, h)
	}
	if rt.isTLSOn() {
		t := msg + fmt.Sprintf(" TLS:%d...", rt.Connection.Downstream.Tls.Port)
		tlsConfig := *httpConfig
		tlsConfig.Addr = ":" + strconv.Itoa(rt.Connection.Downstream.Tls.Port)
		go rt.startTls(&tlsConfig, err, t)
	}

	select {
	case sig := <-err:
		log.Fatal().Err(sig).Msg("... j8a exiting")
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
		cacheErr := runtime.loadAcmeCertAndKeyFromCache(p)
		if cacheErr != nil {
			//so caching didn't work let's go to acmeProvider
			acmeErr := runtime.fetchAcmeCertAndKey(acmeProviders[p])
			if acmeErr != nil {
				err <- acmeErr
				return
			} else {
				runtime.cacheAcmeCertAndKey(p)
			}
		}
	}

	cfg, tlsCfgErr := runtime.tlsConfig()
	if tlsCfgErr == nil {
		server.TLSConfig = cfg
	} else {
		err <- tlsCfgErr
		return
	}

	_, tlsErr := checkFullCertChain(runtime.ReloadableCert.Cert)
	if tlsErr == nil {
		go runtime.tlsHealthCheck(true)
		log.Info().Msg(msg)
		err <- server.ListenAndServeTLS("", "")
	} else {
		err <- tlsErr
	}
}

func (runtime *Runtime) startHTTP(server *http.Server, err chan<- error, msg string) {
	server.Addr = ":" + strconv.Itoa(runtime.Connection.Downstream.Http.Port)
	log.Info().Msg(msg)
	err <- server.ListenAndServe()
}

func (runtime *Runtime) initUserAgent() *Runtime {
	if httpClient == nil {
		httpClient = scaffoldHTTPClient(runtime)
	}
	return runtime
}

func (rt *Runtime) initStats() *Runtime {
	proc, _ := process.NewProcess(int32(os.Getpid()))
	rt.logRuntimeStats(proc)
	rt.logUptime()
	return rt
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

func (runtime *Runtime) tlsConfig() (*tls.Config, error) {
	//keypair and cert from the runtime params. They may have originated from the config file or ACME
	//in both instances the certificate now sits as reloadable in GetCertificateFunc which also uses Runner.
	var cert []byte = []byte(runtime.Connection.Downstream.Tls.Cert)
	var key []byte = []byte(runtime.Connection.Downstream.Tls.Key)

	//tls config validation
	if _, err := checkFullCertChainFromBytes(cert, key); err != nil {
		return nil, err
	}

	if err := runtime.ReloadableCert.triggerInit(); err != nil {
		return nil, err
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

	return config, nil
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

	proxy.writeStandardResponseHeaders()

	b := []byte(statusCodeResponse.AsJSON())
	proxy.Dwn.Resp.Body = &b
	if proxy.Dwn.AcceptEncoding.isCompatible(EncIdentity) {
		proxy.Dwn.Resp.ContentEncoding = EncIdentity
	} else if proxy.Dwn.AcceptEncoding.isCompatible(EncGzip) {
		proxy.Dwn.Resp.Body = Gzip(*proxy.Dwn.Resp.Body)
		proxy.Dwn.Resp.ContentEncoding = EncGzip
	} else if proxy.Dwn.AcceptEncoding.isCompatible(EncBrotli) {
		proxy.Dwn.Resp.Body = BrotliEncode(*proxy.Dwn.Resp.Body)
		proxy.Dwn.Resp.ContentEncoding = EncBrotli
	} else {
		//fallback
		proxy.Dwn.Resp.ContentEncoding = EncIdentity
	}

	if proxy.Dwn.Resp.StatusCode >= clientError {
		//for http1.1 we send a connection:close. Go HTTP/2 server removes this header which is illegal in HTTP/2.
		//but magically maps this to a GOAWAY frame for HTTP/2, see: https://go-review.googlesource.com/c/net/+/121415/
		proxy.Dwn.Resp.Writer.Header().Set(connectionS, closeS)
	}

	proxy.Dwn.Resp.Writer.Header().Set(contentType, applicationJSON)
	proxy.Dwn.Resp.Writer.Header().Set(contentEncoding, proxy.Dwn.Resp.ContentEncoding.print())
	proxy.setContentLengthHeader()
	proxy.Dwn.Resp.Writer.WriteHeader(proxy.Dwn.Resp.StatusCode)

	proxy.Dwn.Resp.Writer.Write(*proxy.Dwn.Resp.Body)

	logHandledDownstreamRoundtrip(proxy)
}
