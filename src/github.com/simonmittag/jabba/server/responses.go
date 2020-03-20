package server

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

var huttese = []string{
	"Achuta!",
	"Bo shuda!",
	"Goodde da lodia!",
	"H'chu apenkee!",
	"Chuba!",
	"Ka Cheesa Crispa Greedo?",
	"De wanna wanga?",
	"Peedunkee, caba dee unko!",
	"Kuba, kayaba dee anko!",
}

//ServerInformationResponse exposes standard environment
type ServerInformationResponse struct {
	Jabba    string
	ServerID string
	Version  string
}

//StatusCodeResponse defines a JSON structure for a canned HTTP response
type StatusCodeResponse struct {
	ServerInformationResponse
	Code       int
	Message    string
	XRequestID string
}

//AsJSON renders the status code response into a JSON string as []byte
func (serverInformationResponse ServerInformationResponse) AsJSON() []byte {
	serverInformationResponse.ServerID = ID
	serverInformationResponse.Version = Version
	serverInformationResponse.Jabba = randomHuttese()
	response, _ := json.Marshal(serverInformationResponse)
	return response
}

//AsJSON renders the status code response into a JSON string as []byte
func (statusCodeResponse StatusCodeResponse) AsJSON() []byte {
	statusCodeResponse.ServerID = ID
	statusCodeResponse.Version = Version
	statusCodeResponse.Jabba = randomHuttese()
	response, _ := json.Marshal(statusCodeResponse)
	return response
}

func randomHuttese() string {
	rand.Seed(time.Now().Unix())
	return huttese[rand.Int()%len(huttese)]
}

func serverInformationHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(ServerInformationResponse{}.AsJSON())
}
