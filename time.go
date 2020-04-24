package jabba

import (
	"os"

	"github.com/rs/zerolog/log"
)

var TZ = "UTC"

func initTime() {
	TZ = os.Getenv("TZ")
	if len(TZ) == 0 {
		TZ = "UTC"
	}
	log.Debug().Str("timeZone", TZ).Msg("timeZone determined")
}
