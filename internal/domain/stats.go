package domain

import "github.com/google/uuid"

// RequestTypeStat is how many fichas were registered for a given request
// type within an analysis window.
type RequestTypeStat struct {
	RequestType string `json:"requestType"`
	Count       int    `json:"count"`
}

// ACSStat is how many fichas were registered against a given ACS within an
// analysis window.
type ACSStat struct {
	ACSID   uuid.UUID `json:"acsId"`
	ACSName string    `json:"acsName"`
	Count   int       `json:"count"`
}
