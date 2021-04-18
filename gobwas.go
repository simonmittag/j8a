package j8a

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	headerHost          = "Host"
	headerUpgrade       = "Upgrade"
	headerConnection    = "Connection"
	headerSecVersion    = "Sec-WebSocket-Version"
	headerSecProtocol   = "Sec-WebSocket-Protocol"
	headerSecExtensions = "Sec-WebSocket-Extensions"
	headerSecKey        = "Sec-WebSocket-Key"
	headerSecAccept     = "Sec-WebSocket-Accept"

	headerHostCanonical          = textproto.CanonicalMIMEHeaderKey(headerHost)
	headerUpgradeCanonical       = textproto.CanonicalMIMEHeaderKey(headerUpgrade)
	headerConnectionCanonical    = textproto.CanonicalMIMEHeaderKey(headerConnection)
	headerSecVersionCanonical    = textproto.CanonicalMIMEHeaderKey(headerSecVersion)
	headerSecProtocolCanonical   = textproto.CanonicalMIMEHeaderKey(headerSecProtocol)
	headerSecExtensionsCanonical = textproto.CanonicalMIMEHeaderKey(headerSecExtensions)
	headerSecKeyCanonical        = textproto.CanonicalMIMEHeaderKey(headerSecKey)
	headerSecAcceptCanonical     = textproto.CanonicalMIMEHeaderKey(headerSecAccept)
)

type RejectOption func(*rejectConnectionError)

func RejectionReason(reason string) RejectOption {
	return func(err *rejectConnectionError) {
		err.reason = reason
	}
}

// RejectionStatus returns an option that makes connection to be rejected with
// given HTTP status code.
func RejectionStatus(code int) RejectOption {
	return func(err *rejectConnectionError) {
		err.code = code
	}
}

var (
	ErrHandshakeBadProtocol = RejectConnectionError(
		RejectionStatus(http.StatusHTTPVersionNotSupported),
		RejectionReason(fmt.Sprintf("handshake error: bad HTTP protocol version")),
	)
	ErrHandshakeBadMethod = RejectConnectionError(
		RejectionStatus(http.StatusMethodNotAllowed),
		RejectionReason(fmt.Sprintf("handshake error: bad HTTP request method")),
	)
	ErrHandshakeBadHost = RejectConnectionError(
		RejectionStatus(http.StatusBadRequest),
		RejectionReason(fmt.Sprintf("handshake error: bad %q header", headerHost)),
	)
	ErrHandshakeBadUpgrade = RejectConnectionError(
		RejectionStatus(http.StatusBadRequest),
		RejectionReason(fmt.Sprintf("handshake error: bad %q header", headerUpgrade)),
	)
	ErrHandshakeBadConnection = RejectConnectionError(
		RejectionStatus(http.StatusBadRequest),
		RejectionReason(fmt.Sprintf("handshake error: bad %q header", headerConnection)),
	)
	ErrHandshakeBadSecAccept = RejectConnectionError(
		RejectionStatus(http.StatusBadRequest),
		RejectionReason(fmt.Sprintf("handshake error: bad %q header", headerSecAccept)),
	)
	ErrHandshakeBadSecKey = RejectConnectionError(
		RejectionStatus(http.StatusBadRequest),
		RejectionReason(fmt.Sprintf("handshake error: bad %q header", headerSecKey)),
	)
	ErrHandshakeBadSecVersion = RejectConnectionError(
		RejectionStatus(http.StatusBadRequest),
		RejectionReason(fmt.Sprintf("handshake error: bad %q header", headerSecVersion)),
	)
)

