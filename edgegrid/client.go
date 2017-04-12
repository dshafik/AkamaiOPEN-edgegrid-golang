package edgegrid

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

const (
	libraryVersion = "0.1.0"
)

type Client struct {
	http.Client

	// HTTP client used to communicate with the Akamai APIs.
	//client *http.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// User agent for client
	UserAgent string

	Config *Config

	ConfigDnsV1 *ConfigDnsV1Service
}

type Response http.Response
type JsonBody map[string]interface{}

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

	baseUrl, err := url.Parse("https://" + config.Host)
	if err != nil {
		return nil, err
	}

	client.BaseURL = baseUrl

	client.ConfigDnsV1 = NewConfigDnsV1Service(client, config)
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

	req, err = http.NewRequest(method, u.String(), body)

	req.Header.Add("User-Agent", c.UserAgent)

	return req, nil
}

func (c *Client) NewJsonRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest(method, urlStr, buf)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json,*/*")

	return req, err
}

func (c *Client) Do(req *http.Request) (*Response, error) {
	req = c.Config.AddRequestHeader(req)
	response, err := c.Client.Do(req)

	if err != nil {
		return nil, err
	}

	res := Response(*response)

	return &res, err
}

func (c *Client) Get(url string) (resp *Response, err error) {
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
		return response, NewApiError(response)
	}

	res := Response(*response)

	return &res, err
}

func (c *Client) send(method string, url string, bodyType string, body io.Reader) (resp *Response, err error) {
	var req *http.Request

	req, err = c.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", bodyType)

	req = c.Config.AddRequestHeader(req)
	response, err := c.Do(req)

	if err != nil {
		return response, err
	}

	res := Response(*response)

	return &res, err
}

func (c *Client) Post(url string, bodyType string, body io.Reader) (resp *Response, err error) {
	return c.send("POST", url, bodyType, body)
}

func (c *Client) PostForm(url string, data url.Values) (resp *Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *Client) PostJson(url string, data interface{}) (resp *Response, err error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.Post(url, "application/json", bytes.NewReader(jsonBody))
}

func (c *Client) Put(url string, bodyType string, body io.Reader) (resp *Response, err error) {
	return c.send("PUT", url, bodyType, body)
}

func (c *Client) PutForm(url string, data url.Values) (resp *Response, err error) {
	return c.Put(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *Client) PutJson(url string, data interface{}) (resp *Response, err error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.Put(url, "application/json", bytes.NewReader(jsonBody))
}

func (c *Client) Head(url string) (resp *Response, err error) {
	req, _ := c.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	res := Response(*response)
	return &res, err
}

func (c *Client) Delete(url string) (resp *Response, err error) {
	req, _ := c.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	res := Response(*response)
	return &res, err
}

func (r *Response) BodyJson(data interface{}) error {
	if data == nil {
		return errors.New("You must pass in an interface{}")
	}

	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &data)

	return err
}

func (r *Response) IsInformational() bool {
	return r.StatusCode > 99 && r.StatusCode < 200
}

func (r *Response) IsSuccess() bool {
	return r.StatusCode > 199 && r.StatusCode < 300
}

func (r *Response) IsRedirection() bool {
	return r.StatusCode > 299 && r.StatusCode < 400
}

func (r *Response) IsClientError() bool {
	return r.StatusCode > 399 && r.StatusCode < 500
}

func (r *Response) IsServerError() bool {
	return r.StatusCode > 499 && r.StatusCode < 600
}

func (r *Response) IsError() bool {
	return r.StatusCode > 399 && r.StatusCode < 600
}
