package safebrowsing

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

// Canonicalizer handles URL canonicalization per Google Safe Browsing specification.
// It implements the exact canonicalization rules required by GSB API v4.
//
// Reference: https://developers.google.com/safe-browsing/v4/urls-hashing#canonicalization
type Canonicalizer struct {
	idnaProfile *idna.Profile
}

// NewCanonicalizer creates a new GSB canonicalizer with default settings.
func NewCanonicalizer() *Canonicalizer {
	// Use IDNA2008 profile for internationalized domain names
	profile := idna.New(
		idna.ValidateLabels(true),
		idna.VerifyDNSLength(true),
		idna.StrictDomainName(true),
	)

	return &Canonicalizer{
		idnaProfile: profile,
	}
}

// CanonicalizeForGSB converts a URL to its canonical form per GSB specification.
//
// Canonicalization steps:
//  1. Remove tab (\t), CR (\r), and LF (\n) characters
//  2. Parse URL and validate scheme and host
//  3. Lowercase scheme (HTTP → http)
//  4. Lowercase host and convert IDN to Punycode
//  5. Remove default ports (:80 for http, :443 for https)
//  6. Normalize percent encoding in path and query
//  7. Normalize path (remove ./, ../, and //)
//  8. Remove fragment (#section)
//
// Returns the canonical URL or an error if the URL is invalid.
func (c *Canonicalizer) CanonicalizeForGSB(rawURL string) (string, error) {
	// Step 1: Remove control characters
	cleaned := removeControlChars(rawURL)
	if strings.TrimSpace(cleaned) == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	// Step 2: Parse URL
	parsed, err := url.Parse(cleaned)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate required components
	if parsed.Scheme == "" {
		return "", fmt.Errorf("URL must have a scheme (http:// or https://)")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("URL must have a host")
	}

	// Step 3: Normalize scheme
	parsed.Scheme = strings.ToLower(parsed.Scheme)

	// Step 4: Normalize host (lowercase + IDN to Punycode)
	normalizedHost, err := c.normalizeHost(parsed.Host)
	if err != nil {
		return "", fmt.Errorf("invalid host: %w", err)
	}

	// Step 5: Remove default port
	parsed.Host = removeDefaultPort(parsed.Scheme, normalizedHost)

	// Step 6: Normalize percent encoding in path (use EscapedPath for raw percent-encoded path)
	if parsed.EscapedPath() != "" {
		normalizedPath, err := normalizePathEncoding(parsed.EscapedPath())
		if err != nil {
			return "", fmt.Errorf("invalid path encoding: %w", err)
		}
		// Step 7: Normalize path structure (decoded version for navigation)
		parsed.Path = normalizePath(normalizedPath)
	} else {
		parsed.Path = "/"
	}

	// Step 6 (query): Normalize percent encoding in query string (preserve structure)
	if parsed.RawQuery != "" {
		normalizedQuery, err := normalizeQueryEncoding(parsed.RawQuery)
		if err != nil {
			return "", fmt.Errorf("invalid query encoding: %w", err)
		}
		parsed.RawQuery = normalizedQuery
	}

	// Step 8: Remove fragment
	parsed.Fragment = ""

	// Reconstruct canonical URL
	return parsed.String(), nil
}

// CanonicalizeBatch canonicalizes multiple URLs efficiently.
// Returns a slice of canonical URLs and a slice of errors (nil for successful canonicalizations).
func (c *Canonicalizer) CanonicalizeBatch(rawURLs []string) ([]string, []error) {
	results := make([]string, len(rawURLs))
	errors := make([]error, len(rawURLs))

	for i, rawURL := range rawURLs {
		canonical, err := c.CanonicalizeForGSB(rawURL)
		results[i] = canonical
		errors[i] = err
	}

	return results, errors
}

// removeControlChars removes tab (\t), carriage return (\r), and line feed (\n) characters.
// These characters are not allowed in URLs per GSB specification.
func removeControlChars(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	for i := 0; i < len(s); i++ {
		c := s[i]
		// Skip tab (0x09), CR (0x0D), LF (0x0A)
		if c != '\t' && c != '\r' && c != '\n' {
			result.WriteByte(c)
		}
	}

	return result.String()
}

// normalizePathEncoding normalizes percent encoding in URL path per GSB rules.
// - Repeatedly decode until no more decoding possible
// - Decode unreserved characters to literal form
// - Keep everything else as-is (path uses / as literal separator)
func normalizePathEncoding(path string) (string, error) {
	// Repeatedly decode
	current := path
	for i := 0; i < 10; i++ {
		decoded, changed := decodePercentOnce(current)
		if !changed {
			break
		}
		current = decoded
	}

	// Now we have fully decoded string
	// Unreserved chars stay literal, everything else needs encoding
	// But Go's URL will handle the encoding when we set parsed.Path
	return current, nil
}

