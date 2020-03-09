package server

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
)

//Live ServerConfig stores global params
var Live Config

//ServerConfig of all Routes mapped to Resources
type Config struct {
	Mode      string
	Port      int
	Routes    []Route
	Resources []Resource
}

func parse(file string) *Config {
	jsonFile, err := os.Open(file)
	defer jsonFile.Close()
	if err != nil {
		msg := "cannot find babyjabba.json, unable to read server configuration, exiting..."
		log.Fatal().Msg(msg)
		panic(msg)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &Live)
	//todo tell me about more of the config, number of routes
	log.Debug().Msgf("parsed server configuration with %d live routes", len(Live.Routes))
	return &Live
}
