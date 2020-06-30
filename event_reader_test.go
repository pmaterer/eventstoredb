package eventstoredb

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"testing"
)

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