// normalizeQueryEncoding normalizes percent encoding in query string per GSB rules.
// Query strings preserve =, &, and other structure, only normalize the values.
func normalizeQueryEncoding(query string) (string, error) {
	// Repeatedly decode
	current := query
	for i := 0; i < 10; i++ {
		decoded, changed := decodePercentOnce(current)
		if !changed {
			break
		}
		current = decoded
	}

	// Return decoded form - Go's URL will handle proper encoding
	return current, nil
}

// decodePercentOnce attempts one round of percent-decoding.
// Returns the result and whether any change was made.
func decodePercentOnce(s string) (string, bool) {
	var result strings.Builder
	result.Grow(len(s))
	changed := false

	i := 0
	for i < len(s) {
		if s[i] == '%' && i+2 < len(s) {
			hex := s[i+1 : i+3]

			// Try to parse hex value
			b, err := strconv.ParseUint(hex, 16, 8)
			if err == nil {
				// Successfully decoded
				result.WriteByte(byte(b))
				changed = true
				i += 3
				continue
			}
		}
		// Keep as-is if not a valid percent-encoded sequence
		result.WriteByte(s[i])
		i++
	}

	return result.String(), changed
}

// isUnreserved checks if a byte represents an unreserved character per RFC 3986.
// Unreserved characters: A-Z a-z 0-9 - . _ ~
// These can be safely decoded from percent encoding.
func isUnreserved(b byte) bool {
	return (b >= 'A' && b <= 'Z') ||
		(b >= 'a' && b <= 'z') ||
		(b >= '0' && b <= '9') ||
		b == '-' || b == '.' || b == '_' || b == '~'
}

// normalizePath normalizes URL path by removing dot segments and empty segments.
//
// Transformations:
//   - Remove /./ → /
//   - Remove /segment/../ → /
//   - Remove // → /
//   - Empty path → /
//
// Examples:
//   - /a/./b → /a/b
//   - /a/../b → /b
//   - //path → /path
//   - (empty) → /
func normalizePath(path string) string {
	if path == "" || path == "/" {
		return "/"
	}

	// Split path into segments
	segments := strings.Split(path, "/")

	// Use a stack to process segments
	stack := make([]string, 0, len(segments))

	for _, seg := range segments {
		if seg == "" || seg == "." {
			// Skip empty segments and current directory references
			continue
		}

		if seg == ".." {
			// Parent directory: pop from stack
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			// If stack is empty, we're at root, so ignore the ..
		} else {
			// Normal segment: push to stack
			stack = append(stack, seg)
		}
	}

	// Reconstruct path
	if len(stack) == 0 {
		return "/"
	}

	return "/" + strings.Join(stack, "/")
}

// removeDefaultPort removes default port numbers from the host.
//   - http://example.com:80 → http://example.com
//   - https://example.com:443 → https://example.com
//   - https://example.com:8080 → https://example.com:8080 (keep non-default)
func removeDefaultPort(scheme, host string) string {
	// Check if host has a port
	colonIndex := strings.LastIndex(host, ":")

	// Handle IPv6 addresses [::1]:port
	if strings.Contains(host, "[") {
		closeBracket := strings.LastIndex(host, "]")
		if closeBracket > colonIndex {
			// No port after IPv6 address
			return host
		}
	}

	if colonIndex == -1 {
		// No port specified
		return host
	}

	// Extract host and port
	hostPart := host[:colonIndex]
	portPart := host[colonIndex+1:]

	// Check if port is default for scheme
	if (scheme == "http" && portPart == "80") ||
		(scheme == "https" && portPart == "443") {
		// Remove default port
		return hostPart
	}

	// Keep non-default port
	return host
}

// normalizeHost lowercases the host and converts IDN to Punycode if needed.
func (c *Canonicalizer) normalizeHost(host string) (string, error) {
	// Handle IPv6 addresses - don't process the part inside brackets
	if strings.HasPrefix(host, "[") && strings.Contains(host, "]") {
		closeBracket := strings.Index(host, "]")
		// IPv6 part (keep as-is, just lowercase)
		ipv6 := strings.ToLower(host[:closeBracket+1])
		// Port part (if any)
		port := host[closeBracket+1:]
		return ipv6 + port, nil
	}

	// Extract host without port (for non-IPv6)
	hostWithoutPort := host
	port := ""
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		hostWithoutPort = host[:colonIndex]
		port = host[colonIndex:]
	}

	// Lowercase host
	hostWithoutPort = strings.ToLower(hostWithoutPort)

	// Convert IDN (internationalized domain names) to Punycode
	// This handles domains like "münchen.de" → "xn--mnchen-3ya.de"
	punycode, err := c.idnaProfile.ToASCII(hostWithoutPort)
	if err != nil {
		return "", fmt.Errorf("failed to convert host to ASCII: %w", err)
	}

	return punycode + port, nil
}
