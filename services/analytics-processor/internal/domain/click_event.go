package domain

import "time"

type ClickEvent struct {
	EventID         string    `json:"event_id"`
	URLID           int64     `json:"url_id"`
	ShortCode       string    `json:"short_code"`
	Timestamp       time.Time `json:"timestamp"`
	UserAgent       string    `json:"user_agent"`
	VisitorHash     string    `json:"visitor_hash"`
	Referrer        *string   `json:"referrer,omitempty"`
	CountryCode     *string   `json:"country_code,omitempty"`
	CountryName     *string   `json:"country_name,omitempty"`
	City            *string   `json:"city,omitempty"`
	Latitude        *float64  `json:"latitude,omitempty"`
	Longitude       *float64  `json:"longitude,omitempty"`
	DeviceType      string    `json:"device_type"`
	Browser         *string   `json:"browser,omitempty"`
	OperatingSystem *string   `json:"operating_system,omitempty"`
}
