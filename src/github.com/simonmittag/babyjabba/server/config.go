package server

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
)

//Server config of all Routes mapped to Resources
type ServerConfig struct {
	Routes    []Route
	Resources []Resource
}

//Route maps a Path to an upstream resource
type Route struct {
	Path  string
	Alias string
	Label string
}

//Resource describes upstream servers
type Resource struct {
	Alias    string
	Labels   []string
	Upstream Upstream
}

//UPstream describes host mapping
type Upstream struct {
	Scheme string
	Host   string
	Port   int16
}

//ServerConfig stores global params
var Live ServerConfig

func parseFromFile() *ServerConfig {
	jsonFile, err := os.Open("babyjabba.json")
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
