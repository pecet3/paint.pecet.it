package wsmanager

import (
	"encoding/json"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Client  *Client
}

type EventHandler = func(evt *Event)
