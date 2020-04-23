package jabba

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

var huttese = []string{
	"Achuta!",
	"Goodde da lodia!",
	"H'chu apenkee!",
	"Chuba!",
	"Ka Cheesa Crispa Greedo?",
	"De wanna wanga?",
	"Peedunkee, caba dee unko!",
}

//AboutResponse exposes standard environment
type AboutResponse struct {
	Jabba    string
	ServerID string
	Version  string
}

//StatusCodeResponse defines a JSON structure for a canned HTTP response
type StatusCodeResponse struct {
	AboutResponse
	Code       int
	Message    string
	XRequestID string
}

//AsJSON renders the status code response into a JSON string as []byte
func (aboutResponse AboutResponse) AsJSON() []byte {
	aboutResponse.ServerID = ID
	aboutResponse.Version = Version
	aboutResponse.Jabba = randomHuttese()
	response, _ := json.Marshal(aboutResponse)
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

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(AboutResponse{}.AsJSON())
}
