package repository

import (
	"testing"

	"github.com/SirNacou/refract/services/analytics-processor/internal/consumer"
)

func TestEventToRow_ValidEvent(t *testing.T) {
	repo := &TimescaleRepository{}

	referrer := "https://example.com"
	countryCode := "US"
	countryName := "United States"
	city := "New York"
	lat := 40.7128
	lon := -74.0060
	browser := "Chrome"
	os := "Windows 10"

	evt := consumer.ClickEvent{
		EventID:         "123e4567-e89b-12d3-a456-426614174000",
		URLID:           42,
		ShortCode:       "abc123",
		Timestamp:       "2024-01-13T12:00:00Z",
		UserAgent:       "Mozilla/5.0",
		IPAddress:       "192.168.1.1",
		Referrer:        &referrer,
		CountryCode:     &countryCode,
		CountryName:     &countryName,
		City:            &city,
		Latitude:        &lat,
		Longitude:       &lon,
		DeviceType:      "desktop",
		Browser:         &browser,
		OperatingSystem: &os,
		CacheTier:       "l1",
		LatencyMS:       5.2,
		RequestID:       "req-123",
	}

	row, err := repo.eventToRow(evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(row) != 14 {
		t.Fatalf("expected 14 columns, got %d", len(row))
	}

	// Validate key fields
	if row[2] != int64(42) {
		t.Errorf("expected url_id 42, got %v", row[2])
	}
	if row[4] != "Mozilla/5.0" {
		t.Errorf("expected user_agent Mozilla/5.0, got %v", row[4])
	}
	if row[11] != "desktop" {
		t.Errorf("expected device_type desktop, got %v", row[11])
	}
}

func TestEventToRow_MinimalEvent(t *testing.T) {
	repo := &TimescaleRepository{}

	evt := consumer.ClickEvent{
		EventID:    "223e4567-e89b-12d3-a456-426614174001",
		URLID:      100,
		ShortCode:  "xyz789",
		Timestamp:  "2024-01-13T13:00:00Z",
		UserAgent:  "curl/7.68.0",
		IPAddress:  "10.0.0.1",
		DeviceType: "bot",
		CacheTier:  "l2",
		LatencyMS:  2.5,
		RequestID:  "req-456",
	}

	row, err := repo.eventToRow(evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check nullable fields are nil
	if row[3] != nil { // referrer
		t.Errorf("expected nil referrer, got %v", row[3])
	}
	if row[6] != nil { // country_code
		t.Errorf("expected nil country_code, got %v", row[6])
	}
	if row[9] != nil { // latitude
		t.Errorf("expected nil latitude, got %v", row[9])
	}
}

func TestEventToRow_InvalidTimestamp(t *testing.T) {
	repo := &TimescaleRepository{}

	evt := consumer.ClickEvent{
		EventID:    "323e4567-e89b-12d3-a456-426614174002",
		URLID:      42,
		Timestamp:  "not-a-timestamp",
		UserAgent:  "Mozilla/5.0",
		IPAddress:  "192.168.1.1",
		DeviceType: "desktop",
	}

	_, err := repo.eventToRow(evt)
	if err == nil {
		t.Fatal("expected error for invalid timestamp, got nil")
	}
	if !contains(err.Error(), "invalid timestamp") {
		t.Errorf("expected 'invalid timestamp' error, got: %v", err)
	}
}

func TestEventToRow_InvalidEventID(t *testing.T) {
	repo := &TimescaleRepository{}

	evt := consumer.ClickEvent{
		EventID:    "not-a-uuid",
		URLID:      42,
		Timestamp:  "2024-01-13T12:00:00Z",
		UserAgent:  "Mozilla/5.0",
		IPAddress:  "192.168.1.1",
		DeviceType: "desktop",
	}

	_, err := repo.eventToRow(evt)
	if err == nil {
		t.Fatal("expected error for invalid event_id, got nil")
	}
	if !contains(err.Error(), "invalid event_id") {
		t.Errorf("expected 'invalid event_id' error, got: %v", err)
	}
}

func TestEventToRow_InvalidIPAddress(t *testing.T) {
	repo := &TimescaleRepository{}

	evt := consumer.ClickEvent{
		EventID:    "423e4567-e89b-12d3-a456-426614174003",
		URLID:      42,
		Timestamp:  "2024-01-13T12:00:00Z",
		UserAgent:  "Mozilla/5.0",
		IPAddress:  "not-an-ip",
		DeviceType: "desktop",
	}

	_, err := repo.eventToRow(evt)
	if err == nil {
		t.Fatal("expected error for invalid ip_address, got nil")
	}
	if !contains(err.Error(), "invalid ip_address") {
		t.Errorf("expected 'invalid ip_address' error, got: %v", err)
	}
}

func TestEventToRow_InvalidDeviceType(t *testing.T) {
	repo := &TimescaleRepository{}

	evt := consumer.ClickEvent{
		EventID:    "523e4567-e89b-12d3-a456-426614174004",
		URLID:      42,
		Timestamp:  "2024-01-13T12:00:00Z",
		UserAgent:  "Mozilla/5.0",
		IPAddress:  "192.168.1.1",
		DeviceType: "smartwatch",
	}

	_, err := repo.eventToRow(evt)
	if err == nil {
		t.Fatal("expected error for invalid device_type, got nil")
	}
	if !contains(err.Error(), "invalid device_type") {
		t.Errorf("expected 'invalid device_type' error, got: %v", err)
	}
}

func TestEventToRow_MissingRequiredFields(t *testing.T) {
	repo := &TimescaleRepository{}

	tests := []struct {
		name        string
		evt         consumer.ClickEvent
		errContains string
	}{
		{
			name: "missing url_id",
			evt: consumer.ClickEvent{
				EventID:    "623e4567-e89b-12d3-a456-426614174005",
				Timestamp:  "2024-01-13T12:00:00Z",
				UserAgent:  "Mozilla/5.0",
				IPAddress:  "192.168.1.1",
				DeviceType: "desktop",
			},
			errContains: "url_id is required",
		},
		{
			name: "missing user_agent",
			evt: consumer.ClickEvent{
				EventID:    "723e4567-e89b-12d3-a456-426614174006",
				URLID:      42,
				Timestamp:  "2024-01-13T12:00:00Z",
				IPAddress:  "192.168.1.1",
				DeviceType: "desktop",
			},
			errContains: "user_agent is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.eventToRow(tt.evt)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !contains(err.Error(), tt.errContains) {
				t.Errorf("expected error containing %q, got: %v", tt.errContains, err)
			}
		})
	}
}

