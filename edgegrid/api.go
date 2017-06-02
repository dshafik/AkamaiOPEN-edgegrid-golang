package edgegrid

import (
	gojson "encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

// APIError exposes an Akamai OPEN Edgegrid Error
type APIError struct {
	error
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Status      int       `json:"status"`
	Detail      string    `json:"detail"`
	Instance    string    `json:"instance"`
	Method      string    `json:"method"`
	ServerIP    string    `json:"serverIp"`
	ClientIP    string    `json:"clientIp"`
	RequestID   string    `json:"requestId"`
	RequestTime string    `json:"requestTime"`
	Response    *Response `json:"-"`
	RawBody     string    `json:"-"`
}

func (error APIError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("API Error: %d %s %s", error.Status, error.Title, error.Detail))
}

// NewAPIError creates a new API error based on a Response,
// or http.Response-like.
func NewAPIError(response *Response) APIError {
	body, _ := ioutil.ReadAll(response.Body)

	return NewAPIErrorFromBody(response, body)
}

// NewAPIErrorFromBody creates a new API error, allowing you to pass in a body
//
// This function is intended to be used after the body has already been read for
// other purposes.
func NewAPIErrorFromBody(response *Response, body []byte) APIError {
	error := APIError{}

	if err := json.Unmarshal(body, &error); err != nil {
		error.Status = response.StatusCode
		error.Title = response.Status
	}

	error.Response = response
	error.RawBody = string(body)

	return error
}

type resource struct {
	Complete chan bool `json:"-"`
}

// Init initializes the Complete channel, if it is necessary
// need to create a resource specific Init(), make sure to
// initialize the channel.
func (resource *resource) Init() {
	resource.Complete = make(chan bool, 1)
}

// PostUnmarshalJSON is a default implementation of the
// PostUnmarshalJSON hook that simply calls Init() and
// sends true to the Complete channel. This is overridden
// in many resources, in particular those that represent
// collections, and have to initialize sub-resources also.
func (resource *resource) PostUnmarshalJSON() error {
	resource.Init()
	resource.Complete <- true
	return nil
}

func (resource *resource) GetJSON() ([]byte, error) {
	return gojson.MarshalIndent(resource, "", "    ")
}
