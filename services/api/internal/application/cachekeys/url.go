package cachekeys

import "github.com/SirNacou/refract/services/api/internal/domain/url"

func RedirectCacheKey(shortCode url.ShortCode) string {
	return "redirect:" + shortCode.String()
}
