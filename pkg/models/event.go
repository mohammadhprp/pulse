package models

type Event struct {
	EventTimeMs uint64 `json:"event_time_ms"`
	Service     string `json:"service"`
	Level       string `json:"level"`
	Message     string `json:"message"`
	Host        string `json:"host"`
	RequestID   string `json:"request_id"`
}

type QueryOptions struct {
	Service     string `json:"service"`
	Level       string `json:"level"`
	Host        string `json:"host"`
	StartTime   uint64 `json:"start_time"`
	EndTime     uint64 `json:"end_time"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
	SortOrder   string `json:"sort_order"`
	RequestID   string `json:"request_id"`
	SearchQuery string `json:"search_query"`
}
