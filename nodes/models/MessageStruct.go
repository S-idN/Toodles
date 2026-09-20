package models

import "time"

type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
	Body      string    `json:"body"`
}
