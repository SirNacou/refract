package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/SirNacou/refract/services/api/internal/config"
)

// ==================== TEST HELPERS ====================

func setupTestRateLimiterWithMemory() *RateLimiter {
	cfg := &config.SecurityConfig{
		RateLimitPerUser: 100,
		RateLimitWindow:  time.Hour,
	}
	// Pass nil for redis to force in-memory fallback
	return NewRateLimiter(nil, cfg)
}

func createTestRequest(userID string) *http.Request {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	return req.WithContext(ctx)
}

func createTestRequestNoAuth() *http.Request {
	return httptest.NewRequest("GET", "/test", nil)
}

// ==================== IN-MEMORY RATE LIMITER TESTS ====================

func TestRateLimiter_InMemory_WithinLimit(t *testing.T) {
	rl := setupTestRateLimiterWithMemory()

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// Make 50 requests (well within 100 limit)
	for i := 0; i < 50; i++ {
		req := createTestRequest("user-123")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, rr.Code)
		}

		// Check headers
		if rr.Header().Get("X-RateLimit-Limit") != "100" {
			t.Errorf("Request %d: expected X-RateLimit-Limit=100, got %s", i+1, rr.Header().Get("X-RateLimit-Limit"))
		}

		expectedRemaining := 100 - (i + 1)
		if rr.Header().Get("X-RateLimit-Remaining") != fmt.Sprintf("%d", expectedRemaining) {
			t.Errorf("Request %d: expected X-RateLimit-Remaining=%d, got %s", i+1, expectedRemaining, rr.Header().Get("X-RateLimit-Remaining"))
		}

		if rr.Header().Get("X-RateLimit-Reset") == "" {
			t.Errorf("Request %d: expected X-RateLimit-Reset to be set", i+1)
		}
	}
}

func TestRateLimiter_InMemory_ExceedsLimit(t *testing.T) {
	rl := setupTestRateLimiterWithMemory()

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	userID := "user-456"

	// Make exactly 100 requests (at limit)
	for i := 0; i < 100; i++ {
		req := createTestRequest(userID)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, rr.Code)
		}
	}

	// 101st request should be rate limited
	req := createTestRequest(userID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Request 101: expected status 429, got %d", rr.Code)
	}

	// Check headers
	if rr.Header().Get("X-RateLimit-Limit") != "100" {
		t.Errorf("Expected X-RateLimit-Limit=100, got %s", rr.Header().Get("X-RateLimit-Limit"))
	}

	if rr.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("Expected X-RateLimit-Remaining=0, got %s", rr.Header().Get("X-RateLimit-Remaining"))
	}

	// Check JSON response
	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["error"] != "RATE_LIMIT_EXCEEDED" {
		t.Errorf("Expected error=RATE_LIMIT_EXCEEDED, got %s", response["error"])
	}

	if response["message"] != "Rate limit exceeded (100 requests per hour)" {
		t.Errorf("Expected proper message, got %s", response["message"])
	}
}

