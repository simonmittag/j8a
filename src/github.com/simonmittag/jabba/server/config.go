package server

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog/log"
)

//Config is the system wide configuration for Jabba
type Config struct {
	Policies map[string]Policy
	Routes   []Route
	Rsrc     map[string][]ResourceMapping
	Cnx      Connection
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

func (config Config) reApplyResourceNames() *Config {
	for name := range config.Rsrc {
		resourceMappings := config.Rsrc[name]
		for i, resourceMapping := range resourceMappings {
			resourceMapping.Name = name
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
	if config.Cnx.Dwn.ReadTimeoutSeconds == 0 {
		config.Cnx.Dwn.ReadTimeoutSeconds = 120
	}
	if config.Cnx.Dwn.RoundTripTimeoutSeconds == 0 {
		config.Cnx.Dwn.RoundTripTimeoutSeconds = 240
	}
	if config.Cnx.Dwn.IdleTimeoutSeconds == 0 {
		config.Cnx.Dwn.IdleTimeoutSeconds = 120
	}

	//Client params
	if config.Cnx.Up.SocketTimeoutSeconds == 0 {
		config.Cnx.Up.SocketTimeoutSeconds = 3
	}
	if config.Cnx.Up.ReadTimeoutSeconds == 0 {
		config.Cnx.Up.ReadTimeoutSeconds = 120
	}
	if config.Cnx.Up.IdleTimeoutSeconds == 0 {
		config.Cnx.Up.IdleTimeoutSeconds = 120
	}
	if config.Cnx.Up.PoolSize == 0 {
		config.Cnx.Up.PoolSize = 32768
	}
	if config.Cnx.Up.MaxAttempts == 0 {
		config.Cnx.Up.MaxAttempts = 1
	}
	return &config
}
