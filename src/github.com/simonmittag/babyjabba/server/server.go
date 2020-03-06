package server

import (
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

//Version is the server version
var Version string = "unknown"

//ID is a unique server ID
var ID string = "unknown"

var Port string

func initPort() {
	Port = os.Getenv("PORT")
	if len(Port) == 0 {
		Port = "8080"
	}
}

func BootStrap() {
	initPort()

	http.HandleFunc("/about", handleAbout)
	log.Info().Msgf("BabyJabba listening on port %s...", Port)
	err := http.ListenAndServe(":"+Port, nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to start HTTP(S) server on port %s, exiting...", Port)
		panic(err.Error())
	}
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	aboutString := "{\"version:\":\"" + Version + "\", \"serverID\":\"" + ID + "\"}"
	w.Write([]byte(aboutString))
}
