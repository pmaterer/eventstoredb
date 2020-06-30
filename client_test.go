package eventstoredb

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	client        *Client
	mux           *http.ServeMux
	server        *httptest.Server
	testEvent     *Event
	testEventData TestEventData
	testEventJSON []byte
)

type TestEventData struct {
	Foo string  `json:"foo"`
	Bar float64 `json:"bar"`
	Qux bool    `json:"qux"`
}

var feedHeadData = []byte(`{
	"title": "32@test-event",
	"id": "http://localhost:2113/streams/test-event/32",
	"updated": "2020-06-25T21:10:31.662912Z",
	"author": {
	  "name": "EventStore"
	},
	"summary": "test-event",
	"content": {
	  "eventStreamId": "test-event",
	  "eventNumber": 32,
	  "eventType": "test-event",
	  "eventId": "97a77ce0-ea0a-4697-a096-fa418adcf68a",
	  "data": {
		  "foo": "Xyzzy",
		  "bar": 909,
		  "qux": false
	  },
	  "metadata": ""
	},
	"links": [
	  {
		"uri": "http://localhost:2113/streams/test-event/32",
		"relation": "edit"
	  },
	  {
		"uri": "http://localhost:2113/streams/test-event/32",
		"relation": "alternate"
	  }
	]
  }`)

func setup() {
	client = NewClient(nil)
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url

	testEventData.Foo = "Xyzzy"
	testEventData.Bar = 909
	testEventData.Qux = false
	testEvent = NewEvent("test-event", testEventData, nil)

	testEventJSON, _ = json.Marshal(testEvent)
}

func teardown() {
	server.Close()
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil)
	if c.BaseURL.String() != defaultBaseURL {
		t.Errorf("NewClient BaseURL = %v, want %v", c.BaseURL.String(), defaultBaseURL)
	}
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent = %v, want %v", c.UserAgent, userAgent)
	}
	if c.BasicAuth.Username != defaultUsername {
		t.Errorf("NewClient BasicAuth username = %v, want %v", c.BasicAuth.Username, defaultUsername)
	}
	if c.BasicAuth.Password != defaultPassword {
		t.Errorf("NewClient BasicAuth password = %v, want %v", c.BasicAuth.Password, defaultPassword)
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header, want string) {
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}
