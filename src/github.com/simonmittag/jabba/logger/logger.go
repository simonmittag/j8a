package logger

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/simonmittag/jabba/server"
)

//ServerID is a unique identifier made up as md5 of hostname and version.

//initServerId creates a unique ID for the server log.
func initServerID() {
	hasher := md5.New()
	hasher.Write([]byte(getHost() + getVersion()))
	server.ID = hex.EncodeToString(hasher.Sum(nil))[0:8]
	log.Debug().Str("serverID", server.ID).Msg("determined serverID")
}

func getHost() string {
	host, _ := os.Hostname()
	log.Debug().Str("hostName", host).Msg("determined hostName")
	return host
}

func getVersion() string {
	osv := os.Getenv("VERSION")
	if len(osv) > 0 {
		server.Version = osv
	}

	log.Debug().Str("version", server.Version).Msg("determined version")
	return server.Version
}

// Init sets up a global logger instance
func Init() {
	logLevel := strings.ToUpper(os.Getenv("LOGLEVEL"))
	switch logLevel {
	case "TRACE":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logColor := strings.ToUpper(os.Getenv("LOGCOLOR"))
	switch logColor {
	case "TRUE", "YES", "y":
		w := zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		}
		log.Logger = log.Output(w)
	default:
		//no color logging
	}

	initServerID()
	log.Logger = log.With().Str("serverId", server.ID).Logger()
	log.Debug().Msgf("setting global log level to %s", logLevel)

}
