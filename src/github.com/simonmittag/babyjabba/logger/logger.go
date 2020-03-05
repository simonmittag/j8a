package logger

import (
	"crypto/md5"
	"encoding/hex"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//ServerID is a unique identifier made up as md5 of hostname and version.
var ServerID string = "unknown"
var host string
var version string

//initServerId creates a unique ID for the server log.
func initServerID() {
	host, _ = os.Hostname()
	log.Debug().Str("hostName", host).Msg("determined hostName")
	version = os.Getenv("VERSION")
	if len(version) == 0 {
		version = "unknown"
	}
	log.Debug().Str("version", version).Msg("determined version")

	data := []byte(host + version)

	hasher := md5.New()
	hasher.Write(data)
	ServerID = hex.EncodeToString(hasher.Sum(nil))[0:8]
	log.Debug().Str("serverID", ServerID).Msg("determined serverID")
}

// Init sets up a global logger instance
func Init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	w := zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: false,
	}
	log.Logger = log.Output(w)
	initServerID()
	log.Logger = log.With().Str("serverId", ServerID).Logger()

}
