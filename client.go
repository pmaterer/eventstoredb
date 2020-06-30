package eventstoredb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	eventstoredbVersion = "0.0.1"
	userAgent           = "eventstoredb/" + eventstoredbVersion
	defaultBaseURL      = "http://localhost:2113"
	defaultUsername     = "admin"
	defaultPassword     = "changeit"
)

var (
	// ErrUnauthorized represents a http 401 response.
	ErrUnauthorized = errors.New("EventStoreDB: unauthorized")
)

// A Client manages communication with the EventStoreDB HTTP API.
type Client struct {
	BaseURL    *url.URL
	UserAgent  string
	Username   string
	Password   string
	HTTPClient *http.Client
	BasicAuth  BasicAuth
}

// BasicAuth represents http basic auth settings.
type BasicAuth struct {
	Enabled  bool
	Username string
	Password string
}

// NewClient returns an EventStoreDB Client with defaults.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)
	c := &Client{
		BaseURL:    baseURL,
		UserAgent:  userAgent,
		HTTPClient: httpClient,
		BasicAuth: BasicAuth{
			Enabled:  true,
			Username: defaultUsername,
			Password: defaultPassword,
		},
	}
	return c
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if c.BasicAuth.Enabled {
		req.SetBasicAuth(c.BasicAuth.Username, c.BasicAuth.Password)
	}
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		defer resp.Body.Close()
		if v != nil {
			err = json.NewDecoder(resp.Body).Decode(v)
			return resp, err
		}
		return resp, nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		return resp, ErrUnauthorized
	} else {
		return resp, fmt.Errorf("http error from eventstoredb: %s", resp.Status)
	}
}
