package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// testKeys holds RSA keys for testing
type testKeys struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	kid        string
}

// generateTestKeys creates RSA keys for testing
func generateTestKeys(kid string) (*testKeys, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &testKeys{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		kid:        kid,
	}, nil
}

// createTestJWT creates a signed JWT for testing
func createTestJWT(keys *testKeys, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keys.kid
	return token.SignedString(keys.privateKey)
}

// mockOIDCServer creates a mock OIDC server for testing.
// Returns the server (caller must close) and its URL.
type mockOIDCServer struct {
	server *httptest.Server
	keys   *testKeys
}

func newMockOIDCServer(keys *testKeys) *mockOIDCServer {
	var serverURL string

	mux := http.NewServeMux()

	// Discovery endpoint
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		doc := map[string]interface{}{
			"issuer":                                serverURL,
			"jwks_uri":                              serverURL + "/jwks",
			"authorization_endpoint":                serverURL + "/authorize",
			"token_endpoint":                        serverURL + "/token",
			"userinfo_endpoint":                     serverURL + "/userinfo",
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"subject_types_supported":               []string{"public"},
			"response_types_supported":              []string{"code"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	})

	// JWKS endpoint
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		jwks := map[string]interface{}{
			"keys": []map[string]string{
				{
					"kty": "RSA",
					"kid": keys.kid,
					"use": "sig",
					"alg": "RS256",
					"n":   base64.RawURLEncoding.EncodeToString(keys.publicKey.N.Bytes()),
					"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(keys.publicKey.E)).Bytes()),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	server := httptest.NewServer(mux)
	serverURL = server.URL

	return &mockOIDCServer{
		server: server,
		keys:   keys,
	}
}

func (m *mockOIDCServer) Close() {
	m.server.Close()
}

func (m *mockOIDCServer) URL() string {
	return m.server.URL
}

func TestNewOIDCVerifier(t *testing.T) {
	keys, err := generateTestKeys("test-key-1")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	mock := newMockOIDCServer(keys)
	defer mock.Close()

	ctx := context.Background()

	t.Run("valid config", func(t *testing.T) {
		v, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
			Issuer:   mock.URL(),
			Audience: "my-api",
		})
		if err != nil {
			t.Fatalf("NewOIDCVerifier() error = %v", err)
		}
		if v == nil {
			t.Error("NewOIDCVerifier() returned nil verifier")
		}
	})

	t.Run("missing issuer", func(t *testing.T) {
		_, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
			Audience: "my-api",
		})
		if err == nil {
			t.Error("expected error for missing issuer")
		}
	})

	t.Run("missing audience", func(t *testing.T) {
		_, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
			Issuer: mock.URL(),
		})
		if err == nil {
			t.Error("expected error for missing audience")
		}
	})

	t.Run("invalid issuer URL", func(t *testing.T) {
		_, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
			Issuer:   "https://nonexistent.example.com",
			Audience: "my-api",
		})
		if err == nil {
			t.Error("expected error for invalid issuer")
		}
	})
}

func TestOIDCVerifier_VerifyToken_Valid(t *testing.T) {
	keys, err := generateTestKeys("test-key-1")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	mock := newMockOIDCServer(keys)
	defer mock.Close()

	ctx := context.Background()
	issuer := mock.URL()

	verifier, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
		Issuer:   issuer,
		Audience: "my-api",
	})
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	// Create a valid token
	now := time.Now()
	tokenString, err := createTestJWT(keys, jwt.MapClaims{
		"iss":   issuer,
		"aud":   "my-api",
		"sub":   "user-123",
		"email": "user@example.com",
		"exp":   now.Add(time.Hour).Unix(),
		"iat":   now.Unix(),
	})
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Verify the token
	claims, err := verifier.VerifyToken(ctx, tokenString)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}

	if claims.Subject != "user-123" {
		t.Errorf("expected Subject=user-123, got %s", claims.Subject)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("expected Email=user@example.com, got %s", claims.Email)
	}
	if claims.Issuer != issuer {
		t.Errorf("expected Issuer=%s, got %s", issuer, claims.Issuer)
	}
}

