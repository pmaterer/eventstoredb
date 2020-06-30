package eventstoredb

import (
	"net/http"
	"reflect"
	"testing"
)

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
