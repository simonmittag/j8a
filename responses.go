package j8a

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var huttese = []string{
	"Achuta!",
	"Goodde da lodia!",
	"H'chu apenkee!",
	"Chuba!",
	"Ka Cheesa Crispa Greedo?",
	"De wanna wanga?",
	"Peedunkee, caba dee unko!",
}

var httpResponses = map[int]string{
	100: "continue",
	101: "switching protocols",
	102: "processing",
	103: "early hints",
	200: "ok",
	201: "created",
	202: "accepted",
	203: "non-authoritative information",
	204: "no content",
	205: "reset content",
	206: "partial content",
	207: "multi-status",
	208: "already reported",
	226: "IM used",
	300: "multiple choices",
	301: "moved permanently",
	302: "found",
	303: "see other",
	304: "not modified",
	305: "use proxy",
	306: "switch proxy",
	307: "temporary redirect",
	308: "permanent redirect",
	400: "bad request",
	401: "unauthorized",
	402: "payment required",
	403: "forbidden",
	404: "not found",
	405: "method not allowed",
	406: "not acceptable",
	407: "proxy authentication required",
	408: "request timeout",
	409: "conflict",
	410: "gone",
	411: "length required",
	412: "precondition failed",
	413: "payload too large",
	414: "URI too long",
	415: "unsupported media type",
	416: "requested range not satisfiable",
	417: "expectation failed",
	418: "i'm a teapot",
	420: "enhance your calm",
	421: "misdirected request",
	422: "unprocessable entity",
	423: "locked",
	424: "failed dependency",
	425: "too early",
	426: "upgrade required",
	428: "precondition required",
	429: "too many requests",
	431: "request header fields too large",
	444: "no response by upstream nginx",
	449: "retry with",
	450: "blocked by windows parental controls",
	451: "unavailable for legal reasons",
	499: "client closed request",
	500: "internal server error",
	501: "not implemented",
	502: "bad gateway",
	503: "service unavailable",
	504: "gateway timeout",
	505: "http version not supported",
	506: "variant also negotiates",
	507: "insufficient storage",
	508: "loop detected",
	509: "bandwidth limit exceeded",
	510: "not extended",
	511: "network authentication required",
	598: "network read timeout error",
	599: "network connect timeout error",
}

//AboutResponse exposes standard environment
type AboutResponse struct {
	J8a      string
	ServerID string
	Version  string
}

//StatusCodeResponse defines a JSON structure for a canned HTTP response
type StatusCodeResponse struct {
	AboutResponse
	Code       int
	Message    string
	XRequestID string
}

//AsJSON renders the status Code response into a JSON string as []byte
func (aboutResponse AboutResponse) AsJSON() []byte {
	aboutResponse.ServerID = ID
	aboutResponse.Version = Version
	aboutResponse.J8a = randomHuttese()
	response, _ := json.Marshal(aboutResponse)
	return response
}

func (aboutResponse AboutResponse) AsString() string {
	return strings.ToLower(string(aboutResponse.AsJSON()))
}

func (statusCodeResponse *StatusCodeResponse) withCode(code int) {
	statusCodeResponse.Code = code
	if msg, ok := httpResponses[code]; ok {
		statusCodeResponse.Message = msg
	} else {
		statusCodeResponse.Message = "unspecified response code"
	}
}

//AsJSON renders the status Code response into a JSON string as []byte
func (statusCodeResponse StatusCodeResponse) AsJSON() []byte {
	statusCodeResponse.ServerID = ID
	statusCodeResponse.Version = Version
	statusCodeResponse.J8a = randomHuttese()
	response, _ := json.Marshal(statusCodeResponse)
	//typo fix so we can continue to use json.Marshal which needs Uppercase struct props
	response[2] = 0x6a
	return response
}

func (statusCodeResponse StatusCodeResponse) AsString() string {
	return strings.ToLower(string(statusCodeResponse.AsJSON()))
}

func randomHuttese() string {
	rand.Seed(time.Now().Unix())
	return huttese[rand.Int()%len(huttese)]
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	proxy := new(Proxy).
		parseIncoming(r).
		setOutgoing(w)

	proxy.writeStandardResponseHeaders()

	ae := r.Header["Accept-Encoding"]
	res := AboutResponse{}.AsJSON()
	w.Header().Set("Content-Type", "application/json")
	if len(ae) > 0 {
		s := strings.Join(ae, " ")
		if strings.Contains(s, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			w.Write(Gzip(res))
		} else {
			w.Header().Set("Content-Encoding", "identity")
			w.Write(res)
		}
	} else {
		w.Header().Set("Content-Encoding", "identity")
		w.Write(res)
	}

	logHandledDownstreamRoundtrip(proxy)
}
