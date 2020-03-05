package server

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

//Version is the server version
var Version string = "unknown"

//ID is a unique server ID
var ID string = "unknown"

func BootStrap() {
	http.HandleFunc("/about", handleAbout)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal().Msg("uh oh")
		panic(err.Error())
	}
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	version := "1"
	aboutString := "{\"version:\":\"" + version + "\"}"
	w.Write([]byte(aboutString))
}
