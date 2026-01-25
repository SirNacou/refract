package geo

import (
	"net/netip"
	"testing"
)

func TestIsPrivateIP_IPv4(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		private bool
	}{
		{"private 10.0.0.0/8", "10.0.0.1", true},
		{"private 172.16.0.0/12", "172.16.0.1", true},
		{"private 192.168.0.0/16", "192.168.1.1", true},
		{"loopback", "127.0.0.1", true},
		{"link-local", "169.254.1.1", true},
		{"unspecified", "0.0.0.0", true},
		{"public", "8.8.8.8", false},
		{"public", "1.1.1.1", false},
		{"public", "203.0.113.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := netip.ParseAddr(tt.ip)
			if err != nil {
				t.Fatalf("failed to parse IP %q: %v", tt.ip, err)
			}
			result := isPrivateIP(addr)
			if result != tt.private {
				t.Errorf("isPrivateIP(%q) = %v, want %v", tt.ip, result, tt.private)
			}
		})
	}
}

func TestIsPrivateIP_IPv6(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		private bool
	}{
		{"loopback", "::1", true},
		{"unspecified", "::", true},
		{"public", "2001:4860:4860::8888", false},
		{"public", "2606:4700:4700::1111", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := netip.ParseAddr(tt.ip)
			if err != nil {
				t.Fatalf("failed to parse IP %q: %v", tt.ip, err)
			}
			result := isPrivateIP(addr)
			if result != tt.private {
				t.Errorf("isPrivateIP(%q) = %v, want %v", tt.ip, result, tt.private)
			}
		})
	}
}

func TestLookup_InvalidIP(t *testing.T) {
	// Mock lookup without real DB (will fail on reader.Lookup anyway)
	lookup := &GeoLookup{reader: nil}

	result := lookup.Lookup("not-an-ip")
	if result != nil {
		t.Errorf("expected nil for invalid IP, got %+v", result)
	}
}

func TestLookup_PrivateIP(t *testing.T) {
	// Mock lookup without real DB
	lookup := &GeoLookup{reader: nil}

	privateIPs := []string{
		"10.0.0.1",
		"192.168.1.1",
		"127.0.0.1",
		"169.254.1.1",
		"::1",
		"::",
	}

	for _, ip := range privateIPs {
		t.Run(ip, func(t *testing.T) {
			result := lookup.Lookup(ip)
			if result != nil {
				t.Errorf("expected nil for private IP %q, got %+v", ip, result)
			}
		})
	}
}

func TestGeoInfo_NilFields(t *testing.T) {
	// Test that GeoInfo correctly handles nil pointer fields
	info := &GeoInfo{
		CountryCode: nil,
		CountryName: nil,
		City:        nil,
		Latitude:    nil,
		Longitude:   nil,
	}

	if info.CountryCode != nil {
		t.Errorf("expected nil CountryCode")
	}
	if info.CountryName != nil {
		t.Errorf("expected nil CountryName")
	}
	if info.City != nil {
		t.Errorf("expected nil City")
	}
	if info.Latitude != nil {
		t.Errorf("expected nil Latitude")
	}
	if info.Longitude != nil {
		t.Errorf("expected nil Longitude")
	}
}

func TestGeoInfo_WithValues(t *testing.T) {
	countryCode := "US"
	countryName := "United States"
	city := "New York"
	lat := 40.7128
	lon := -74.0060

	info := &GeoInfo{
		CountryCode: &countryCode,
		CountryName: &countryName,
		City:        &city,
		Latitude:    &lat,
		Longitude:   &lon,
	}

	if info.CountryCode == nil || *info.CountryCode != "US" {
		t.Errorf("expected CountryCode US, got %v", info.CountryCode)
	}
	if info.CountryName == nil || *info.CountryName != "United States" {
		t.Errorf("expected CountryName 'United States', got %v", info.CountryName)
	}
	if info.City == nil || *info.City != "New York" {
		t.Errorf("expected City 'New York', got %v", info.City)
	}
	if info.Latitude == nil || *info.Latitude != 40.7128 {
		t.Errorf("expected Latitude 40.7128, got %v", info.Latitude)
	}
	if info.Longitude == nil || *info.Longitude != -74.0060 {
		t.Errorf("expected Longitude -74.0060, got %v", info.Longitude)
	}
}