func TestRateLimiter_InMemory_IsolatedUsers(t *testing.T) {
	rl := setupTestRateLimiterWithMemory()

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// User A makes 100 requests
	for i := 0; i < 100; i++ {
		req := createTestRequest("user-a")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("User A request %d: expected status 200, got %d", i+1, rr.Code)
		}
	}

	// User A's 101st request should be rate limited
	req := createTestRequest("user-a")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("User A request 101: expected status 429, got %d", rr.Code)
	}

	// User B should still have full quota
	req = createTestRequest("user-b")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("User B request 1: expected status 200, got %d", rr.Code)
	}

	if rr.Header().Get("X-RateLimit-Remaining") != "99" {
		t.Errorf("User B: expected X-RateLimit-Remaining=99, got %s", rr.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestRateLimiter_InMemory_MissingUserID(t *testing.T) {
	rl := setupTestRateLimiterWithMemory()

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request without user ID in context
	req := createTestRequestNoAuth()
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// With empty user ID, it should still work but use empty string as key
	// The rate limiter doesn't explicitly check for empty user ID
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 (rate limiter uses empty key), got %d", rr.Code)
	}
}

func TestRateLimiter_HeadersFormat(t *testing.T) {
	rl := setupTestRateLimiterWithMemory()

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := createTestRequest("user-headers")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Verify header format
	limit := rr.Header().Get("X-RateLimit-Limit")
	remaining := rr.Header().Get("X-RateLimit-Remaining")
	reset := rr.Header().Get("X-RateLimit-Reset")

	if limit != "100" {
		t.Errorf("X-RateLimit-Limit: expected '100', got '%s'", limit)
	}

	if remaining != "99" {
		t.Errorf("X-RateLimit-Remaining: expected '99', got '%s'", remaining)
	}

	// Reset should be a Unix timestamp (numeric)
	resetInt, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		t.Errorf("X-RateLimit-Reset: expected numeric Unix timestamp, got '%s'", reset)
	}

	// Reset time should be in the future
	if resetInt <= time.Now().Unix() {
		t.Errorf("X-RateLimit-Reset: expected future timestamp, got %d (current: %d)", resetInt, time.Now().Unix())
	}
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	rl := setupTestRateLimiterWithMemory()

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	numGoroutines := 10
	requestsPerGoroutine := 10
	userID := "user-concurrent"

	var wg sync.WaitGroup
	successCount := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localSuccess := 0
			for j := 0; j < requestsPerGoroutine; j++ {
				req := createTestRequest(userID)
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, req)
				if rr.Code == http.StatusOK {
					localSuccess++
				}
			}
			successCount <- localSuccess
		}()
	}

	wg.Wait()
	close(successCount)

	totalSuccess := 0
	for count := range successCount {
		totalSuccess += count
	}

	// All 100 requests should succeed (10 goroutines * 10 requests = 100)
	if totalSuccess != 100 {
		t.Errorf("Expected 100 successful requests, got %d", totalSuccess)
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	cfg := &config.SecurityConfig{
		RateLimitPerUser: 5,
		RateLimitWindow:  2 * time.Second, // Short window for testing
	}
	rl := NewRateLimiter(nil, cfg)

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	userID := "user-reset"

	// Make 5 requests (hit limit)
	for i := 0; i < 5; i++ {
		req := createTestRequest(userID)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, rr.Code)
		}
	}

	// 6th request should be rate limited
	req := createTestRequest(userID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Request 6: expected status 429, got %d", rr.Code)
	}

	// Wait for window to expire
	time.Sleep(3 * time.Second)

	// Request should succeed after window reset
	req = createTestRequest(userID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Request after reset: expected status 200, got %d", rr.Code)
	}

	if rr.Header().Get("X-RateLimit-Remaining") != "4" {
		t.Errorf("After reset: expected X-RateLimit-Remaining=4, got %s", rr.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestRateLimiter_CheckInMemory_ThreadSafety(t *testing.T) {
	rl := setupTestRateLimiterWithMemory()

	userID := "concurrent-user"
	window := time.Hour

	var wg sync.WaitGroup
	numGoroutines := 100
	results := make(chan int, numGoroutines)

	// Concurrently call checkInMemory
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := rl.checkInMemory(userID, window)
			results <- count
		}()
	}

	wg.Wait()
	close(results)

	// Collect all counts
	counts := make(map[int]int)
	for count := range results {
		counts[count]++
	}

	// Verify all counts are unique (1-100)
	if len(counts) != numGoroutines {
		t.Errorf("Expected %d unique counts, got %d", numGoroutines, len(counts))
	}

	// Final count should be exactly numGoroutines
	finalCount := rl.checkInMemory(userID, window)
	if finalCount != numGoroutines+1 {
		t.Errorf("Expected final count %d, got %d", numGoroutines+1, finalCount)
	}
}

func TestRateLimiter_ErrorConstants(t *testing.T) {
	// Verify error constants are defined
	if ErrRateLimitExceeded == nil {
		t.Error("ErrRateLimitExceeded should be defined")
	}

	if ErrUserIDNotFound == nil {
		t.Error("ErrUserIDNotFound should be defined")
	}

	// Verify error messages
	if ErrRateLimitExceeded.Error() != "rate limit exceeded" {
		t.Errorf("Expected 'rate limit exceeded', got '%s'", ErrRateLimitExceeded.Error())
	}

	if ErrUserIDNotFound.Error() != "user ID not found in context" {
		t.Errorf("Expected 'user ID not found in context', got '%s'", ErrUserIDNotFound.Error())
	}
}

// ==================== BENCHMARKS ====================

func BenchmarkRateLimiter_InMemory(b *testing.B) {
	rl := setupTestRateLimiterWithMemory()

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := createTestRequest("bench-user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkRateLimiter_CheckInMemory(b *testing.B) {
	rl := setupTestRateLimiterWithMemory()
	userID := "bench-user"
	window := time.Hour

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.checkInMemory(userID, window)
	}
}

func BenchmarkRateLimiter_InMemory_Parallel(b *testing.B) {
	rl := setupTestRateLimiterWithMemory()

	handler := rl.RateLimitPerUser()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			req := createTestRequest(fmt.Sprintf("bench-user-%d", i%10))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			i++
		}
	})
}
