package eventstoredb

import (
	"fmt"
	"path"
	"strconv"
)

// WriteEvent writes the given Event to an event stream.
func (c *Client) WriteEvent(event *Event, stream string) (int, error) {

	req, err := c.newRequest("POST", fmt.Sprintf("/streams/%s", stream), []*Event{event})
	if err != nil {
		return -1, err
	}
	req.Header.Set("Content-Type", "application/vnd.eventstore.events+json")

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
