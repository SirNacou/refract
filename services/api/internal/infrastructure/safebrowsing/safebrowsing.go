package safebrowsing

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/SirNacou/refract/services/api/internal/infrastructure/cache"
	pb "github.com/SirNacou/refract/services/api/internal/infrastructure/safebrowsing/sbproto"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

type SafeBrowsing struct {
	apiKey string
	cache  valkeyaside.TypedCacheAsideClient[pb.SearchHashesResponse]
}

func NewSafeBrowsing(apiKey, redisURL string, cache *cache.RedisCache) (*SafeBrowsing, error) {
	client, err := valkeyaside.NewClient(valkeyaside.ClientOption{
		ClientOption: valkey.MustParseURL(redisURL),
	})
	if err != nil {
		return nil, err
	}

	typedClient := valkeyaside.NewTypedCacheAsideClient(client,
		func(t *pb.SearchHashesResponse) (string, error) {
			b, err := json.Marshal(t)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		func(s string) (*pb.SearchHashesResponse, error) {
			var t *pb.SearchHashesResponse
			if err := json.Unmarshal([]byte(s), &t); err != nil {
				return nil, err
			}
			return t, nil
		})

	return &SafeBrowsing{apiKey: apiKey, cache: typedClient}, nil
}

func getCanonicalHash(rawURL string) ([]byte, error) {
	// A strictly correct implementation is complex (RFC 3986).
	// For this example, we assume the URL is already somewhat clean
	// and we focus on the hashing mechanism.
	// In production, ensure you strip fragments, lower-case hostnames, etc.
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// Minimal canonicalization: lowercase host, remove fragments
	canonical := strings.ToLower(u.Host) + u.Path
	if u.RawQuery != "" {
		canonical += "?" + u.RawQuery
	}

	hash := sha256.Sum256([]byte(canonical))
	return hash[:], nil
}

func (s *SafeBrowsing) CheckURLv5Proto(ctx context.Context, targetURL string) (bool, error) {
	fullHash, _ := getCanonicalHash(targetURL)
	prefix := fullHash[:4]
	encodedPrefix := base64.URLEncoding.EncodeToString(prefix)

	result, err := s.cache.Get(ctx, time.Hour, encodedPrefix, func(ctx context.Context, key string) (val *pb.SearchHashesResponse, err error) {

		// NOTE: Removed "&alt=json". We WANT the binary default now.
		reqURL := fmt.Sprintf("https://safebrowsing.googleapis.com/v5/hashes:search?key=%s&hashPrefixes=%s", s.apiKey, encodedPrefix)

		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return nil, err
		}

		// Tell Google we accept Protobuf (Standard MIME type)
		req.Header.Set("Accept", "application/x-protobuf")

		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Read the raw binary bytes
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("api error: %d", resp.StatusCode)
		}

		// --- THE MAGIC HAPPENS HERE ---
		// We unmarshal the raw bytes into our generated Go struct
		var result pb.SearchHashesResponse
		if err := proto.Unmarshal(bodyBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to parse protobuf: %w", err)
		}

		duration := time.Second * 60
		if result.GetCacheDuration() != nil {
			slog.DebugContext(ctx, "cache duration", "duration", result.GetCacheDuration().AsDuration())
			duration = result.GetCacheDuration().AsDuration()
		}

		valkeyaside.OverrideCacheTTL(ctx, duration)

		return &result, nil
	})

	if err != nil {
		return false, err
	}

	for _, match := range result.FullHashes {
		if string(match.FullHash) == string(fullHash) {
			return true, nil // Match found
		}
	}

	return false, nil
}
