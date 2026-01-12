package consumer

import (
	"testing"
)

func TestDecodeClickEvents_Success(t *testing.T) {
	entries := []StreamEntry{
		{
			ID: "1234567890000-0",
			Fields: map[string]string{
				"data": `{
					"event_id": "123e4567-e89b-12d3-a456-426614174000",
					"url_id": 42,
					"short_code": "abc123",
					"timestamp": "2024-01-13T12:00:00Z",
					"user_agent": "Mozilla/5.0",
					"ip_address": "192.168.1.1",
					"device_type": "desktop",
					"cache_tier": "l1",
					"latency_ms": 5.2,
					"request_id": "req-123"
				}`,
			},
		},
		{
			ID: "1234567890000-1",
			Fields: map[string]string{
				"data": `{
					"event_id": "223e4567-e89b-12d3-a456-426614174001",
					"url_id": 43,
					"short_code": "xyz789",
					"timestamp": "2024-01-13T12:01:00Z",
					"user_agent": "Chrome/91.0",
					"ip_address": "10.0.0.1",
					"referrer": "https://example.com",
					"country_code": "US",
					"country_name": "United States",
					"city": "New York",
					"latitude": 40.7128,
					"longitude": -74.0060,
					"device_type": "mobile",
					"browser": "Chrome",
					"operating_system": "Android",
					"cache_tier": "l2",
					"latency_ms": 12.5,
					"request_id": "req-456"
				}`,
			},
		},
	}

	events, ids, err := decodeClickEvents(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}

	// Check first event
	evt1 := events[0]
	if evt1.EventID != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("expected event_id 123e4567-e89b-12d3-a456-426614174000, got %s", evt1.EventID)
	}
	if evt1.URLID != 42 {
		t.Errorf("expected url_id 42, got %d", evt1.URLID)
	}
	if evt1.ShortCode != "abc123" {
		t.Errorf("expected short_code abc123, got %s", evt1.ShortCode)
	}
	if evt1.DeviceType != "desktop" {
		t.Errorf("expected device_type desktop, got %s", evt1.DeviceType)
	}
	if evt1.CacheTier != "l1" {
		t.Errorf("expected cache_tier l1, got %s", evt1.CacheTier)
	}
	if evt1.LatencyMS != 5.2 {
		t.Errorf("expected latency_ms 5.2, got %f", evt1.LatencyMS)
	}

	// Check second event (with optional fields)
	evt2 := events[1]
	if evt2.URLID != 43 {
		t.Errorf("expected url_id 43, got %d", evt2.URLID)
	}
	if evt2.Referrer == nil || *evt2.Referrer != "https://example.com" {
		t.Errorf("expected referrer https://example.com, got %v", evt2.Referrer)
	}
	if evt2.CountryCode == nil || *evt2.CountryCode != "US" {
		t.Errorf("expected country_code US, got %v", evt2.CountryCode)
	}
	if evt2.City == nil || *evt2.City != "New York" {
		t.Errorf("expected city New York, got %v", evt2.City)
	}
	if evt2.Latitude == nil || *evt2.Latitude != 40.7128 {
		t.Errorf("expected latitude 40.7128, got %v", evt2.Latitude)
	}

	// Check IDs
	if ids[0] != "1234567890000-0" {
		t.Errorf("expected id 1234567890000-0, got %s", ids[0])
	}
	if ids[1] != "1234567890000-1" {
		t.Errorf("expected id 1234567890000-1, got %s", ids[1])
	}
}

func TestDecodeClickEvents_MissingDataField(t *testing.T) {
	entries := []StreamEntry{
		{
			ID: "1234567890000-0",
			Fields: map[string]string{
				"other": "value",
			},
		},
	}

	_, _, err := decodeClickEvents(entries)
	if err == nil {
		t.Fatal("expected error for missing data field, got nil")
	}

	expected := "missing data field"
	if !contains(err.Error(), expected) {
		t.Errorf("expected error to contain %q, got %q", expected, err.Error())
	}
}

func TestDecodeClickEvents_EmptyDataField(t *testing.T) {
	entries := []StreamEntry{
		{
			ID: "1234567890000-0",
			Fields: map[string]string{
				"data": "",
			},
		},
	}

	_, _, err := decodeClickEvents(entries)
	if err == nil {
		t.Fatal("expected error for empty data field, got nil")
	}

	expected := "missing data field"
	if !contains(err.Error(), expected) {
		t.Errorf("expected error to contain %q, got %q", expected, err.Error())
	}
}

func TestDecodeClickEvents_InvalidJSON(t *testing.T) {
	entries := []StreamEntry{
		{
			ID: "1234567890000-0",
			Fields: map[string]string{
				"data": `{"invalid json`,
			},
		},
	}

	_, _, err := decodeClickEvents(entries)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	expected := "bad click event json"
	if !contains(err.Error(), expected) {
		t.Errorf("expected error to contain %q, got %q", expected, err.Error())
	}
}

func TestDecodeClickEvents_MissingRequiredField(t *testing.T) {
	entries := []StreamEntry{
		{
			ID: "1234567890000-0",
			Fields: map[string]string{
				"data": `{
					"event_id": "123e4567-e89b-12d3-a456-426614174000",
					"url_id": 42
				}`,
			},
		},
	}

	events, ids, err := decodeClickEvents(entries)
	// JSON unmarshal will succeed but fields will be zero values
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	// Check that zero values are present for missing fields
	if events[0].ShortCode != "" {
		t.Errorf("expected empty short_code, got %s", events[0].ShortCode)
	}
	if events[0].DeviceType != "" {
		t.Errorf("expected empty device_type, got %s", events[0].DeviceType)
	}

	if len(ids) != 1 || ids[0] != "1234567890000-0" {
		t.Errorf("expected 1 id, got %v", ids)
	}
}

func TestDecodeClickEvents_EmptyEntries(t *testing.T) {
	entries := []StreamEntry{}

	events, ids, err := decodeClickEvents(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}

	if len(ids) != 0 {
		t.Errorf("expected 0 ids, got %d", len(ids))
	}
}

func TestIsTimeoutError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "timeout error",
			err:      &testError{msg: "connection timeout"},
			expected: true,
		},
		{
			name:     "nil error message",
			err:      &testError{msg: "nil response"},
			expected: true,
		},
		{
			name:     "other error",
			err:      &testError{msg: "some other error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTimeoutError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for error: %v", tt.expected, result, tt.err)
			}
		})
	}
}

// Helper types and functions

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && anyMatch(s, substr)
}

func anyMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
