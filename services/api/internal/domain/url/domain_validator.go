package url

// DomainValidator defines the interface for domain validation
type DomainValidator interface {
	// IsAllowed checks if a domain is in the whitelist
	IsAllowed(domain Domain) bool

	// GetAllowedDomains returns the list of allowed domains
	GetAllowedDomains() []string
}
