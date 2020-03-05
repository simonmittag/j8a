package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/simonmittag/babyjabba/logger"
)

func main() {
	logger.Init()
	log.Debug().Str("TimeZone", os.Getenv("TZ")).Msg("BabyJabba is starting up....")

	for {

	}
}
