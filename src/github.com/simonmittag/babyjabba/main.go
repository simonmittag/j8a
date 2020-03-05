package main

import (
	"github.com/rs/zerolog/log"
	"github.com/simonmittag/babyjabba/logger"
	"github.com/simonmittag/babyjabba/time"
)

func main() {
	logger.Init()
	time.Init()
	log.Info().Msg("BabyJabba is starting up....")

	for {

	}
}
