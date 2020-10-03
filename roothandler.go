package j8a

import "net/http"

type RootHandler struct{}

func (rh RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI=="/about" {
		aboutHandler(w,r)
	} else {
		proxyHandler(w,r)
	}
}