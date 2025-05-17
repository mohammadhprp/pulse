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
	Page        int    `json:"page"`
	PerPage     int    `json:"per_page"`
	SortOrder   string `json:"sort_order"`
	RequestID   string `json:"request_id"`
	SearchQuery string `json:"search_query"`
}

type PaginatedResponse struct {
	Data        []Event `json:"data"`
	Total       int64   `json:"total"`
	PerPage     int     `json:"per_page"`
	CurrentPage int     `json:"current_page"`
	LastPage    int     `json:"last_page"`
	From        int     `json:"from"`
	To          int     `json:"to"`
}
