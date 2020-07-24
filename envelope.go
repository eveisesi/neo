package neo

import (
	"encoding/json"
)

// Killmail Envelope is a container around a raw killmail with the ID and Hash extracted
type Envelope struct {
	ID       uint64          `json:"id"`
	Hash     string          `json:"hash"`
	Killmail json.RawMessage `json:"killmail"`
}

type Message struct {
	ID   uint64 `json:"id"`
	Hash string `json:"hash"`
}
