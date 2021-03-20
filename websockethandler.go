package j8a

import "net/http"

func websocketHandler(response http.ResponseWriter, request *http.Request) {
	proxyHandler(response, request, upgrade)
}

func upgrade(proxy *Proxy) {
	//
}