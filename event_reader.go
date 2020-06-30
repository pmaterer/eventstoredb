package eventstoredb

import (
	"fmt"

	"github.com/pmaterer/eventstoredb/atom"
)

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
