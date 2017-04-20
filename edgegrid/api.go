package edgegrid

import (
	"fmt"
	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
	"io/ioutil"
	"strings"
)

type ApiError struct {
	error
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Status      int       `json:"status"`
	Detail      string    `json:"detail"`
	Instance    string    `json:"instance"`
	Method      string    `json:"method"`
	ServerIP    string    `json:"serverIp"`
	ClientIP    string    `json:"clientIp"`
	RequestId   string    `json:"requestId"`
	RequestTime string    `json:"requestTime"`
	Response    *Response `json:"-"`
	RawBody     string    `json:"-"`
}

func (error ApiError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("API Error: %d %s %s", error.Status, error.Title, error.Detail))
}

func NewApiError(response *Response) ApiError {
	error := ApiError{}

	body, _ := ioutil.ReadAll(response.Body)

	if err := json.Unmarshal(body, &error); err != nil {
		error.Status = response.StatusCode
		error.Title = response.Status
	}

	error.Response = response
	error.RawBody = string(body)

	return error
}

type Resource struct {
	Complete chan bool
}

func (resource *Resource) Init() {
	resource.Complete = make(chan bool, 1)
}

func (resource *Resource) PostUnmarshalJSON() error {
	resource.Init()
	resource.Complete <- true
	return nil
}
