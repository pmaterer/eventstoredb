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
	"testing"
)

var (
	client        *Client
	mux           *http.ServeMux
	server        *httptest.Server
	testEvent     *Event
	testEventData TestEventData
)

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
}

func teardown() {
	server.Close()
}

type TestEventData struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
	Qux bool   `json:"qux"`
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

func TestGetEvent(t *testing.T) {
	setup()
	defer teardown()

}

func TestWriteEvent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/streams/%s", testEvent.EventType), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "ES-EventType", testEvent.EventType)
		testHeader(t, r, "ES-EventId", testEvent.eventID)
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
