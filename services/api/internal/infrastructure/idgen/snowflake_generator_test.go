package idgen

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestNewSnowflakeGenerator_FromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantID    int64
		wantError bool
	}{
		{"Valid worker ID 42", "42", 42, false},
		{"Valid worker ID 0", "0", 0, false},
		{"Valid worker ID 1023", "1023", 1023, false},
		{"Invalid negative", "-1", 0, true},
		{"Invalid too large", "1024", 0, true},
		{"Invalid non-numeric", "abc", 0, true},
		{"Empty falls back to 0", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("WORKER_ID", tt.envValue)
			} else {
				os.Unsetenv("WORKER_ID")
			}
			defer os.Unsetenv("WORKER_ID")

			workerID, err := getWorkerIDFromEnv()
			if err != nil {
				t.Errorf("Unexpected error for WORKER_ID=%s: %v", tt.envValue, err)
			}

			gen, err := NewSnowflakeGenerator(workerID)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for WORKER_ID=%s, got nil", tt.envValue)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for WORKER_ID=%s: %v", tt.envValue, err)
				}
				if gen.WorkerID() != tt.wantID {
					t.Errorf("Expected worker ID %d, got %d", tt.wantID, gen.WorkerID())
				}
			}
		})
	}
}

func TestNewSnowflakeGeneratorWithWorkerID(t *testing.T) {
	tests := []struct {
		name      string
		workerID  int64
		wantError bool
	}{
		{"Valid worker ID 10", 10, false},
		{"Valid worker ID 0", 0, false},
		{"Valid worker ID 1023", 1023, false},
		{"Invalid worker ID -1", -1, true},
		{"Invalid worker ID 1024", 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewSnowflakeGeneratorWithWorkerID(tt.workerID)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for worker ID %d, got nil", tt.workerID)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for worker ID %d: %v", tt.workerID, err)
				}
				if gen.WorkerID() != tt.workerID {
					t.Errorf("Expected worker ID %d, got %d", tt.workerID, gen.WorkerID())
				}
			}
		})
	}
}

func TestSnowflakeGenerator_NextID(t *testing.T) {
	gen, err := NewSnowflakeGeneratorWithWorkerID(5)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Generate multiple IDs
	ids := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		id, err := gen.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		// Check uniqueness
		if ids[id] {
			t.Errorf("Generated duplicate ID: %d", id)
		}
		ids[id] = true

		// Check non-zero
		if id == 0 {
			t.Error("Generated ID is zero")
		}
	}

	if len(ids) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(ids))
	}
}

func TestSnowflakeGenerator_ImplementsIDGenerator(t *testing.T) {
	// Verify SnowflakeGenerator implements IDGenerator interface
	var _ IDGenerator = (*SnowflakeGenerator)(nil)
}

func TestGetWorkerIDFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		want      int64
		wantError bool
	}{
		{"Not set returns 0", "", 0, false},
		{"Valid 0", "0", 0, false},
		{"Valid 42", "42", 42, false},
		{"Valid 1023", "1023", 1023, false},
		{"Invalid negative", "-5", 0, true},
		{"Invalid too large", "2000", 0, true},
		{"Invalid string", "not-a-number", 0, true},
		{"Invalid float", "3.14", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("WORKER_ID", tt.envValue)
			} else {
				os.Unsetenv("WORKER_ID")
			}
			defer os.Unsetenv("WORKER_ID")

			got, err := getWorkerIDFromEnv()
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for WORKER_ID=%s, got nil", tt.envValue)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for WORKER_ID=%s: %v", tt.envValue, err)
				}
				if got != tt.want {
					t.Errorf("Expected worker ID %d, got %d", tt.want, got)
				}
			}
		})
	}
}

func BenchmarkSnowflakeGenerator_NextID(b *testing.B) {
	gen, err := NewSnowflakeGeneratorWithWorkerID(1)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.NextID()
		if err != nil {
			b.Fatalf("Failed to generate ID: %v", err)
		}
	}
}

// ==================== ADDITIONAL COMPREHENSIVE TESTS ====================

