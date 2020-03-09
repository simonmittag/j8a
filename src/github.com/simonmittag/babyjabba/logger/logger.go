package logger

import (
	"crypto/md5"
	"encoding/hex"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/simonmittag/babyjabba/server"
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
	server.Version = os.Getenv("VERSION")
	if len(server.Version) == 0 {
		server.Version = "unknown"
	}
	log.Debug().Str("version", server.Version).Msg("determined version")
	return server.Version
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
	log.Logger = log.With().Str("serverId", server.ID).Logger()

}