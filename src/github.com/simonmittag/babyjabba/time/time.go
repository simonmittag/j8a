package time

import (
	"os"

	"github.com/rs/zerolog/log"
)

func Init() {
	tz := os.Getenv("TZ")
	if len(tz) == 0 {
		tz = "UTC"
	}
	log.Debug().Str("timeZone", tz).Msg("timeZone determined")
}