func httpGetHeader(h http.Header, key string) string {
	if h == nil {
		return ""
	}
	v := h[key]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// RejectConnectionError constructs an error that could be used to control the way
// handshake is rejected by Upgrader.
func RejectConnectionError(options ...RejectOption) error {
	err := new(rejectConnectionError)
	for _, opt := range options {
		opt(err)
	}
	return err
}

type HandshakeHeaderString string

func (s HandshakeHeaderString) WriteTo(w io.Writer) (int64, error) {
	n, err := io.WriteString(w, string(s))
	return int64(n), err
}

var ErrHandshakeUpgradeRequired = RejectConnectionError(
	RejectionStatus(http.StatusUpgradeRequired),
	RejectionHeader(HandshakeHeaderString(headerSecVersion+": 13\r\n")),
	RejectionReason(fmt.Sprintf("handshake error: bad %q header", headerSecVersion)),
)

var ErrMalformedRequest = RejectConnectionError(
	RejectionStatus(http.StatusBadRequest),
	RejectionReason("malformed HTTP request"),
)

// rejectConnectionError represents a rejection of upgrade error.
//
// It can be returned by Upgrader's On* hooks to control the way WebSocket
// handshake is rejected.
type rejectConnectionError struct {
	reason string
	code   int
	header ws.HandshakeHeader
}

// Error implements error interface.
func (r *rejectConnectionError) Error() string {
	return r.reason
}

func RejectionHeader(h ws.HandshakeHeader) RejectOption {
	return func(err *rejectConnectionError) {
		err.header = h
	}
}

func strHasToken(header, token string) (has bool) {
	return btsHasToken(strToBytes(header), strToBytes(token))
}

func btsHasToken(header, token []byte) (has bool) {
	httphead.ScanTokens(header, func(v []byte) bool {
		has = bytes.EqualFold(v, token)
		return !has
	})
	return
}

func strToBytes(str string) (bts []byte) {
	s := (*reflect.StringHeader)(unsafe.Pointer(&str))
	b := (*reflect.SliceHeader)(unsafe.Pointer(&bts))
	b.Data = s.Data
	b.Len = s.Len
	b.Cap = s.Len
	return
}

func strSelectProtocol(h string, check func(string) bool) (ret string, ok bool) {
	ok = httphead.ScanTokens(strToBytes(h), func(v []byte) bool {
		if check(btsToString(v)) {
			ret = string(v)
			return false
		}
		return true
	})
	return
}

func strSelectExtensions(h string, selected []httphead.Option, check func(httphead.Option) bool) ([]httphead.Option, bool) {
	return btsSelectExtensions(strToBytes(h), selected, check)
}

func btsSelectExtensions(h []byte, selected []httphead.Option, check func(httphead.Option) bool) ([]httphead.Option, bool) {
	s := httphead.OptionSelector{
		Flags: httphead.SelectUnique | httphead.SelectCopy,
		Check: check,
	}
	return s.Select(h, selected)
}

func btsToString(bts []byte) (str string) {
	return *(*string)(unsafe.Pointer(&bts))
}

var (
	// This variables are set like in net/net.go.
	// noDeadline is just zero value for readability.
	noDeadline = time.Time{}
	// aLongTimeAgo is a non-zero time, far in the past, used for immediate
	// cancelation of dials.
	aLongTimeAgo = time.Unix(42, 0)
)

const nonceSize = 24
const acceptSize = 28

type HandshakeHeader interface {
	io.WriterTo
}

type handshakeHeader [2]HandshakeHeader
type HandshakeHeaderHTTP http.Header

type writer struct {
	n int64
	w io.Writer
}

func (w *writer) WriteString(s string) (int, error) {
	n, err := io.WriteString(w.w, s)
	w.n += int64(n)
	return n, err
}

func (w *writer) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.n += int64(n)
	return n, err
}

func (h HandshakeHeaderHTTP) WriteTo(w io.Writer) (int64, error) {
	wr := writer{w: w}
	err := http.Header(h).Write(&wr)
	return wr.n, err
}

const (
	crlf          = "\r\n"
	colonAndSpace = ": "
	commaAndSpace = ", "
)

const (
	textHeadUpgrade = "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"
)

func httpWriteHeaderKey(bw *bufio.Writer, key string) {
	bw.WriteString(key)
	bw.WriteString(colonAndSpace)
}

func writeAccept(bw *bufio.Writer, nonce []byte) (int, error) {
	accept := make([]byte, acceptSize)
	initAcceptFromNonce(accept, nonce)
	// NOTE: write accept bytes as a string to prevent heap allocation –
	// WriteString() copy given string into its inner buffer, unlike Write()
	// which may write p directly to the underlying io.Writer – which in turn
	// will lead to p escape.
	return bw.WriteString(btsToString(accept))
}

func initAcceptFromNonce(accept, nonce []byte) {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	if len(accept) != acceptSize {
		panic("accept buffer is invalid")
	}
	if len(nonce) != nonceSize {
		panic("nonce is invalid")
	}

	p := make([]byte, nonceSize+len(magic))
	copy(p[:nonceSize], nonce)
	copy(p[nonceSize:], magic)

	sum := sha1.Sum(p)
	base64.StdEncoding.Encode(accept, sum[:])

	return
}