func TestOIDCVerifier_VerifyToken_ExpiredToken(t *testing.T) {
	keys, err := generateTestKeys("test-key-1")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	mock := newMockOIDCServer(keys)
	defer mock.Close()

	ctx := context.Background()
	issuer := mock.URL()

	verifier, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
		Issuer:   issuer,
		Audience: "my-api",
	})
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	// Create an expired token
	now := time.Now()
	tokenString, err := createTestJWT(keys, jwt.MapClaims{
		"iss": issuer,
		"aud": "my-api",
		"sub": "user-123",
		"exp": now.Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat": now.Add(-2 * time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Verify the token - should fail
	_, err = verifier.VerifyToken(ctx, tokenString)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestOIDCVerifier_VerifyToken_InvalidAudience(t *testing.T) {
	keys, err := generateTestKeys("test-key-1")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	mock := newMockOIDCServer(keys)
	defer mock.Close()

	ctx := context.Background()
	issuer := mock.URL()

	verifier, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
		Issuer:   issuer,
		Audience: "my-api",
	})
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	// Create a token with wrong audience
	now := time.Now()
	tokenString, err := createTestJWT(keys, jwt.MapClaims{
		"iss": issuer,
		"aud": "wrong-api", // Wrong audience
		"sub": "user-123",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
	})
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Verify the token - should fail
	_, err = verifier.VerifyToken(ctx, tokenString)
	if err == nil {
		t.Fatal("expected error for invalid audience")
	}
	if err != ErrInvalidAudience {
		t.Errorf("expected ErrInvalidAudience, got %v", err)
	}
}

func TestOIDCVerifier_VerifyToken_AudienceArray(t *testing.T) {
	keys, err := generateTestKeys("test-key-1")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	mock := newMockOIDCServer(keys)
	defer mock.Close()

	ctx := context.Background()
	issuer := mock.URL()

	verifier, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
		Issuer:   issuer,
		Audience: "my-api",
	})
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	// Create a token with audience as array (common pattern)
	now := time.Now()
	tokenString, err := createTestJWT(keys, jwt.MapClaims{
		"iss": issuer,
		"aud": []string{"other-service", "my-api", "another-service"},
		"sub": "user-123",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
	})
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Verify the token - should succeed (my-api is in the array)
	claims, err := verifier.VerifyToken(ctx, tokenString)
	if err != nil {
		t.Fatalf("VerifyToken() error = %v", err)
	}

	if claims.Subject != "user-123" {
		t.Errorf("expected Subject=user-123, got %s", claims.Subject)
	}
}

func TestOIDCVerifier_VerifyToken_UnknownKey(t *testing.T) {
	keys1, err := generateTestKeys("key-1")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	keys2, err := generateTestKeys("key-2")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	// Server only has key-1
	mock := newMockOIDCServer(keys1)
	defer mock.Close()

	ctx := context.Background()
	issuer := mock.URL()

	verifier, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
		Issuer:   issuer,
		Audience: "my-api",
	})
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	// Create a token signed with key-2 (not in JWKS)
	now := time.Now()
	tokenString, err := createTestJWT(keys2, jwt.MapClaims{
		"iss": issuer,
		"aud": "my-api",
		"sub": "user-123",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
	})
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Verify the token - should fail
	_, err = verifier.VerifyToken(ctx, tokenString)
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestOIDCVerifier_IsHealthy(t *testing.T) {
	keys, err := generateTestKeys("test-key-1")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	mock := newMockOIDCServer(keys)
	defer mock.Close()

	ctx := context.Background()

	verifier, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
		Issuer:   mock.URL(),
		Audience: "my-api",
	})
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	// Check health - should succeed
	err = verifier.IsHealthy(ctx)
	if err != nil {
		t.Errorf("IsHealthy() error = %v", err)
	}
}

func TestOIDCVerifier_Getters(t *testing.T) {
	keys, err := generateTestKeys("test-key-1")
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	mock := newMockOIDCServer(keys)
	defer mock.Close()

	ctx := context.Background()

	verifier, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
		Issuer:   mock.URL(),
		Audience: "my-api",
	})
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	if verifier.Issuer() != mock.URL() {
		t.Errorf("Issuer() = %s, want %s", verifier.Issuer(), mock.URL())
	}

	if verifier.Audience() != "my-api" {
		t.Errorf("Audience() = %s, want my-api", verifier.Audience())
	}
}

// BenchmarkVerifyToken benchmarks token verification
func BenchmarkVerifyToken(b *testing.B) {
	keys, err := generateTestKeys("bench-key")
	if err != nil {
		b.Fatalf("failed to generate keys: %v", err)
	}

	mock := newMockOIDCServer(keys)
	defer mock.Close()

	ctx := context.Background()
	issuer := mock.URL()

	verifier, err := NewOIDCVerifier(ctx, OIDCVerifierConfig{
		Issuer:   issuer,
		Audience: "my-api",
	})
	if err != nil {
		b.Fatalf("failed to create verifier: %v", err)
	}

	// Create a valid token
	now := time.Now()
	tokenString, _ := createTestJWT(keys, jwt.MapClaims{
		"iss":   issuer,
		"aud":   "my-api",
		"sub":   "user-123",
		"email": "user@example.com",
		"exp":   now.Add(time.Hour).Unix(),
		"iat":   now.Unix(),
	})

	// Warm up
	verifier.VerifyToken(ctx, tokenString)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		verifier.VerifyToken(ctx, tokenString)
	}
}
