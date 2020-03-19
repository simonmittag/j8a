package server

import (
	"encoding/json"
)

//StatusCodeResponse defines a JSON structure for a canned HTTP response
type StatusCodeResponse struct {
	Code       int
	Message    string
	ServerID   string
	XRequestID string
	Version    string
	Welcome    string
}

func (statusCodeResponse StatusCodeResponse) AsJSON() []byte {
	statusCodeResponse.ServerID = ID
	statusCodeResponse.Version = Version
	statusCodeResponse.Welcome = "ka cheesa crispa Greedo?"
	response, _ := json.Marshal(statusCodeResponse)
	return response
}
