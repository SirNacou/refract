package domain

import (
	"fmt"
	"strings"
)

const (
	// Base58 Alphabet (Bitcoin style)
	// Removes: 0 (zero), O (capital o), I (capital i), l (lower L)
	alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	base     = uint64(len(alphabet))

	// OPTIONAL: Obfuscation Constants (Knuth's Multiplicative Hash)
	// This prime should be large and ideally a little random.
	// This specific prime is just an example.
	prime        = uint64(982451653)
	primeInverse = uint64(6268802909) // Modular multiplicative inverse of prime
	xorKey       = uint64(123456789)  // Random XOR key
)

type ShortCode string

func NewShortCode(s string) (*ShortCode, error) {
	if s == "" {
		return nil, fmt.Errorf("short code cannot be empty")
	}

	if len(s) > 20 {
		return nil, fmt.Errorf("short code cannot be longer than 20 characters")
	}

	sc := ShortCode(s)
	return &sc, nil
}

func (s ShortCode) String() string {
	return string(s)
}

// Encode: Scramble -> Encode
func GenerateShortcode(snowflakeID SnowflakeID) ShortCode {
	// 1. Obfuscate (Scramble the bits so it looks random)
	// We use integer overflow intentionally here for the modular arithmetic
	scrambled := (uint64(snowflakeID) * prime) ^ xorKey

	// 2. Base58 Encode
	return ShortCode(encodeBase58(scrambled))
}

// Decode: Decode -> Unscramble
func ResolveShortcode(code string) (SnowflakeID, error) {
	// 1. Base58 Decode
	scrambled, err := decodeBase58(code)
	if err != nil {
		return 0, err
	}

	// 2. De-obfuscate (Reverse the math)
	// Apply XOR again to reverse it
	unXored := scrambled ^ xorKey
	// Multiply by the modular inverse to reverse the multiplication
	// Note: You must calculate the correct inverse for your specific prime!
	original := unXored * primeInverse

	return SnowflakeID(original), nil
}

// --- Low Level Helpers ---

func encodeBase58(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}
	var sb strings.Builder
	for n > 0 {
		mod := n % base
		sb.WriteByte(alphabet[mod])
		n = n / base
	}
	// Reverse string
	chars := []rune(sb.String())
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func decodeBase58(s string) (uint64, error) {
	var n uint64
	for _, char := range s {
		index := strings.IndexRune(alphabet, char)
		if index == -1 {
			return 0, fmt.Errorf("invalid char: %v", char)
		}
		n = n*base + uint64(index)
	}
	return n, nil
}
