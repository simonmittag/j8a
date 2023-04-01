package j8a

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
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

// ServerID is a unique identifier made up as md5 of hostname and version.
// initServerId creates a unique ID for the server log.
func initServerID() {
	hasher := md5.New()
	hasher.Write([]byte(getHost() + getVersion()))
	ID = hex.EncodeToString(hasher.Sum(nil))[0:8]
	log.Info().Str("srvID", ID).Msg("srvID determined")
	log.Logger = log.With().Str("srvId", ID).Logger()
}

// ServerID is a unique identifier made up as md5 of hostname and version.
// initServerId creates a unique ID for the server log.
func initPID() {
	pid := os.Getpid()
	log.Info().Int("pid", pid).Msg("pid determined")
	log.Logger = log.With().Int("pid", pid).Logger()
}

func getHost() string {
	host := os.Getenv("HOSTNAME")
	if len(host) == 0 {
		host, _ = os.Hostname()
	}
	log.Info().Str("hostName", host).Msg("hostName determined")
	return host
}

func getVersion() string {
	log.Info().Str("version", Version).Msg("version determined")
	return Version
}

// Init sets up a global logger instance
func initLogger() {
	defaultLevel := zerolog.InfoLevel
	zerolog.SetGlobalLevel(defaultLevel)

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
	initPID()
	log.Info().Msgf("setting global log level to %s", strings.ToUpper(defaultLevel.String()))
}
