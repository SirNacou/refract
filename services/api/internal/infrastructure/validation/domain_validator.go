package validation

import (
	"strings"

	"github.com/SirNacou/refract/services/api/internal/domain/url"
)

// WhitelistDomainValidator validates domains against a whitelist
type WhitelistDomainValidator struct {
	allowedDomains map[string]bool
	domainsList    []string
}

// NewWhitelistDomainValidator creates a new whitelist-based validator
func NewWhitelistDomainValidator(domains []string) *WhitelistDomainValidator {
	allowedDomains := make(map[string]bool, len(domains))
	domainsList := make([]string, len(domains))

	for i, domain := range domains {
		// Normalize to lowercase for case-insensitive comparison
		normalized := strings.ToLower(strings.TrimSpace(domain))
		allowedDomains[normalized] = true
		domainsList[i] = normalized
	}

	return &WhitelistDomainValidator{
		allowedDomains: allowedDomains,
		domainsList:    domainsList,
	}
}

// IsAllowed checks if a domain is in the whitelist
func (v *WhitelistDomainValidator) IsAllowed(domain url.Domain) bool {
	return v.allowedDomains[strings.ToLower(domain.String())]
}

// GetAllowedDomains returns a copy of the allowed domains list
func (v *WhitelistDomainValidator) GetAllowedDomains() []string {
	result := make([]string, len(v.domainsList))
	copy(result, v.domainsList)
	return result
}
