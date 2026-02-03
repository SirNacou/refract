package getdashboard

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/SirNacou/refract/api/internal/domain"
)

type Query struct {
	UserID string
}
type QueryResult struct {
	TotalURLs      uint `json:"total_urls"`
	TotalClicks    uint `json:"total_clicks"`
	ActiveURLs     uint `json:"active_urls"`
	ClicksThisWeek uint `json:"clicks_this_week"`

	ClickTrends      []ClickTrend     `json:"click_trends"`
	RecentActivities []RecentActivity `json:"recent_activities"`

	TopURLs []TopURL `json:"top_urls"`
}

type ClickTrend struct {
	Date   time.Time `json:"date" ch:"date"`
	Clicks uint64    `json:"clicks" ch:"clicks"`
}

type RecentActivity struct {
	Timestamp time.Time `json:"timestamp" ch:"timestamp"`
	Activity  string    `json:"activity" ch:"activity"`
}

type TopURL struct {
	OriginalURL    string       `json:"original_url" ch:"original_url"`
	ShortURL       string       `json:"short_url" ch:"short_url"`
	Clicks         uint64       `json:"clicks" ch:"clicks"`
	ThisWeekTrends []ClickTrend `json:"this_week_trends" ch:"this_week_trends"`
}

type QueryHandler struct {
	repo           domain.URLRepository
	ch             clickhouse.Conn
	defaultBaseURL string
}

func NewQueryHandler(repo domain.URLRepository, ch clickhouse.Conn, defaultBaseURL string) *QueryHandler {
	return &QueryHandler{repo: repo, ch: ch, defaultBaseURL: defaultBaseURL}
}

func (h *QueryHandler) Handle(ctx context.Context, q *Query) (*QueryResult, error) {
	totalURLs, err := h.repo.CountByUser(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	activeURLs, err := h.repo.CountActiveByUser(ctx, q.UserID)
	if err != nil {
		return nil, err
	}

	var totalClicks uint64
	err = h.ch.QueryRow(ctx, `
		SELECT sumMerge(clicks) as total_clicks
		FROM refract.url_daily_stats
	`).Scan(&totalClicks)
	if err != nil {
		return nil, err
	}

	var clicksThisWeek uint64
	err = h.ch.QueryRow(ctx, `
		SELECT sumMerge(clicks) as clicks_this_week
		FROM refract.url_daily_stats
		WHERE date >= today() - INTERVAL 7 DAY
	`).Scan(&clicksThisWeek)
	if err != nil {
		return nil, err
	}

	clickTrends := make([]ClickTrend, 0)
	rows, err := h.ch.Query(ctx, `
		SELECT date, sumMerge(clicks) as clicks
		FROM refract.url_daily_stats
		WHERE date >= today() - INTERVAL 30 DAY
		GROUP BY date
		ORDER BY date ASC
	`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ct ClickTrend
		if err := rows.ScanStruct(&ct); err != nil {
			return nil, err
		}
		clickTrends = append(clickTrends, ct)
	}

	rows, err = h.ch.Query(ctx, `
	SELECT short_code, clicked_at FROM refract.clicks
	ORDER BY clicked_at DESC
	LIMIT 5
	`)
	if err != nil {
		return nil, err
	}

	recentActivities := make([]RecentActivity, 0)
	for rows.Next() {
		var ra RecentActivity
		var shortCode string
		if err := rows.Scan(&shortCode, &ra.Timestamp); err != nil {
			return nil, err
		}
		ra.Activity = "Link " + strings.Join([]string{h.defaultBaseURL, shortCode}, "/") + " was clicked"

		recentActivities = append(recentActivities, ra)
	}

	rows, err = h.ch.Query(ctx, `
		SELECT
			u.original_url,
			u.short_code,
			sum(week.clicks) AS clicks,
			groupArray(week.date) AS this_week_trends_dates,
			groupArray(week.clicks) AS this_week_clicks
		FROM refract.urls u
		JOIN (
			SELECT
				short_code,
				date,
				sumMerge(clicks) AS clicks
			FROM refract.url_daily_stats
			WHERE date >= today() - INTERVAL 7 DAY
			GROUP BY short_code, date
		) week ON u.short_code = week.short_code
		WHERE u.created_by = ?
		GROUP BY u.short_code, u.original_url
		ORDER BY clicks DESC
		LIMIT 10
	`, q.UserID)
	if err != nil {
		return nil, err
	}

	topURLs := make([]TopURL, 0)
	for rows.Next() {
		var tu struct {
			OriginalURL    string      `ch:"original_url"`
			ShortCode      string      `ch:"short_code"`
			Clicks         uint64      `ch:"clicks"`
			ThisWeekTrends []time.Time `ch:"this_week_trends_dates"`
			ThisWeekClicks []uint64    `ch:"this_week_clicks"`
		}
		if err := rows.ScanStruct(&tu); err != nil {
			slog.ErrorContext(ctx, "Failed to scan top URL", "error", err)
			return nil, err
		}
		topURLs = append(topURLs, TopURL{
			OriginalURL: tu.OriginalURL,
			ShortURL:    strings.Join([]string{h.defaultBaseURL, tu.ShortCode}, "/"),
			Clicks:      tu.Clicks,
			ThisWeekTrends: func() []ClickTrend {
				trends := make([]ClickTrend, 0, len(tu.ThisWeekTrends))
				for i, t := range tu.ThisWeekTrends {
					slog.InfoContext(ctx, "ThisWeekTrends item", "key", t)

					trends = append(trends, ClickTrend{
						Date:   t,
						Clicks: tu.ThisWeekClicks[i],
					})
				}
				return trends
			}(),
		})
	}

	return &QueryResult{
		TotalURLs:        uint(totalURLs),
		TotalClicks:      uint(totalClicks),
		ActiveURLs:       uint(activeURLs),
		ClicksThisWeek:   uint(clicksThisWeek),
		ClickTrends:      clickTrends,
		RecentActivities: recentActivities,
		TopURLs:          topURLs,
	}, nil
}
