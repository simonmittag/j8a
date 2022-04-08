package j8a

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/ghodss/yaml"
	isd "github.com/jbenet/go-is-domain"
	"github.com/rs/zerolog"
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
	Jwt        map[string]*Jwt
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
	log.WithLevel(zerolog.FatalLevel).Msg(msg)
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

func (config Config) validateRoutes() *Config {
	//prep routes with leading slash
	for i, _ := range config.Routes {
		if strings.Index(config.Routes[i].Path, "/") != 0 {
			config.Routes[i].Path = "/" + config.Routes[i].Path
		}
		if config.Routes[i].hasJwt() {
			if _, ok := config.Jwt[config.Routes[i].Jwt]; !ok {
				config.panic(fmt.Sprintf("route [%s] jwt [%s] not found, check your configuration", config.Routes[i].Path, config.Routes[i].Jwt))
			}
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
		if len(config.Routes[i].Transform) > 0 && config.Routes[i].Transform[:1] != "/" {
			config.panic(fmt.Sprintf("config error, illegal route transform %s", route.Transform))
		}
	}
	return &config
}

func (config Config) reApplyResourceURLDefaults() *Config {
	const http = "http"
	const https = "https"
	const p80 = 80
	const p443 = 443
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

			//insert port here if not specified
			if resourceMapping.URL.Port == 0 {
				if scheme == http {
					resourceMappings[i].URL.Port = p80
				}
				if scheme == https {
					resourceMappings[i].URL.Port = p443
				}
			}
		}
	}
	return &config
}

func (config Config) validateHTTPConfig() *Config {
	if !config.isTLSOn() &&
		config.Connection.Downstream.Http.Redirecttls == true {
		config.panic("cannot redirect to TLS if not properly configured.")
	}

	return &config
}

const wildcardDomainPrefix = "*."
const dot = "."
const noreply = "noreply@"

func (config Config) validateAcmeConfig() *Config {
	acmeProvider := len(config.Connection.Downstream.Tls.Acme.Provider) > 0
	acmeDomain := len(config.Connection.Downstream.Tls.Acme.Domains) > 0 && len(config.Connection.Downstream.Tls.Acme.Domains[0]) > 0

	if acmeProvider || acmeDomain {
		if len(config.Connection.Downstream.Tls.Cert) > 0 {
			config.panic("cannot specify TLS cert with ACME configuration, it would be overridden.")
		}

		if len(config.Connection.Downstream.Tls.Key) > 0 {
			config.panic("cannot specify TLS private key with ACME configuration, it would be overridden.")
		}

		if config.Connection.Downstream.Http.Port != 80 {
			config.panic("HTTP listener must be configured and set to port 80 for ACME challenge")
		}
	}

	if acmeDomain && !acmeProvider {
		config.panic("ACME provider must be specified with ACME domain")
	}

	if acmeProvider && !acmeDomain {
		config.panic("ACME domain must be specified with ACME provider")
	}

	if acmeDomain {
		for _, domain := range config.Connection.Downstream.Tls.Acme.Domains {
			if !govalidator.IsDNSName(domain) {
				config.panic(fmt.Sprintf("ACME domain must be a valid DNS name, but was %s", domain))
			}

			if !isd.IsDomain(domain) {
				config.panic(fmt.Sprintf("ACME domain must be a valid FQDN name, but was %s", domain))
			}

			if domain[0:1] == wildcardDomainPrefix {
				config.panic(fmt.Sprintf("ACME domain validation does not support wildcard domain names, was %s", domain))
			}

			if string(domain[0]) == dot {
				config.panic(fmt.Sprintf("ACME domain validation does not support domains starting with '.', was %s", domain))
			}

			if strings.HasSuffix(domain, dot) {
				config.panic(fmt.Sprintf("ACME domain validation does not support domains ending with '.', was %s", domain))
			}
		}

		config.Connection.Downstream.Tls.Acme.Email = noreply + config.Connection.Downstream.Tls.Acme.Domains[0]
	}

	if acmeProvider {
		if _, supported := acmeProviders[config.Connection.Downstream.Tls.Acme.Provider]; !supported {
			config.panic(fmt.Sprintf("ACME provider not supported: %s", config.Connection.Downstream.Tls.Acme.Provider))
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
		config.Connection.Downstream.ReadTimeoutSeconds = 5
	}

	if config.Connection.Downstream.RoundTripTimeoutSeconds == 0 {
		config.Connection.Downstream.RoundTripTimeoutSeconds = 10
	}

	if config.Connection.Downstream.IdleTimeoutSeconds == 0 {
		config.Connection.Downstream.IdleTimeoutSeconds = 5
	}

	if config.Connection.Downstream.MaxBodyBytes == 0 {
		//set to 2MB default value
		config.Connection.Downstream.MaxBodyBytes = 2 << 20
	}

	if !config.isHTTPOn() {
		config.Connection.Downstream.Http.Redirecttls = false
	}

	return &config
}

func (config Config) isTLSOn() bool {
	return config.Connection.Downstream.Tls.Port > 0
}

func (config Config) isHTTPOn() bool {
	return config.Connection.Downstream.Http.Port > 0
}

func (config Config) setDefaultUpstreamParams() *Config {

	if config.Connection.Upstream.SocketTimeoutSeconds == 0 {
		config.Connection.Upstream.SocketTimeoutSeconds = 3
	}
	if config.Connection.Upstream.ReadTimeoutSeconds == 0 {
		config.Connection.Upstream.ReadTimeoutSeconds = 10
	}
	if config.Connection.Upstream.IdleTimeoutSeconds == 0 {
		config.Connection.Upstream.IdleTimeoutSeconds = 5
	}
	if config.Connection.Upstream.PoolSize == 0 {
		config.Connection.Upstream.PoolSize = 32768
	}
	if config.Connection.Upstream.MaxAttempts == 0 {
		config.Connection.Upstream.MaxAttempts = 1
	}
	return &config
}

func (config Config) validateJwt() *Config {
	if len(config.Jwt) > 0 {
		for name, jwt := range config.Jwt {
			//update name on resource
			jwt.Name = name
			jwt.Init()
			err := jwt.Validate()
			if err != nil {
				config.panic(err.Error())
			}
			config.Jwt[name] = jwt
		}
		log.Debug().Msgf("parsed %d jwt configurations", len(config.Jwt))
	}
	return &config
}

func (config Config) getDownstreamRoundTripTimeoutDuration() time.Duration {
	return time.Duration(time.Second * time.Duration(config.Connection.Downstream.RoundTripTimeoutSeconds))
}
