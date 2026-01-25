package dto

type HealthResponse struct {
	Status       string              `json:"status"` // healthy, degraded, unhealthy
	Version      string              `json:"version"`
	Dependencies DependencyStatusMap `json:"dependencies"`
}

type DependencyStatusMap struct {
	Database DependencyStatus `json:"database"`
	Cache    DependencyStatus `json:"cache"`
	OIDC     DependencyStatus `json:"oidc"`
}

type DependencyStatus struct {
	Status         string  `json:"status"` // up, down, degraded
	ResponseTimeMs float64 `json:"response_time_ms"`
}
