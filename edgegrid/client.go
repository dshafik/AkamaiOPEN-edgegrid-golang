package edgegrid

import (
	"bytes"
	gojson "encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/akamai-open/AkamaiOPEN-edgegrid-golang/edgegrid/json"
)

const (
	libraryVersion = "0.1.0"
)

// Client is a simple wrapper around http.Client that transparently
// signs requests made to the Akamai OPEN Edgegrid APIs.
type Client struct {
	http.Client

	// HTTP client used to communicate with the Akamai APIs.
	//client *http.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// User agent for client
	UserAgent string

	Config *Config

	ConfigDNSV1 *ConfigDNSV1Service
	PapiV0      *PapiV0Service
}

// JSONBody represents an anonymous JSON Response
type JSONBody map[string]interface{}

// NewClient creates a new Client wrapping a given http.Client and using
// a specified Config.
//
// Passing nil for the httpClient will result in using http.DefaultClient.
func NewClient(httpClient *http.Client, config *Config) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	client := &Client{
		Client: *httpClient,
		UserAgent: "Akamai-Open-Edgegrid-golang/" + libraryVersion +
			" golang/" + strings.TrimPrefix(runtime.Version(), "go"),
		Config: config,
	}

	baseURL, err := url.Parse("https://" + config.Host)
	if err != nil {
		return nil, err
	}

	client.BaseURL = baseURL

	client.ConfigDNSV1 = NewConfigDNSV1Service(client, config)
	client.PapiV0 = NewPapiV0Service(client, config)
	return client, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr, which will be resolved to the
// BaseURL of the Client. If specified, the value pointed to by body is JSON encoded and included in as the request body.
func (c *Client) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	var req *http.Request

	urlStr = strings.TrimPrefix(urlStr, "/")

	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	if req, err = http.NewRequest(method, u.String(), body); err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", c.UserAgent)

	return req, nil
}

// NewJSONRequest creates a new Request with a given body encoded using JSON
func (c *Client) NewJSONRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	buf := new(bytes.Buffer)
	// Todo: Decide if we need to wrap this for pre/post
	err := gojson.NewEncoder(buf).Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json,*/*")

	return req, nil
}

// Do performs a given HTTP Request, signed with the Akamai OPEN Edgegrid
// Authorization header. An edgegrid.Response or an error is returned.
func (c *Client) Do(req *http.Request) (*Response, error) {
	req = c.Config.AddRequestHeader(req)
	response, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	res := Response{response}
	return &res, nil
}

// Get performs a GET request to a given URL, signed  with the Akamai OPEN Edgegrid
// Authorization header. An edgegrid.Response or an error is returned.
func (c *Client) Get(url string) (*Response, error) {
	req, err := c.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req = c.Config.AddRequestHeader(req)
	response, err := c.Do(req)

	if err != nil {
		return nil, err
	}

	if response.IsError() {
		return response, NewAPIError(response)
	}

	return response, nil
}

func (c *Client) send(method string, url string, bodyType string, body io.Reader) (*Response, error) {
	req, err := c.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", bodyType)

	req = c.Config.AddRequestHeader(req)
	response, err := c.Do(req)

	if err != nil {
		return nil, err
	}

	return response, nil
}

// Post performs a generic POST request, signed with the Akamai OPEN Edgegrid
// Authorization header. An edgegrid.Response or an error is returned.
func (c *Client) Post(url string, bodyType string, body io.Reader) (*Response, error) {
	return c.send("POST", url, bodyType, body)
}

// PostForm performs a POST request with it's body encoded as HTML form data,
// signed with the Akamai OPEN Edgegrid Authorization header. An edgegrid.Response
// or an error is returned.
func (c *Client) PostForm(url string, data url.Values) (resp *Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

// PostJSON performs a POST request with it's body encoded as JSON, signed with the
// Akamai OPEN Edgegrid Authorization header. An edgegrid.Response or an error is returned.
func (c *Client) PostJSON(url string, data interface{}) (resp *Response, err error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.Post(url, "application/json", bytes.NewReader(jsonBody))
}

// Put performs a generic PUT request, signed with the Akamai OPEN Edgegrid
// Authorization header. An edgegrid.Response or an error is returned.
func (c *Client) Put(url string, bodyType string, body io.Reader) (resp *Response, err error) {
	return c.send("PUT", url, bodyType, body)
}

// PutForm performs a PUT request with it's body encoded as HTML form data,
// signed with the Akamai OPEN Edgegrid Authorization header. An edgegrid.Response
// or an error is returned.
func (c *Client) PutForm(url string, data url.Values) (resp *Response, err error) {
	return c.Put(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

// PutJSON performs a PUT request with it's body encoded as JSON, signed with the
// Akamai OPEN Edgegrid Authorization header. An edgegrid.Response or an error is returned.
func (c *Client) PutJSON(url string, data interface{}) (resp *Response, err error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.Put(url, "application/json", bytes.NewReader(jsonBody))
}

// Head performs a HEAD request, signed with the Akamai OPEN Edgegrid Authorization header.
// An edgegrid.Response or an error is returned.
func (c *Client) Head(url string) (*Response, error) {
	req, err := c.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Delete performs a DELETE request, signed with the Akamai OPEN Edgegrid Authorization header.
// An edgegrid.Response or an error is returned.
func (c *Client) Delete(url string) (resp *Response, err error) {
	req, _ := c.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return response, err
}

// Response aliases http.Response for the addition of custom behavior
type Response struct {
	*http.Response
}

// BodyJSON unmarshals the Response.Body into a given data structure
func (r *Response) BodyJSON(data interface{}) error {
	if data == nil {
		return errors.New("You must pass in an interface{}")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, data)

	return err
}

// IsInformational determines if a response was informational (1XX status)
func (r *Response) IsInformational() bool {
	return r.StatusCode > 99 && r.StatusCode < 200
}

// IsSuccess determines if a response was successful (2XX status)
func (r *Response) IsSuccess() bool {
	return r.StatusCode > 199 && r.StatusCode < 300
}

// IsRedirection determines if a response was a redirect (3XX status)
func (r *Response) IsRedirection() bool {
	return r.StatusCode > 299 && r.StatusCode < 400
}

// IsClientError determines if a response was a client error (4XX status)
func (r *Response) IsClientError() bool {
	return r.StatusCode > 399 && r.StatusCode < 500
}

// IsServerError determines if a response was a server error (5XX status)
func (r *Response) IsServerError() bool {
	return r.StatusCode > 499 && r.StatusCode < 600
}

// IsError determines if the response was a client or server error (4XX or 5XX status)
func (r *Response) IsError() bool {
	return r.StatusCode > 399 && r.StatusCode < 600
}
