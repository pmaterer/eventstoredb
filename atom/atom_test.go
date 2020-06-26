package atom

import "testing"

var testFeed = &Feed{
	Title: "test-feed",
	Links: []Link{
		{
			URI:      "http://localhost:2113/streams/test-event/2",
			Relation: "edit",
		},
		{
			URI:      "http://localhost:2113/streams/test-event/2",
			Relation: "alternate",
		},
	},
}

func TestGetAlternateLink(t *testing.T) {
	alternateLink, err := testFeed.GetAlternateLink()
	if err != nil {
		t.Errorf("atom.GetAlternateLink returned an error: %s", err)
	}
	want := "http://localhost:2113/streams/test-event/2"
	if alternateLink != want {
		t.Errorf("atom.GetAlternateLink returned the wrong link, got: %s, want: %s", alternateLink, want)
	}
}
