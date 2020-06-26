package eventstoredb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/google/uuid"
	"github.com/pmaterer/eventstoredb/atom"
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

// Info represents information returned from the /info endpoint.
type Info struct {
	ESVersion      string `json:"esVersion"`
	State          string `json:"state"`
	ProjectionMode string `json:"projectionsMode"`
}

// Event represents a generic EventStoreDB event.
type Event struct {
	EventID   string      `json:"eventId"`
	EventType string      `json:"eventType"`
	Data      interface{} `json:"data"`
	Metadata  interface{} `json:"metadata"`
}

// NewEvent constructs a new Event type.
func NewEvent(eventType string, data, metadata interface{}) *Event {
	return &Event{
		EventID:   uuid.New().String(),
		EventType: eventType,
		Data:      data,
		Metadata:  metadata,
	}
}

// GetInfo returns an Info type.
func (c *Client) GetInfo() (*Info, error) {
	info := &Info{}
	req, err := c.newRequest("GET", "/info", nil)
	if err != nil {
		return nil, err
	}
	_, err = c.do(req, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// GetStreamHead returns the head of a stream
func (c *Client) GetStreamHead(stream string) (*atom.Feed, error) {
	feed := &atom.Feed{}
	req, err := c.newRequest("GET", fmt.Sprintf("/streams/%s/head", stream), nil)
	if err != nil {
		return nil, err
	}
	_, err = c.do(req, feed)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

// GetEvent returns the eventNo event of the stream
func (c *Client) GetEvent(stream string, eventNo int) (*Event, error) {

	event := &Event{}
	req, err := c.newRequest("GET", fmt.Sprintf("/streams/%s/%d", stream, eventNo), nil)
	if err != nil {
		return nil, err
	}
	_, err = c.do(req, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// WriteEvent writes the given Event to an event stream.
func (c *Client) WriteEvent(event *Event) (int, error) {

	req, err := c.newRequest("POST", fmt.Sprintf("/streams/%s", event.EventType), event.Data)
	if err != nil {
		return -1, err
	}

	req.Header.Set("ES-EventType", event.EventType)
	req.Header.Set("ES-EventId", event.EventID)

	resp, err := c.do(req, nil)
	if err != nil {
		return -1, err
	}
	loc, err := resp.Location()
	if err != nil {
		return -1, fmt.Errorf("event created but no location header found")
	}
	_, eventNoStr := path.Split(loc.Path)
	eventNo, err := strconv.Atoi(eventNoStr)
	if err != nil {
		return -1, fmt.Errorf("event created but invalid event number received: %v", err)
	}
	return eventNo, nil

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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.BasicAuth.Enabled {
		req.SetBasicAuth(c.BasicAuth.Username, c.BasicAuth.Password)
	}
	req.Header.Set("Accept", "application/json")
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
