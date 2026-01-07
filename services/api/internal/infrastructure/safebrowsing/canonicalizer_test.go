package safebrowsing

import (
	"strings"
	"testing"
)

// TestCanonicalizeBasic tests basic URL canonicalization operations
func TestCanonicalizeBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase scheme",
			input:    "HTTP://example.com/path",
			expected: "http://example.com/path",
		},
		{
			name:     "lowercase host",
			input:    "https://EXAMPLE.COM/path",
			expected: "https://example.com/path",
		},
		{
			name:     "mixed case scheme and host",
			input:    "HTTPS://Example.COM/Path",
			expected: "https://example.com/Path",
		},
		{
			name:     "remove http default port",
			input:    "http://example.com:80/path",
			expected: "http://example.com/path",
		},
		{
			name:     "remove https default port",
			input:    "https://example.com:443/path",
			expected: "https://example.com/path",
		},
		{
			name:     "keep non-default http port",
			input:    "http://example.com:8080/path",
			expected: "http://example.com:8080/path",
		},
		{
			name:     "keep non-default https port",
			input:    "https://example.com:8443/path",
			expected: "https://example.com:8443/path",
		},
		{
			name:     "remove fragment",
			input:    "https://example.com/path#section",
			expected: "https://example.com/path",
		},
		{
			name:     "remove fragment with query",
			input:    "https://example.com/path?query=value#section",
			expected: "https://example.com/path?query=value",
		},
		{
			name:     "preserve query parameters",
			input:    "https://example.com/path?a=1&b=2",
			expected: "https://example.com/path?a=1&b=2",
		},
		{
			name:     "preserve query parameter order (no sorting)",
			input:    "https://example.com/path?z=3&a=1&m=2",
			expected: "https://example.com/path?z=3&a=1&m=2",
		},
		{
			name:     "preserve tracking parameters",
			input:    "https://example.com/path?utm_source=twitter&id=123",
			expected: "https://example.com/path?utm_source=twitter&id=123",
		},
		{
			name:     "empty path becomes root",
			input:    "https://example.com",
			expected: "https://example.com/",
		},
		{
			name:     "root path unchanged",
			input:    "https://example.com/",
			expected: "https://example.com/",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CanonicalizeForGSB(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("CanonicalizeForGSB(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCanonicalizeControlChars tests removal of control characters
func TestCanonicalizeControlChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove tab in path",
			input:    "https://example.com/pa\tth",
			expected: "https://example.com/path",
		},
		{
			name:     "remove CR in path",
			input:    "https://example.com/pa\rth",
			expected: "https://example.com/path",
		},
		{
			name:     "remove LF in path",
			input:    "https://example.com/pa\nth",
			expected: "https://example.com/path",
		},
		{
			name:     "remove multiple control chars",
			input:    "https://example.com/\t\r\npath",
			expected: "https://example.com/path",
		},
		{
			name:     "remove control chars in query",
			input:    "https://example.com/path?a=1\t&b=2\r\n",
			expected: "https://example.com/path?a=1&b=2",
		},
		{
			name:     "remove control chars in host",
			input:    "https://exa\tmple.com/path",
			expected: "https://example.com/path",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CanonicalizeForGSB(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("CanonicalizeForGSB(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCanonicalizePercentEncoding tests percent encoding normalization
func TestCanonicalizePercentEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "decode unreserved tilde",
			input:    "https://example.com/%7Euser",
			expected: "https://example.com/~user",
		},
		{
			name:     "decode unreserved uppercase tilde",
			input:    "https://example.com/%7euser",
			expected: "https://example.com/~user",
		},
		{
			name:     "decode unreserved letters",
			input:    "https://example.com/%41%42%43",
			expected: "https://example.com/ABC",
		},
		{
			name:     "decode unreserved digits",
			input:    "https://example.com/%30%31%32",
			expected: "https://example.com/012",
		},
		{
			name:     "decode unreserved hyphen",
			input:    "https://example.com/%2Dtest",
			expected: "https://example.com/-test",
		},
		{
			name:     "decode unreserved period",
			input:    "https://example.com/%2Etest",
			expected: "https://example.com/.test",
		},
		{
			name:     "decode unreserved underscore",
			input:    "https://example.com/%5Ftest",
			expected: "https://example.com/_test",
		},
		{
			name:     "uppercase hex for reserved slash",
			input:    "https://example.com/%2f",
			expected: "https://example.com/%2F",
		},
		{
			name:     "uppercase hex for reserved question mark",
			input:    "https://example.com/%3f",
			expected: "https://example.com/%3F",
		},
		{
			name:     "keep reserved chars encoded",
			input:    "https://example.com/%2F%3F%23",
			expected: "https://example.com/%2F%3F%23",
		},
		{
			name:     "mixed unreserved and reserved",
			input:    "https://example.com/%7E%2F%41",
			expected: "https://example.com/~%2FA",
		},
		{
			name:     "percent encode in query",
			input:    "https://example.com/path?key=%7evalue",
			expected: "https://example.com/path?key=~value",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CanonicalizeForGSB(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("CanonicalizeForGSB(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCanonicalizePathNormalization tests path segment normalization
func TestCanonicalizePathNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove single dot",
			input:    "https://example.com/./path",
			expected: "https://example.com/path",
		},
		{
			name:     "remove double dot parent",
			input:    "https://example.com/a/../path",
			expected: "https://example.com/path",
		},
		{
			name:     "remove multiple dots",
			input:    "https://example.com/a/./b/../c",
			expected: "https://example.com/a/c",
		},
		{
			name:     "remove double slash",
			input:    "https://example.com//path",
			expected: "https://example.com/path",
		},
		{
			name:     "remove multiple double slashes",
			input:    "https://example.com///path///to",
			expected: "https://example.com/path/to",
		},
		{
			name:     "complex path normalization",
			input:    "https://example.com/a/./b//c/../d",
			expected: "https://example.com/a/b/d",
		},
		{
			name:     "parent beyond root",
			input:    "https://example.com/../path",
			expected: "https://example.com/path",
		},
		{
			name:     "multiple parents beyond root",
			input:    "https://example.com/../../path",
			expected: "https://example.com/path",
		},
		{
			name:     "dots at end",
			input:    "https://example.com/path/.",
			expected: "https://example.com/path",
		},
		{
			name:     "parent at end",
			input:    "https://example.com/a/b/..",
			expected: "https://example.com/a",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CanonicalizeForGSB(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("CanonicalizeForGSB(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCanonicalizeIPAddresses tests IPv4 and IPv6 address handling
func TestCanonicalizeIPAddresses(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 address",
			input:    "http://192.168.1.1/path",
			expected: "http://192.168.1.1/path",
		},
		{
			name:     "IPv4 with port",
			input:    "http://192.168.1.1:8080/path",
			expected: "http://192.168.1.1:8080/path",
		},
		{
			name:     "IPv4 with default port",
			input:    "http://192.168.1.1:80/path",
			expected: "http://192.168.1.1/path",
		},
		{
			name:     "IPv6 address",
			input:    "http://[2001:db8::1]/path",
			expected: "http://[2001:db8::1]/path",
		},
		{
			name:     "IPv6 with port",
			input:    "http://[2001:db8::1]:8080/path",
			expected: "http://[2001:db8::1]:8080/path",
		},
		{
			name:     "IPv6 localhost",
			input:    "http://[::1]/path",
			expected: "http://[::1]/path",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CanonicalizeForGSB(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("CanonicalizeForGSB(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCanonicalizeGSBOfficialVectors tests against official Google Safe Browsing test vectors
// Reference: https://developers.google.com/safe-browsing/v4/urls-hashing#canonicalization
func TestCanonicalizeGSBOfficialVectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "GSB Example: simple URL",
			input:    "http://www.google.com/",
			expected: "http://www.google.com/",
		},
		{
			name:     "GSB Example: percent encoding",
			input:    "http://host/%25%32%35",
			expected: "http://host/%25",
		},
		{
			name:     "GSB Example: double percent encoding",
			input:    "http://host/%25%32%35%25%32%35",
			expected: "http://host/%25%25",
		},
		{
			name:     "GSB Example: multiple percent encoding",
			input:    "http://host/%2525252525252525",
			expected: "http://host/%25",
		},
		{
			name:     "GSB Example: percent in middle",
			input:    "http://host/asdf%25%32%35asd",
			expected: "http://host/asdf%25asd",
		},
		{
			name:     "GSB Example: mixed percent encoding",
			input:    "http://host/%%%25%32%35asd%%",
			expected: "http://host/%25%25%25asd%25%25",
		},
		{
			name:     "GSB Example: encoded IP and path",
			input:    "http://%31%36%38%2e%31%38%38%2e%39%39%2e%32%36/%2E%73%65%63%75%72%65/%77%77%77%2E%65%62%61%79%2E%63%6F%6D/",
			expected: "http://168.188.99.26/.secure/www.ebay.com/",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CanonicalizeForGSB(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("CanonicalizeForGSB(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCanonicalizeErrors tests error handling for invalid URLs
func TestCanonicalizeErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "empty URL",
			input:       "",
			expectError: true,
			errorSubstr: "cannot be empty",
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectError: true,
			errorSubstr: "cannot be empty",
		},
		{
			name:        "no scheme",
			input:       "example.com/path",
			expectError: true,
			errorSubstr: "must have a scheme",
		},
		{
			name:        "scheme only",
			input:       "http://",
			expectError: true,
			errorSubstr: "must have a host",
		},
		{
			name:        "invalid scheme separator",
			input:       "http:/example.com",
			expectError: true,
			errorSubstr: "must have a host",
		},
		{
			name:        "incomplete percent encoding at end",
			input:       "http://example.com/path%2",
			expectError: true,
			errorSubstr: "incomplete percent encoding",
		},
		{
			name:        "invalid percent encoding hex",
			input:       "http://example.com/path%ZZ",
			expectError: true,
			errorSubstr: "invalid percent encoding",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CanonicalizeForGSB(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("CanonicalizeForGSB(%q) expected error containing %q, got nil error and result %q",
						tt.input, tt.errorSubstr, result)
				} else if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("CanonicalizeForGSB(%q) expected error containing %q, got %q",
						tt.input, tt.errorSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("CanonicalizeForGSB(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

// TestCanonicalizeBatch tests batch canonicalization
func TestCanonicalizeBatch(t *testing.T) {
	c := NewCanonicalizer()

	inputs := []string{
		"HTTP://EXAMPLE.COM/path",
		"https://example.com:443/path",
		"http://example.com/a/../b",
		"invalid-url",
		"https://example.com#fragment",
	}

	expectedResults := []string{
		"http://example.com/path",
		"https://example.com/path",
		"http://example.com/b",
		"",
		"https://example.com/",
	}

	expectedErrors := []bool{
		false, // valid
		false, // valid
		false, // valid
		true,  // invalid
		false, // valid
	}

	results, errors := c.CanonicalizeBatch(inputs)

	if len(results) != len(inputs) {
		t.Fatalf("CanonicalizeBatch returned %d results, want %d", len(results), len(inputs))
	}
	if len(errors) != len(inputs) {
		t.Fatalf("CanonicalizeBatch returned %d errors, want %d", len(errors), len(inputs))
	}

	for i := range inputs {
		if expectedErrors[i] {
			if errors[i] == nil {
				t.Errorf("CanonicalizeBatch[%d](%q) expected error, got nil", i, inputs[i])
			}
		} else {
			if errors[i] != nil {
				t.Errorf("CanonicalizeBatch[%d](%q) unexpected error: %v", i, inputs[i], errors[i])
			}
			if results[i] != expectedResults[i] {
				t.Errorf("CanonicalizeBatch[%d](%q) = %q, want %q", i, inputs[i], results[i], expectedResults[i])
			}
		}
	}
}

// TestCanonicalizeComplex tests complex real-world URLs
func TestCanonicalizeComplex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "complex URL with all features",
			input:    "HTTPS://EXAMPLE.COM:443/./path//%7Euser/../file.html?b=2&a=1#section",
			expected: "https://example.com/path/~user/file.html?b=2&a=1",
		},
		{
			name:     "URL with multiple normalizations",
			input:    "HTTP://User:Pass@EXAMPLE.COM:80/a/./b//c/../%7Efile?z=3&a=1#frag",
			expected: "http://User:Pass@example.com/a/b/~file?z=3&a=1",
		},
		{
			name:     "international domain with path",
			input:    "https://münchen.de/path",
			expected: "https://xn--mnchen-3ya.de/path",
		},
	}

	c := NewCanonicalizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.CanonicalizeForGSB(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("CanonicalizeForGSB(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Benchmark tests

func BenchmarkCanonicalizeSimple(b *testing.B) {
	c := NewCanonicalizer()
	url := "https://example.com/path"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.CanonicalizeForGSB(url)
	}
}

func BenchmarkCanonicalizeComplex(b *testing.B) {
	c := NewCanonicalizer()
	url := "HTTP://EXAMPLE.COM:80/./a//b/../%7Euser/%2Fpath?z=3&a=1#frag"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.CanonicalizeForGSB(url)
	}
}

func BenchmarkCanonicalizeWithEncoding(b *testing.B) {
	c := NewCanonicalizer()
	url := "https://example.com/%7Euser/%41%42%43/%2Fpath"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.CanonicalizeForGSB(url)
	}
}

func BenchmarkCanonicalizeBatch(b *testing.B) {
	c := NewCanonicalizer()
	urls := make([]string, 100)
	for i := 0; i < 100; i++ {
		urls[i] = "https://example.com/path"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.CanonicalizeBatch(urls)
	}
}
