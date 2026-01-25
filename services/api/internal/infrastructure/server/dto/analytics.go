package dto

import "time"

// Query params
type GetAnalyticsParams struct {
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Granularity string     `json:"granularity" validate:"oneof=hour day week month"`
}

// Responses
type AnalyticsResponse struct {
	URLID                  int64                 `json:"url_id"`
	Summary                AnalyticsSummary      `json:"summary"`
	TimeSeries             []TimeSeriesDataPoint `json:"time_series"`
	GeographicDistribution []GeographicDataPoint `json:"geographic_distribution"`
	DeviceBreakdown        DeviceBreakdown       `json:"device_breakdown"`
	BrowserBreakdown       []BrowserDataPoint    `json:"browser_breakdown"`
	TopReferrers           []ReferrerDataPoint   `json:"top_referrers"`
}

type AnalyticsSummary struct {
	TotalClicks         int64   `json:"total_clicks"`
	UniqueVisitors      int64   `json:"unique_visitors"`
	AverageClicksPerDay float64 `json:"average_clicks_per_day"`
}

type TimeSeriesDataPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	Clicks         int64     `json:"clicks"`
	UniqueVisitors int64     `json:"unique_visitors"`
}

type GeographicDataPoint struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	Clicks      int64   `json:"clicks"`
	Percentage  float64 `json:"percentage"`
}

type DeviceBreakdown struct {
	Desktop int64 `json:"desktop"`
	Mobile  int64 `json:"mobile"`
	Tablet  int64 `json:"tablet"`
}

type BrowserDataPoint struct {
	Browser    string  `json:"browser"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}

type ReferrerDataPoint struct {
	Referrer   string  `json:"referrer"`
	Clicks     int64   `json:"clicks"`
	Percentage float64 `json:"percentage"`
}
