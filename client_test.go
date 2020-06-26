package eventstoredb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"reflect"
	"strconv"
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
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
	Qux bool   `json:"qux"`
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

func TestGetInfo(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"esVersion":"5.0.7.0","state":"master","projectionsMode":"All"}`))
	})

	info, err := client.GetInfo()
	if err != nil {
		t.Errorf("GetInfo returned error: %v", err)
	}

	want := &Info{
		ESVersion:      "5.0.7.0",
		State:          "master",
		ProjectionMode: "All",
	}

	if !reflect.DeepEqual(info, want) {
		t.Errorf("GetInfo returned %+v, want %+v", info, want)
	}
}

func TestGetLatestEvent(t *testing.T) {
	setup()
	defer teardown()

	// Get head atom feed
	mux.HandleFunc(fmt.Sprintf("/streams/%s/head", testEvent.EventType), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write(feedHeadData)
	})

	headEvent, err := client.GetStreamHead(testEvent.EventType)
	if err != nil {
		t.Errorf("GetStreamHead returned error: %v", err)
	}
	latestEventURI := ""
	for _, link := range headEvent.Links {
		if link.Relation == "alternate" {
			latestEventURI = link.URI
		}
	}
	if latestEventURI == "" {
		t.Errorf("GetLatestEvent did not find latest event no")
	}
	latestEventNo, err := strconv.Atoi(path.Base(latestEventURI))
	if err != nil {
		t.Errorf("GetLatestEvent did not return an event number: %v", err)
	}

	// Get latest event
	mux.HandleFunc(fmt.Sprintf("/streams/%s/%d", testEvent.EventType, latestEventNo), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write(testEventJSON)
	})

	latestEvent, err := client.GetEvent(testEvent.EventType, latestEventNo)
	if err != nil {
		t.Errorf("GetLatestEvent / GetEvent returned error: %v", err)
	}

	if latestEvent.EventID != testEvent.EventID {
		t.Errorf("GetInfo returned Event ID %+v, want %+v", latestEvent, testEventData)
	}

	if latestEvent.EventType != testEvent.EventType {
		t.Errorf("GetInfo returned Event type %+v, want %+v", latestEvent, testEventData)
	}

}

func TestWriteEvent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/streams/%s", testEvent.EventType), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "ES-EventType", testEvent.EventType)
		testHeader(t, r, "ES-EventId", testEvent.EventID)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("error reading request body: %v", err)
		}
		mockData := TestEventData{}
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&mockData)
		if err != nil {
			t.Fatalf("error parsing body: %v", err)
		}
		if !reflect.DeepEqual(mockData, testEventData) {
			t.Errorf("invalid body received: %v, want %v", mockData, testEventData)
		}
		location := path.Join(r.URL.String(), "1")
		w.Header().Add("Location", location)
		w.Write(nil)
	})
	eventNo, err := client.WriteEvent(testEvent)
	if err != nil {
		t.Errorf("error creating event: %v", err)
	}
	if eventNo != 1 {
		t.Errorf("invalid event no: %v, want: %v", eventNo, 1)
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