func TestEventToRow_InvalidCountryCode(t *testing.T) {
	repo := &TimescaleRepository{}

	invalidCode := "USA" // should be 2 chars
	evt := consumer.ClickEvent{
		EventID:     "823e4567-e89b-12d3-a456-426614174007",
		URLID:       42,
		Timestamp:   "2024-01-13T12:00:00Z",
		UserAgent:   "Mozilla/5.0",
		IPAddress:   "192.168.1.1",
		DeviceType:  "desktop",
		CountryCode: &invalidCode,
	}

	_, err := repo.eventToRow(evt)
	if err == nil {
		t.Fatal("expected error for invalid country_code length, got nil")
	}
	if !contains(err.Error(), "invalid country_code") {
		t.Errorf("expected 'invalid country_code' error, got: %v", err)
	}
}

func TestIsValidDeviceType(t *testing.T) {
	tests := []struct {
		deviceType string
		valid      bool
	}{
		{"desktop", true},
		{"mobile", true},
		{"tablet", true},
		{"bot", true},
		{"smartwatch", false},
		{"", false},
		{"Desktop", false}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.deviceType, func(t *testing.T) {
			result := isValidDeviceType(tt.deviceType)
			if result != tt.valid {
				t.Errorf("isValidDeviceType(%q) = %v, want %v", tt.deviceType, result, tt.valid)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
