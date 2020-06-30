package eventstoredb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"testing"
)

func TestWriteEvent(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/streams/%s", testEvent.EventType), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", "application/vnd.eventstore.events+json")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("error reading request body: %+v", err)
		}
		mockEvent := []*Event{}
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&mockEvent)
		if err != nil {
			t.Fatalf("error parsing body: %+v", err)
		}
		if mockEvent[0].EventID != testEvent.EventID {
			t.Errorf("WriteEvent got wrong event ID: %s, want: %s", mockEvent[0].EventID, testEvent.EventID)
		}
		if mockEvent[0].EventType != testEvent.EventType {
			t.Errorf("WriteEvent got wrong event type: %s, want: %s", mockEvent[0].EventType, testEvent.EventType)
		}
		location := path.Join(r.URL.String(), "1")
		w.Header().Add("Location", location)
		w.Write(nil)
	})

	eventNo, err := client.WriteEvent(testEvent, "test-event")
	if err != nil {
		t.Errorf("WriteEvent error creating event: %v", err)
	}
	if eventNo != 1 {
		t.Errorf("WriteEvent invalid event number got: %v, want: %v", eventNo, 1)
	}

}
