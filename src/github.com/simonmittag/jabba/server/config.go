package server

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
)

//Config is the system wide configuration for Jabba
type Config struct {
	Policies   map[string]Policy
	Routes     []Route
	Resources  map[string][]ResourceMapping
	Connection Connection
}

func (config Config) parse(file string) *Config {
	jsonFile, err := os.Open(file)
	defer jsonFile.Close()
	if err != nil {
		msg := "cannot find config file or unable to read server configuration, exiting..."
		log.Fatal().Msg(msg)
		panic(msg)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &config)
	//todo tell me about more of the config, number of routes
	log.Debug().Msgf("parsed server configuration with %d live routes", len(config.Routes))
	return &config
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

func (config Config) setDefaultTimeouts() *Config {
	//Downstrea params
	if config.Connection.Downstream.ReadTimeoutSeconds == 0 {
		config.Connection.Downstream.ReadTimeoutSeconds = 120
	}
	if config.Connection.Downstream.RoundTripTimeoutSeconds == 0 {
		config.Connection.Downstream.RoundTripTimeoutSeconds = 240
	}
	if config.Connection.Downstream.IdleTimeoutSeconds == 0 {
		config.Connection.Downstream.IdleTimeoutSeconds = 120
	}

	//Client params
	if config.Connection.Upstream.SocketTimeoutSeconds == 0 {
		config.Connection.Upstream.SocketTimeoutSeconds = 3
	}
	if config.Connection.Upstream.ReadTimeoutSeconds == 0 {
		config.Connection.Upstream.ReadTimeoutSeconds = 120
	}
	if config.Connection.Upstream.IdleTimeoutSeconds == 0 {
		config.Connection.Upstream.IdleTimeoutSeconds = 120
	}
	if config.Connection.Upstream.PoolSize == 0 {
		config.Connection.Upstream.PoolSize = 32768
	}
	if config.Connection.Upstream.MaxAttempts == 0 {
		config.Connection.Upstream.MaxAttempts = 1
	}
	return &config
}