func httpWriteResponseUpgrade(bw *bufio.Writer, nonce []byte, hs ws.Handshake, header ws.HandshakeHeaderFunc) {
	bw.WriteString(textHeadUpgrade)

	httpWriteHeaderKey(bw, headerSecAccept)
	writeAccept(bw, nonce)
	bw.WriteString(crlf)

	if hs.Protocol != "" {
		httpWriteHeader(bw, headerSecProtocol, hs.Protocol)
	}
	if len(hs.Extensions) > 0 {
		httpWriteHeaderKey(bw, headerSecExtensions)
		httphead.WriteOptions(bw, hs.Extensions)
		bw.WriteString(crlf)
	}
	if header != nil {
		header(bw)
	}

	bw.WriteString(crlf)
}

func writeStatusText(bw *bufio.Writer, code int) {
	bw.WriteString("HTTP/1.1 ")
	bw.WriteString(strconv.Itoa(code))
	bw.WriteByte(' ')
	bw.WriteString(http.StatusText(code))
	bw.WriteString(crlf)
	bw.WriteString("Content-Type: text/plain; charset=utf-8")
	bw.WriteString(crlf)
}

func writeErrorText(bw *bufio.Writer, err error) {
	body := err.Error()
	bw.WriteString("Content-Length: ")
	bw.WriteString(strconv.Itoa(len(body)))
	bw.WriteString(crlf)
	bw.WriteString(crlf)
	bw.WriteString(body)
}

func statusText(code int) string {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	writeStatusText(bw, code)
	bw.Flush()
	return buf.String()
}

// errorText is a non-performant error text generator.
// NOTE: Used only to generate constants.
func errorText(err error) string {
	var buf bytes.Buffer
	bw := bufio.NewWriter(&buf)
	writeErrorText(bw, err)
	bw.Flush()
	return buf.String()
}

func httpWriteHeader(bw *bufio.Writer, key, value string) {
	httpWriteHeaderKey(bw, key)
	bw.WriteString(value)
	bw.WriteString(crlf)
}

var (
	textHeadBadRequest          = statusText(http.StatusBadRequest)
	textHeadInternalServerError = statusText(http.StatusInternalServerError)
	textHeadUpgradeRequired     = statusText(http.StatusUpgradeRequired)

	textTailErrHandshakeBadProtocol   = errorText(ErrHandshakeBadProtocol)
	textTailErrHandshakeBadMethod     = errorText(ErrHandshakeBadMethod)
	textTailErrHandshakeBadHost       = errorText(ErrHandshakeBadHost)
	textTailErrHandshakeBadUpgrade    = errorText(ErrHandshakeBadUpgrade)
	textTailErrHandshakeBadConnection = errorText(ErrHandshakeBadConnection)
	textTailErrHandshakeBadSecAccept  = errorText(ErrHandshakeBadSecAccept)
	textTailErrHandshakeBadSecKey     = errorText(ErrHandshakeBadSecKey)
	textTailErrHandshakeBadSecVersion = errorText(ErrHandshakeBadSecVersion)
	textTailErrUpgradeRequired        = errorText(ErrHandshakeUpgradeRequired)
)


func httpWriteResponseError(bw *bufio.Writer, err error, code int, header ws.HandshakeHeaderFunc) {
	switch code {
	case http.StatusBadRequest:
		bw.WriteString(textHeadBadRequest)
	case http.StatusInternalServerError:
		bw.WriteString(textHeadInternalServerError)
	case http.StatusUpgradeRequired:
		bw.WriteString(textHeadUpgradeRequired)
	default:
		writeStatusText(bw, code)
	}

	// Write custom headers.
	if header != nil {
		header(bw)
	}

	switch err {
	case ErrHandshakeBadProtocol:
		bw.WriteString(textTailErrHandshakeBadProtocol)
	case ErrHandshakeBadMethod:
		bw.WriteString(textTailErrHandshakeBadMethod)
	case ErrHandshakeBadHost:
		bw.WriteString(textTailErrHandshakeBadHost)
	case ErrHandshakeBadUpgrade:
		bw.WriteString(textTailErrHandshakeBadUpgrade)
	case ErrHandshakeBadConnection:
		bw.WriteString(textTailErrHandshakeBadConnection)
	case ErrHandshakeBadSecAccept:
		bw.WriteString(textTailErrHandshakeBadSecAccept)
	case ErrHandshakeBadSecKey:
		bw.WriteString(textTailErrHandshakeBadSecKey)
	case ErrHandshakeBadSecVersion:
		bw.WriteString(textTailErrHandshakeBadSecVersion)
	case ErrHandshakeUpgradeRequired:
		bw.WriteString(textTailErrUpgradeRequired)
	case nil:
		bw.WriteString(crlf)
	default:
		writeErrorText(bw, err)
	}
}

