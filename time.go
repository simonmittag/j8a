package jabba

import (
	"os"

	"github.com/rs/zerolog/log"
)

var TZ = "UTC"

func initTime() string {
	tz := os.Getenv("TZ")
	if len(tz) == 0 {
		tz = "UTC"
	}
	log.Debug().Str("timeZone", tz).Msg("timeZone determined")
	TZ = tz
	return tz
}
