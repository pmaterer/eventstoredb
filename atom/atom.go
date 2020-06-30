package atom

import (
	"errors"
	"time"
)

// Feed represents an atom feed
type Feed struct {
	Title        string    `json:"title"`
	ID           string    `json:"id"`
	Updated      time.Time `json:"updated"`
	StreamID     string    `json:"streamId"`
	Author       Author    `json:"author"`
	HeadOfStream bool      `json:"headOfStream"`
	SelfURL      string    `json:"selfUrl"`
	ETag         string    `json:"eTag"`
	Links        []Link    `json:"links"`
	Entries      []Entry   `json:"entries"`
}

// Author represents the author object
type Author struct {
	Name string `json:"name"`
}

// Link represents an object in the links list
type Link struct {
	URI      string `json:"uri"`
	Relation string `json:"relation"`
}

// Entry represents an object in the entries list
type Entry struct {
	Title   string    `json:"title"`
	ID      string    `json:"id"`
	Updated time.Time `json:"updated"`
	Author  Author    `json:"author"`
	Summary string    `json:"summary"`
	Links   []Link    `json:"links"`
}

type LinkRelation string

const (
	Self      LinkRelation = "self"
	First                  = "first"
	Last                   = "last"
	Next                   = "next"
	Previous               = "previous"
	Metadata               = "metadata"
	Edit                   = "edit"
	Alternate              = "alternate"
)

func (feed *Feed) GetLink(l LinkRelation) (string, error) {
	for _, link := range feed.Links {
		if link.Relation == string(l) {
			return link.URI, nil
		}
	}
	return "", errors.New("Alternate link not found")
}