func (hs handshakeHeader) WriteTo(w io.Writer) (n int64, err error) {
	for i := 0; i < len(hs) && err == nil; i++ {
		if h := hs[i]; h != nil {
			var m int64
			m, err = h.WriteTo(w)
			n += m
		}
	}
	return n, err
}

type MyUpgrader ws.HTTPUpgrader

func (u MyUpgrader) Upgrade(r *http.Request, w http.ResponseWriter) (conn net.Conn, rw *bufio.ReadWriter, hs ws.Handshake, err error) {
	// Hijack connection first to get the ability to write rejection errors the
	// same way as in Upgrader.
	hj, ok := w.(http.Hijacker)
	if ok {
		conn, rw, err = hj.Hijack()
	} else {
		err = errors.New("500 - unable to hijack connection")
	}
	if err != nil {
		//httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// See https://tools.ietf.org/html/rfc6455#section-4.1
	// The method of the request MUST be GET, and the HTTP version MUST be at least 1.1.
	var nonce string
	if r.Method != http.MethodGet {
		err = ErrHandshakeBadMethod
	} else if r.ProtoMajor < 1 || (r.ProtoMajor == 1 && r.ProtoMinor < 1) {
		err = ErrHandshakeBadProtocol
	} else if r.Host == "" {
		err = ErrHandshakeBadHost
	} else if u := httpGetHeader(r.Header, headerUpgradeCanonical); u != "websocket" && !strings.EqualFold(u, "websocket") {
		err = ErrHandshakeBadUpgrade
	} else if c := httpGetHeader(r.Header, headerConnectionCanonical); c != "Upgrade" && !strHasToken(c, "upgrade") {
		err = ErrHandshakeBadConnection
	} else if nonce = httpGetHeader(r.Header, headerSecKeyCanonical); len(nonce) != nonceSize {
		err = ErrHandshakeBadSecKey
	} else if v := httpGetHeader(r.Header, headerSecVersionCanonical); v != "13" {
		// According to RFC6455:
		//
		// If this version does not match a version understood by the server,
		// the server MUST abort the WebSocket handshake described in this
		// section and instead send an appropriate HTTP error code (such as 426
		// Upgrade Required) and a |Sec-WebSocket-Version| header field
		// indicating the version(s) the server is capable of understanding.
		//
		// So we branching here cause empty or not present version does not
		// meet the ABNF rules of RFC6455:
		//
		// version = DIGIT | (NZDIGIT DIGIT) |
		// ("1" DIGIT DIGIT) | ("2" DIGIT DIGIT)
		// ; Limited to 0-255 range, with no leading zeros
		//
		// That is, if version is really invalid – we sent 426 status, if it
		// not present or empty – it is 400.
		if v != "" {
			err = ErrHandshakeUpgradeRequired
		} else {
			err = ErrHandshakeBadSecVersion
		}
	}
	if check := u.Protocol; err == nil && check != nil {
		ps := r.Header[headerSecProtocolCanonical]
		for i := 0; i < len(ps) && err == nil && hs.Protocol == ""; i++ {
			var ok bool
			hs.Protocol, ok = strSelectProtocol(ps[i], check)
			if !ok {
				err = ErrMalformedRequest
			}
		}
	}
	if check := u.Extension; err == nil && check != nil {
		xs := r.Header[headerSecExtensionsCanonical]
		for i := 0; i < len(xs) && err == nil; i++ {
			var ok bool
			hs.Extensions, ok = strSelectExtensions(xs[i], hs.Extensions, check)
			if !ok {
				err = ErrMalformedRequest
			}
		}
	}

	// Clear deadlines set by server.
	conn.SetDeadline(noDeadline)
	if t := u.Timeout; t != 0 {
		conn.SetWriteDeadline(time.Now().Add(t))
		defer conn.SetWriteDeadline(noDeadline)
	}

	var header handshakeHeader
	if h := u.Header; h != nil {
		header[0] = HandshakeHeaderHTTP(h)
	}
	if err == nil {
		httpWriteResponseUpgrade(rw.Writer, strToBytes(nonce), hs, header.WriteTo)
		err = rw.Writer.Flush()
	} else {
		var code int
		if rej, ok := err.(*rejectConnectionError); ok {
			code = rej.code
			header[1] = rej.header
		}
		if code == 0 {
			code = http.StatusInternalServerError
		}
		httpWriteResponseError(rw.Writer, err, code, header.WriteTo)
		// Do not store Flush() error to not override already existing one.
		rw.Writer.Flush()
	}
	return
}
