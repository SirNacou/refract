package idgen

import (
	"testing"
)

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		name     string
		id       uint64
		expected string
	}{
		{
			name:     "zero",
			id:       0,
			expected: "0",
		},
		{
			name:     "small number",
			id:       123,
			expected: "1Z",
		},
		{
			name:     "medium number",
			id:       1234567890,
			expected: "1ly7vk",
		},
		{
			name:     "large snowflake ID",
			id:       1234567890123456,
			expected: "5Ezg7yb1S",
		},
		{
			name:     "max uint64",
			id:       ^uint64(0), // 18446744073709551615
			expected: "lYGhA16ahyf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeBase62(tt.id)
			if result != tt.expected {
				t.Errorf("EncodeBase62(%d) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

func TestDecodeBase62(t *testing.T) {
	tests := []struct {
		name     string
		encoded  string
		expected uint64
		wantErr  bool
	}{
		{
			name:     "zero",
			encoded:  "0",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "small number",
			encoded:  "1Z",
			expected: 123,
			wantErr:  false,
		},
		{
			name:     "medium number",
			encoded:  "1ly7vk",
			expected: 1234567890,
			wantErr:  false,
		},
		{
			name:     "large snowflake ID",
			encoded:  "5Ezg7yb1S",
			expected: 1234567890123456,
			wantErr:  false,
		},
		{
			name:     "max uint64",
			encoded:  "lYGhA16ahyf",
			expected: ^uint64(0),
			wantErr:  false,
		},
		{
			name:     "empty string",
			encoded:  "",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid character - space",
			encoded:  "abc def",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid character - special char",
			encoded:  "abc@def",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeBase62(tt.encoded)
			if tt.wantErr {
				if err == nil {
					t.Errorf("DecodeBase62(%s) expected error, got nil", tt.encoded)
				}
				return
			}

			if err != nil {
				t.Errorf("DecodeBase62(%s) unexpected error: %v", tt.encoded, err)
				return
			}

			if result != tt.expected {
				t.Errorf("DecodeBase62(%s) = %d, want %d", tt.encoded, result, tt.expected)
			}
		})
	}
}

func TestBase62RoundTrip(t *testing.T) {
	testIDs := []uint64{
		0,
		1,
		62,
		123456789,
		1234567890123456,
		^uint64(0) - 1,
		^uint64(0),
	}

	for _, id := range testIDs {
		t.Run("", func(t *testing.T) {
			encoded := EncodeBase62(id)
			decoded, err := DecodeBase62(encoded)
			if err != nil {
				t.Errorf("DecodeBase62 failed for encoded ID %s: %v", encoded, err)
				return
			}

			if decoded != id {
				t.Errorf("Round trip failed: %d -> %s -> %d", id, encoded, decoded)
			}
		})
	}
}

func TestBase62URLSafety(t *testing.T) {
	// Generate some IDs and ensure they're URL-safe (alphanumeric only)
	gen, err := NewSnowflakeGeneratorWithWorkerID(0)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	for i := 0; i < 100; i++ {
		id, err := gen.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		encoded := EncodeBase62(id)

		// Check that all characters are in the Base62 alphabet
		for _, char := range encoded {
			found := false
			for _, validChar := range base62Alphabet {
				if char == validChar {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Encoded string %s contains non-Base62 character: %c", encoded, char)
			}
		}

		// Verify length is reasonable (8-11 characters for Snowflake IDs)
		if len(encoded) < 8 || len(encoded) > 11 {
			t.Logf("Warning: Encoded string length %d is outside expected range 8-11: %s", len(encoded), encoded)
		}
	}
}

func BenchmarkEncodeBase62(b *testing.B) {
	id := uint64(1234567890123456)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeBase62(id)
	}
}

func BenchmarkDecodeBase62(b *testing.B) {
	encoded := "dBvJIX9uyO"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodeBase62(encoded)
	}
}
