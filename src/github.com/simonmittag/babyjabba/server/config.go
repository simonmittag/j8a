package server

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
)

//Live ServerConfig stores global params
var Live Config

//Config is the system wide configuration for BabyJabba
type Config struct {
	Mode      string
	Port      int
	Policies  map[string]Policy
	Routes    []Route
	Resources map[string][]ResourceMapping
}

func (config Config) parse(file string) *Config {
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

func (config Config) reAliasResources() *Config {
	for alias := range config.Resources {
		resourceMappings := config.Resources[alias]
		for i, resourceMapping := range resourceMappings {
			resourceMapping.Alias = alias
			resourceMappings[i] = resourceMapping
		}
	}
	return &config
}

func (config Config) addDefaultPolicy() *Config {
	defaultPolicy := new(Policy)
	lw := LabelWeight{
		Label:  "default",
		Weight: 1.0,
	}
	var labelWeights []LabelWeight
	labelWeights = append(labelWeights, lw)
	*defaultPolicy = labelWeights
	config.Policies["default"] = *defaultPolicy
	return &config
}
