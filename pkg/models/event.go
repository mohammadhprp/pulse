package models

type Event struct {
	EventTimeMs uint64 `json:"event_time_ms"`
	Service     string `json:"service"`
	Level       string `json:"level"`
	Message     string `json:"message"`
	Host        string `json:"host"`
	RequestID   string `json:"request_id"`
}
