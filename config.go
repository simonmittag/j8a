package jabba

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

//Config is the system wide configuration for Jabba
type Config struct {
	Policies   map[string]Policy
	Routes     Routes
	Resources  map[string][]ResourceMapping
	Connection Connection
}

const HTTP = "HTTP"
const TLS = "TLS"

func (config Config) read(file string) *Config {
	jsonFile, err := os.Open(file)
	defer jsonFile.Close()
	if err != nil {
		msg := "cannot find config file or unable to read server configuration, exiting..."
		log.Fatal().Msg(msg)
		panic(msg)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return config.parse(byteValue)
}

func (config Config) parse(jsonConfig []byte) *Config {
	json.Unmarshal(jsonConfig, &config)
	//todo tell me about more of the config, number of routes
	log.Debug().Msgf("parsed server configuration with %d live routes", len(config.Routes))
	return &config
}

func (config Config) reApplyResourceNames() *Config {
	for name := range config.Resources {
		resourceMappings := config.Resources[name]
		for i, resourceMapping := range resourceMappings {
			resourceMapping.Name = name
			resourceMappings[i] = resourceMapping
		}
	}
	return &config
}

func (config Config) sortRoutes() *Config {
	//prep routes with leading slash
	for i, _ := range config.Routes {
		if strings.Index(config.Routes[i].Path, "/") != 0 {
			config.Routes[i].Path = "/" + config.Routes[i].Path
		}
	}
	sort.Sort(Routes(config.Routes))
	return &config
}

func (config Config) compileRoutePaths() *Config {
	for i, route := range config.Routes {
		config.Routes[i].Regex, _ = regexp.Compile("^"+route.Path)
	}
	return &config
}

func (config Config) reApplyResourceSchemes() *Config {
	const http = "http"
	const https = "https"
	for name := range config.Resources {
		resourceMappings := config.Resources[name]
		for i, resourceMapping := range resourceMappings {
			scheme := resourceMapping.URL.Scheme
			if len(scheme) == 0 {
				scheme = http
			} else {
				if strings.Contains(scheme, https) {
					scheme = https
				} else if strings.Contains(scheme, http) {
					scheme = http
				}
			}
			resourceMappings[i].URL.Scheme = scheme
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
	if config.Policies == nil {
		config.Policies = make(map[string]Policy)
	}
	config.Policies["default"] = *defaultPolicy
	return &config
}

func (config Config) setDefaultDownstreamParams() *Config {

	if config.Connection.Downstream.ReadTimeoutSeconds == 0 {
		config.Connection.Downstream.ReadTimeoutSeconds = 120
	}

	if config.Connection.Downstream.RoundTripTimeoutSeconds == 0 {
		config.Connection.Downstream.RoundTripTimeoutSeconds = 240
	}

	if config.Connection.Downstream.IdleTimeoutSeconds == 0 {
		config.Connection.Downstream.IdleTimeoutSeconds = 120
	}

	if len(config.Connection.Downstream.Mode) == 0 {
		config.Connection.Downstream.Mode = HTTP
	} else {
		config.Connection.Downstream.Mode = strings.ToUpper(config.Connection.Downstream.Mode)
	}

	if config.Connection.Downstream.Port == 0 {
		if config.isTLSMode() {
			config.Connection.Downstream.Port = 443
		} else {
			config.Connection.Downstream.Port = 8080
		}
	}

	return &config
}

func (config Config) isTLSMode() bool {
	return strings.ToUpper(config.Connection.Downstream.Mode) == TLS
}

func (config Config) setDefaultUpstreamParams() *Config {

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
