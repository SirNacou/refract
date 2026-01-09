package idgen

import (
	"os"
	"testing"
)

func TestNewSnowflakeGenerator_DefaultWorkerID(t *testing.T) {
	// Clear environment variable
	os.Unsetenv("WORKER_ID")

	gen, err := NewSnowflakeGenerator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if gen.WorkerID() != 0 {
		t.Errorf("Expected default worker ID 0, got %d", gen.WorkerID())
	}
}

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

			gen, err := NewSnowflakeGenerator()
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
