package j8a

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

//Config is the system wide configuration for j8a
type Config struct {
	Policies   map[string]Policy
	Routes     Routes
	Resources  map[string][]ResourceMapping
	Connection Connection
}

const HTTP = "HTTP"
const TLS = "TLS"
const J8ACFG_YML = "J8ACFG_YML"

var ConfigFile = ""

const DefaultConfigFile = "j8acfg.yml"

func (config Config) load() *Config {
	if len(ConfigFile) > 0 {
		log.Debug().Msgf("attempting config from file '%s'", ConfigFile)
		if cfg := *config.readYmlFile(ConfigFile); cfg.ok() {
			config = cfg
		} else {
			config.panic(fmt.Sprintf("error loading config from file '%s'", ConfigFile))
		}
	} else if len(os.Getenv(J8ACFG_YML)) > 0 {
		log.Debug().Msgf("attempting config from env %s", J8ACFG_YML)
		if cfg := *config.readYmlEnv(); cfg.ok() {
			config = cfg
		} else {
			config.panic(fmt.Sprintf("error loading config from env '%s'", J8ACFG_YML))
		}
	} else {
		log.Debug().Msgf("attempting config from default file '%s'", DefaultConfigFile)
		if cfg := config.readYmlFile(DefaultConfigFile); cfg.ok() {
			config = *cfg
		} else {
			config.panic(fmt.Sprintf("error loading config from default file '%s'", DefaultConfigFile))
		}
	}
	log.Debug().Msgf("parsed config with %d live routes", config.Routes.Len())
	return &config
}

func (config Config) panic(msg string) {
	if len(msg) == 0 {
		msg = "error loading config."
	}
	msg = msg + " exiting..."
	log.Fatal().Msg(msg)
	panic(msg)
}

func (config Config) readYmlEnv() *Config {
	conf := []byte(os.Getenv(J8ACFG_YML))
	return config.parse(conf)
}

func (config Config) ok() bool {
	//lightweight test if config was loaded, not necessarily valid
	return config.Routes.Len() > 0
}

func (config Config) readYmlFile(file string) *Config {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		msg := fmt.Sprintf("unable to load config from %s, exiting...", file)
		log.Fatal().Msg(msg)
		panic(msg)
	}
	byteValue, _ := ioutil.ReadAll(f)
	return config.parse(byteValue)
}

func (config Config) parse(yml []byte) *Config {
	jsn, _ := yaml.YAMLToJSON(yml)
	json.Unmarshal(jsn, &config)
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
	var err error
	for i, route := range config.Routes {
		config.Routes[i].PathRegex, err = regexp.Compile("^" + route.Path)
		if err != nil {
			config.panic(fmt.Sprintf("config error, illegal route path %s", route.Path))
		}
	}
	return &config
}

func (config Config) compileRouteTransforms() *Config {
	for i, route := range config.Routes {
		if len(config.Routes[i].Transform)>0 && config.Routes[i].Transform[:1] != "/" {
			config.panic(fmt.Sprintf("config error, illegal route transform %s", route.Transform))
		}
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

	if config.Connection.Downstream.MaxBodyBytes == 0 {
		//set to 2MB default value
		config.Connection.Downstream.MaxBodyBytes = 2 << 20
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

func (config Config) getDownstreamRoundTripTimeoutDuration() time.Duration {
	return time.Duration(time.Second * time.Duration(config.Connection.Downstream.RoundTripTimeoutSeconds))
}
