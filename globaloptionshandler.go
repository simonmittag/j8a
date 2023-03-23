package j8a

import "net/http"

const allow = "Allow"

func globalOptionsHandler(w http.ResponseWriter, r *http.Request) {
	proxy := new(Proxy).
		parseIncoming(r).
		setOutgoing(w)

	proxy.respondWith(200, "ok")
	proxy.writeStandardResponseHeaders()
	//send the allowable HTTP methods
	for _, method := range httpLegalMethods {
		w.Header().Add(allow, method)
	}

	//do we need this? there's no body
	proxy.Dwn.Resp.ContentEncoding = EncIdentity
	w.Header().Set(contentEncoding, string(proxy.Dwn.Resp.ContentEncoding))
	proxy.setContentLengthHeader()
	proxy.sendDownstreamStatusCodeHeader()

	logHandledDownstreamRoundtrip(proxy)
}
