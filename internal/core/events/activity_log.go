package events

import "encoding/json"

type ActivityLogEvent struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Action    string          `json:"action"`
	Timestamp string          `json:"timestamp"`
	Metadata  json.RawMessage `json:"metadata"`
}

type ActivityLogRegisteredEvent struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Action    string          `json:"action"`
	Timestamp string          `json:"timestamp"`
	Metadata  json.RawMessage `json:"metadata"`
}


type LogActivityEvent struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Action    string          `json:"action"`
	Timestamp string          `json:"timestamp"`
	Metadata  json.RawMessage `json:"metadata"`
}