func TestSnowflakeGenerator_NextID_Monotonic(t *testing.T) {
	gen, err := NewSnowflakeGeneratorWithWorkerID(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	var lastID uint64
	for i := 0; i < 100; i++ {
		id, err := gen.NextID()
		if err != nil {
			t.Fatalf("NextID() error = %v", err)
		}

		if id <= lastID {
			t.Errorf("IDs not monotonically increasing: last=%d, current=%d", lastID, id)
		}
		lastID = id
	}
}

func TestSnowflakeGenerator_ConcurrentGeneration(t *testing.T) {
	gen, err := NewSnowflakeGeneratorWithWorkerID(5)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	numGoroutines := 10
	idsPerGoroutine := 100
	totalIDs := numGoroutines * idsPerGoroutine

	idChan := make(chan uint64, totalIDs)
	var wg sync.WaitGroup

	// Generate IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := gen.NextID()
				if err != nil {
					t.Errorf("NextID() error = %v", err)
					return
				}
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// Check uniqueness
	ids := make(map[uint64]bool)
	for id := range idChan {
		if ids[id] {
			t.Errorf("Duplicate ID in concurrent generation: %d", id)
		}
		ids[id] = true
	}

	if len(ids) != totalIDs {
		t.Errorf("Expected %d unique IDs, got %d", totalIDs, len(ids))
	}
}

func TestSnowflakeGenerator_ExtractComponents(t *testing.T) {
	workerID := int64(42)
	gen, err := NewSnowflakeGeneratorWithWorkerID(workerID)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id, err := gen.NextID()
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}

	// Extract worker ID
	extractedWorkerID := ExtractWorkerID(id)
	if extractedWorkerID != workerID {
		t.Errorf("ExtractWorkerID() = %d, want %d", extractedWorkerID, workerID)
	}

	// Extract timestamp (should be recent)
	extractedTimestamp := ExtractTimestamp(id)
	if extractedTimestamp <= 0 {
		t.Errorf("ExtractTimestamp() = %d, want positive value", extractedTimestamp)
	}

	// Extract sequence
	extractedSequence := ExtractSequence(id)
	if extractedSequence < 0 || extractedSequence > maxSequence {
		t.Errorf("ExtractSequence() = %d, want 0-%d", extractedSequence, maxSequence)
	}
}

func TestSnowflakeGenerator_IDToTime(t *testing.T) {
	gen, err := NewSnowflakeGeneratorWithWorkerID(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	beforeGen := time.Now()
	id, err := gen.NextID()
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	afterGen := time.Now()

	idTime := IDToTime(id)

	// ID timestamp should be between before and after generation (with tolerance)
	if idTime.Before(beforeGen.Add(-time.Second)) || idTime.After(afterGen.Add(time.Second)) {
		t.Errorf("IDToTime() = %v, expected between %v and %v", idTime, beforeGen, afterGen)
	}
}

// ==================== BASE62 ENCODING TESTS ====================

func TestEncodeBase62_DecodeBase62_Roundtrip(t *testing.T) {
	tests := []struct {
		name string
		id   uint64
	}{
		{"Zero", 0},
		{"One", 1},
		{"Small", 42},
		{"Medium", 123456},
		{"Large", 1234567890},
		{"Very large", 1234567890123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeBase62(tt.id)
			decoded, err := DecodeBase62(encoded)
			if err != nil {
				t.Errorf("DecodeBase62() error = %v", err)
				return
			}
			if decoded != tt.id {
				t.Errorf("Roundtrip failed: original=%d, encoded=%s, decoded=%d", tt.id, encoded, decoded)
			}
		})
	}
}

func TestEncodeBase62_Format(t *testing.T) {
	tests := []struct {
		name        string
		id          uint64
		wantEncoded string
	}{
		{"Zero", 0, "0"},
		{"Single digit", 5, "5"},
		{"Letter a", 10, "a"},
		{"Letter A", 36, "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeBase62(tt.id)

			if tt.wantEncoded != "" && encoded != tt.wantEncoded {
				t.Errorf("EncodeBase62() = %v, want %v", encoded, tt.wantEncoded)
			}

			// Verify only valid Base62 characters
			for _, c := range encoded {
				if !isValidBase62Char(c) {
					t.Errorf("Invalid Base62 character in encoded string: %c", c)
				}
			}
		})
	}
}

func TestDecodeBase62_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		wantErr bool
	}{
		{"Empty string", "", true},
		{"Invalid char -", "abc-def", true},
		{"Invalid char +", "abc+def", true},
		{"Invalid char /", "abc/def", true},
		{"Invalid char space", "abc def", true},
		{"Valid lowercase", "abc", false},
		{"Valid uppercase", "ABC", false},
		{"Valid digits", "123", false},
		{"Valid mixed", "a1B2c3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeBase62(tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeBase62() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecodeBase62_OverflowProtection(t *testing.T) {
	// Create a very long Base62 string that would overflow uint64
	longString := ""
	for i := 0; i < 20; i++ {
		longString += "ZZZZZZZZZZ" // Max Base62 chars
	}

	_, err := DecodeBase62(longString)
	if err == nil {
		t.Error("DecodeBase62() should error on overflow, but didn't")
	}
}

func TestBase62_URLSafe(t *testing.T) {
	// Generate several IDs and verify encoded strings are URL-safe
	gen, _ := NewSnowflakeGeneratorWithWorkerID(1)

	for i := 0; i < 100; i++ {
		id, _ := gen.NextID()
		encoded := EncodeBase62(id)

		// Check for URL-unsafe characters
		for _, c := range encoded {
			if c == '+' || c == '/' || c == '=' {
				t.Errorf("URL-unsafe character found: %c in %s", c, encoded)
			}
		}
	}
}

// ==================== HELPER FUNCTIONS ====================

func isValidBase62Char(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// ==================== ADDITIONAL BENCHMARKS ====================

func BenchmarkSnowflakeGenerator_NextID_Parallel(b *testing.B) {
	gen, _ := NewSnowflakeGeneratorWithWorkerID(1)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gen.NextID()
		}
	})
}
