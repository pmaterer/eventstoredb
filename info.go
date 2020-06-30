package eventstoredb

// Info represents information returned from the /info endpoint.
type Info struct {
	ESVersion      string `json:"esVersion"`
	State          string `json:"state"`
	ProjectionMode string `json:"projectionsMode"`
}

// GetInfo returns an Info type.
func (c *Client) GetInfo() (*Info, error) {
	info := &Info{}
	req, err := c.newRequest("GET", "/info", nil)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	_, err = c.do(req, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
