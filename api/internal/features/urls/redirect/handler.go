package redirect

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/domain"
	"github.com/SirNacou/refract/api/internal/infrastructure/publisher"
	"github.com/go-chi/chi/v5"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

type RedirectHandler struct {
	repo           domain.URLRepository
	valkey         valkeyaside.CacheAsideClient
	clickPublisher *publisher.ClicksPublisher
	redirectKey    string
}

func NewRedirectHandler(valkey valkeyaside.CacheAsideClient, repo domain.URLRepository, publisher *publisher.ClicksPublisher, cfg *config.Config) *RedirectHandler {
	return &RedirectHandler{
		valkey:         valkey,
		repo:           repo,
		clickPublisher: publisher,
		redirectKey:    cfg.Valkey.RedirectKey,
	}
}

func (h *RedirectHandler) Handle(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	slog.Info("Handling redirect", "short_code", shortCode)

	key := strings.Replace(h.redirectKey, "{short_code}", shortCode, 1)
	url, err := h.valkey.Get(r.Context(), time.Minute, key, func(ctx context.Context, key string) (val string, err error) {
		url, err := h.repo.GetActiveURLByShortCode(ctx, domain.ShortCode(shortCode))
		if err != nil {
			return "", err
		}

		maxTTL := time.Hour * 24 * 365
		if url.ExpiresAt != nil {
			ttl := time.Until(*url.ExpiresAt)
			if ttl < maxTTL {
				maxTTL = ttl
			}
		}
		valkeyaside.OverrideCacheTTL(ctx, maxTTL)

		return url.OriginalURL, nil
	})

	if err != nil {
		WriteNotFoundPage(w)
		return
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	err = h.clickPublisher.Publish(r.Context(), &publisher.ClicksPublisherRequest{
		ShortCode: shortCode,
		IPAddress: host,
		UserAgent: r.UserAgent(),
		Referer:   r.Referer(),
		ClickedAt: time.Now(),
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to track click", "error", err)
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func WriteNotFoundPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html>
		<head><title>404 Not Found</title></head>
		<body>
			<h1>404 Not Found</h1>
			<p>The requested URL was not found on this server.</p>
		</body>
	</html>`)
}
