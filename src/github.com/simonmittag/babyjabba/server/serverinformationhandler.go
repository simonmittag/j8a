package server

import (
	"net/http"
)

func serverInformationHandler(w http.ResponseWriter, r *http.Request) {
	aboutString := "{\"version:\":\"" + Version + "\", \"serverID\":\"" + ID + "\"}"
	w.Write([]byte(aboutString))
}
