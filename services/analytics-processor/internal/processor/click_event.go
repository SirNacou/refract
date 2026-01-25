package processor

import "time"

type ClickEvent struct {
	EventID   string    `json:"event_id"`
	URLID     int64     `json:"url_id"`
	ShortCode string    `json:"short_code"`
	Timestamp time.Time `json:"timestamp"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	Referrer  *string   `json:"referrer,omitempty"`
}

type StreamEntry struct {
	entryID    string
	clickEvent ClickEvent
}
