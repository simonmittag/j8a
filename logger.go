package j8a

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const dwnReqRemoteAddr = "dwnReqRemoteAddr"
const dwnReqPort = "dwnReqPort"
const dwnReqPath = "dwnReqPath"
const dwnReqMethod = "dwnReqMethod"
const dwnReqUserAgent = "dwnReqUserAgent"
const dwnReqHttpVer = "dwnReqHttpVer"
const dwnReqTlsVer = "dwnReqTlsVer"
const dwnReqListnr = "dwnReqListnr"
const upBytesRead = "upBytesRead"
const upBytesWrite = "upBytesWrite"

const dwnElpsdMicros = "dwnElpsdMicros"
const dwnResErrMsg = "dwnResErrMsg"
const dwnResCode = "dwnResCode"
const dwnResCntntEnc = "dwnResCntntEnc"
const dwnResCntntLen = "dwnResCntntLen"
const dwnResElpsdMicros = "dwnResElpsdMicros"
const dwnBytesRead = "dwnBytesRead"
const dwnBytesWrite = "dwnBytesWrite"

const upReqURI = "upReqURI"
const upAtmtpElpsdMicros = "upAtmptElpsdMicros"
const upAtmpt = "upAtmpt"
const upLabel = "upLabel"
const upAtmptResCode = "upAtmptResCode"
const upAtmptResBodyBytes = "upAtmptResBodyBytes"
const upAtmptElpsdMicros = "upAtmptElpsdMicros"
const upAtmptAbort = "upAtmptAbort"

//ServerID is a unique identifier made up as md5 of hostname and version.
//initServerId creates a unique ID for the server log.
func initServerID() {
	hasher := md5.New()
	hasher.Write([]byte(getHost() + getVersion()))
	ID = hex.EncodeToString(hasher.Sum(nil))[0:8]
	log.Debug().Str("srvID", ID).Msg("srvID determined")
	log.Logger = log.With().Str("srvId", ID).Logger()
}

func getHost() string {
	host := os.Getenv("HOSTNAME")
	if len(host) == 0 {
		host, _ = os.Hostname()
	}
	log.Debug().Str("hostName", host).Msg("hostName determined")
	return host
}

func getVersion() string {
	osv := os.Getenv("VERSION")
	if len(osv) > 0 {
		Version = osv
	}

	log.Debug().Str("version", Version).Msg("version determined")
	return Version
}

// Init sets up a global logger instance
func initLogger() {
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
	case "TRUE", "YES", "Y":
		w := zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		}
		log.Logger = log.Output(w)
	default:
		//no color logging
	}

	initServerID()
	initTime()
	log.Debug().Msgf("setting global log level to %s", logLevel)
}
