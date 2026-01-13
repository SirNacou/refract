// Package auth provides authentication and authorization infrastructure.
package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// OIDCVerifier verifies JWT tokens using OIDC discovery and JWKS.
// It is provider-agnostic and works with any OIDC-compliant identity provider.
// This implementation uses the well-maintained coreos/go-oidc library.
type OIDCVerifier struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	audience string
	issuer   string
}

// OIDCVerifierConfig holds configuration for the OIDC verifier.
type OIDCVerifierConfig struct {
	Issuer          string
	Audience        string
	SkipIssuerCheck bool // For testing only
	SkipExpiryCheck bool // For testing only
}

// Claims represents the JWT claims we extract from tokens.
type Claims struct {
	Subject string // sub claim (user ID)
	Email   string // email claim (optional)
	Issuer  string // iss claim
}

// Errors
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidIssuer    = errors.New("invalid issuer")
	ErrInvalidAudience  = errors.New("invalid audience")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrDiscoveryFailed  = errors.New("OIDC discovery failed")
)

// NewOIDCVerifier creates a new OIDC verifier with the given configuration.
// It performs OIDC discovery to fetch the provider's JWKS endpoint.
func NewOIDCVerifier(ctx context.Context, cfg OIDCVerifierConfig) (*OIDCVerifier, error) {
	if cfg.Issuer == "" {
		return nil, errors.New("issuer is required")
	}
	if cfg.Audience == "" {
		return nil, errors.New("audience is required")
	}

	issuer := strings.TrimSuffix(cfg.Issuer, "/")

	// Create OIDC provider (performs discovery)
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, errors.Join(ErrDiscoveryFailed, err)
	}

	// Configure the verifier
	verifierConfig := &oidc.Config{
		ClientID:             cfg.Audience, // Audience is validated against ClientID
		SkipIssuerCheck:      cfg.SkipIssuerCheck,
		SkipExpiryCheck:      cfg.SkipExpiryCheck,
		SupportedSigningAlgs: []string{oidc.RS256, oidc.RS384, oidc.RS512},
	}

	verifier := provider.Verifier(verifierConfig)

	return &OIDCVerifier{
		provider: provider,
		verifier: verifier,
		audience: cfg.Audience,
		issuer:   issuer,
	}, nil
}

// VerifyToken verifies a JWT token and returns the claims.
func (v *OIDCVerifier) VerifyToken(ctx context.Context, tokenString string) (*Claims, error) {
	// Verify the token (checks signature, expiry, issuer, audience)
	idToken, err := v.verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, mapOIDCError(err)
	}

	// Extract standard claims
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, ErrInvalidToken
	}

	return &Claims{
		Subject: idToken.Subject,
		Email:   claims.Email,
		Issuer:  idToken.Issuer,
	}, nil
}

// IsHealthy checks if the verifier can reach the OIDC provider.
// It does this by attempting to fetch the provider's JWKS.
func (v *OIDCVerifier) IsHealthy(ctx context.Context) error {
	// The go-oidc library caches the JWKS, so we just verify the provider exists
	// by checking that we can create a new provider (which does discovery)
	_, err := oidc.NewProvider(ctx, v.issuer)
	if err != nil {
		return errors.Join(ErrDiscoveryFailed, err)
	}
	return nil
}

// Issuer returns the configured issuer URL.
func (v *OIDCVerifier) Issuer() string {
	return v.issuer
}

// Audience returns the configured audience.
func (v *OIDCVerifier) Audience() string {
	return v.audience
}

// mapOIDCError maps go-oidc errors to our error types.
func mapOIDCError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for specific error patterns from go-oidc
	switch {
	case strings.Contains(errStr, "token is expired"):
		return ErrTokenExpired
	case strings.Contains(errStr, "issuer"):
		return ErrInvalidIssuer
	case strings.Contains(errStr, "audience"):
		return ErrInvalidAudience
	case strings.Contains(errStr, "signature"):
		return ErrInvalidSignature
	default:
		return errors.Join(ErrInvalidToken, err)
	}
}
