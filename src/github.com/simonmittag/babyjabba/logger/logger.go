package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

// Init sets up a global logger instance
func Init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	w := zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: false,
	}
	log.Logger = log.Output(w)
	log.Logger = log.With().Str("serverId", "12345").Logger()
}
