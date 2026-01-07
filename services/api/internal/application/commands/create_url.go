package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
)

// CreateURLCommand represents the command to create a shortened URL
type CreateURLCommand struct {
	OriginalURL string
	Domain      string
	ExpiresAt   *time.Time // nil = activity-based expiration
}

// CreateURLResult contains the result of creating a URL
type CreateURLResult struct {
	ShortCode          string
	ShortURL           string
	ExpiresAt          time.Time
	HasFixedExpiration bool
}

// CreateURLHandler handles the CreateURL command
type CreateURLHandler struct {
	repo        url.Repository
	generator   url.ShortCodeGenerator
	validator   url.DomainValidator
	idGenerator url.IDGenerator
}

// NewCreateURLHandler creates a new CreateURLHandler
func NewCreateURLHandler(
	repo url.Repository,
	generator url.ShortCodeGenerator,
	validator url.DomainValidator,
	idGenerator url.IDGenerator,
) *CreateURLHandler {
	return &CreateURLHandler{
		repo:        repo,
		generator:   generator,
		validator:   validator,
		idGenerator: idGenerator,
	}
}

// Handle executes the CreateURL command
func (h *CreateURLHandler) Handle(ctx context.Context, cmd CreateURLCommand) (*CreateURLResult, error) {
	// Validate and create original URL value object
	originalURL, err := url.NewOriginalURL(cmd.OriginalURL)
	if err != nil {
		return nil, err
	}

	// Validate and create domain value object
	domain, err := url.NewDomain(cmd.Domain)
	if err != nil {
		return nil, err
	}

	// Check domain whitelist
	if !h.validator.IsAllowed(domain) {
		allowedDomains := h.validator.GetAllowedDomains()
		return nil, url.NewValidationError(
			"INVALID_DOMAIN",
			fmt.Sprintf("Domain '%s' is not allowed. Allowed domains: %v", domain.String(), allowedDomains),
		)
	}

	// Generate unique ID using Snowflake algorithm
	id := h.idGenerator.Generate()

	// Generate short code from ID
	shortCode, err := h.generator.Generate(id)
	if err != nil {
		return nil, url.NewInternalError("GENERATION_ERROR", "Failed to generate short code", err)
	}

	// Create URL entity based on expiration type
	var urlEntity *url.URL
	if cmd.ExpiresAt != nil {
		// Fixed expiration (user-specified)
		urlEntity, err = url.NewURLWithFixedExpiration(
			id,
			shortCode,
			originalURL,
			domain,
			*cmd.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
	} else {
		// Activity-based expiration (auto-renew)
		urlEntity = url.NewURLWithActivityExpiration(
			id,
			shortCode,
			originalURL,
			domain,
		)
	}

	// Save to repository
	if err := h.repo.Save(ctx, urlEntity); err != nil {
		return nil, err
	}

	// Construct full short URL
	shortURL := fmt.Sprintf("https://%s/%s", domain.String(), shortCode.String())

	return &CreateURLResult{
		ShortCode:          shortCode.String(),
		ShortURL:           shortURL,
		ExpiresAt:          urlEntity.ExpiresAt(),
		HasFixedExpiration: urlEntity.HasFixedExpiration(),
	}, nil
}
