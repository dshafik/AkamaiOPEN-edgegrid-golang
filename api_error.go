package edgegrid

import (
	"fmt"
	"strings"
)

type ApiError struct {
	error
	Type        string `json:"type"`
	Title       string `json:"title"`
	Status      int    `json:"status"`
	Detail      string `json:"detail"`
	Instance    string `json:"instance"`
	Method      string `json:"method"`
	ServerIP    string `json:"serverIp"`
	ClientIP    string `json:"clientIp"`
	requestId   string `json:"requestId"`
	requestTime string `json:"requestTime"`
}

func (error ApiError) Error() string {
	return strings.TrimSpace(fmt.Sprintf("API Error: %d %s %s", error.Status, error.Title, error.Detail))
}

func NewApiError(response *Response) ApiError {
	error := ApiError{}
	if err := response.BodyJson(&error); err != nil {
		error.Status = response.StatusCode
		error.Title = response.Status
	}

	return error
}
