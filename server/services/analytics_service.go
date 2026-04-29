package services

import (
	"context"
	"database/sql"
	"fmt"
	"math"
)

// AnalyticsSummary contains all KPI values for the analytics dashboard.
type AnalyticsSummary struct {
	TotalUploaded      int            `json:"totalUploaded"`
	TotalPosted        int            `json:"totalPosted"`
	TotalScheduled     int            `json:"totalScheduled"`
	TotalDraft         int            `json:"totalDraft"`
	TotalPartial       int            `json:"totalPartial"`
	PostsThisWeek      int            `json:"postsThisWeek"`
	PostsLastWeek      int            `json:"postsLastWeek"`
	WeeklyTrend        string         `json:"weeklyTrend"`
	AvgPostsPerWeek    float64        `json:"avgPostsPerWeek"`
	MostActivePlatform *string        `json:"mostActivePlatform"`
	TotalStorageBytes  int64          `json:"totalStorageBytes"`
	PlatformBreakdown  map[string]int `json:"platformBreakdown"`
}

// HeatmapEntry contains a daily posted-video count.
type HeatmapEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// AnalyticsService runs aggregation queries for dashboard views.
type AnalyticsService struct {
	db *sql.DB
}

// NewAnalyticsService creates an analytics service.
func NewAnalyticsService(db *sql.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// GetSummary returns all analytics KPIs.
func (s *AnalyticsService) GetSummary(ctx context.Context) (AnalyticsSummary, error) {
	query := `
WITH
  totals AS (
    SELECT
      COUNT(*) AS totalUploaded,
      COALESCE(SUM(CASE WHEN status = 'posted' THEN 1 ELSE 0 END), 0) AS totalPosted,
      COALESCE(SUM(CASE WHEN status = 'scheduled' THEN 1 ELSE 0 END), 0) AS totalScheduled,
      COALESCE(SUM(CASE WHEN status = 'draft' THEN 1 ELSE 0 END), 0) AS totalDraft,
      COALESCE(SUM(CASE WHEN status = 'partial' THEN 1 ELSE 0 END), 0) AS totalPartial,
      COALESCE(SUM(file_size), 0) AS totalStorageBytes
    FROM videos
  ),
  weekly AS (
    SELECT
      SUM(CASE WHEN status = 'posted' AND posted_at >= strftime('%Y-%m-%d', 'now', 'weekday 0', '-7 days') THEN 1 ELSE 0 END) AS postsThisWeek,
      SUM(CASE WHEN status = 'posted' AND posted_at >= strftime('%Y-%m-%d', 'now', 'weekday 0', '-14 days') AND posted_at < strftime('%Y-%m-%d', 'now', 'weekday 0', '-7 days') THEN 1 ELSE 0 END) AS postsLastWeek
    FROM videos
  ),
  avg_week AS (
    SELECT
      CASE
        WHEN COUNT(DISTINCT strftime('%Y-%W', posted_at)) = 0 THEN 0
        ELSE ROUND(COUNT(*) * 1.0 / COUNT(DISTINCT strftime('%Y-%W', posted_at)), 1)
      END AS avgPostsPerWeek
    FROM videos
    WHERE status = 'posted' AND posted_at IS NOT NULL
  ),
  platform_rank AS (
    SELECT platform AS mostActivePlatform
    FROM videos
    WHERE status = 'posted'
    GROUP BY platform
    ORDER BY COUNT(*) DESC, platform ASC
    LIMIT 1
  )
SELECT totalUploaded, totalPosted, totalScheduled, totalDraft, totalPartial, totalStorageBytes,
       postsThisWeek, postsLastWeek, avgPostsPerWeek, mostActivePlatform
FROM totals
CROSS JOIN weekly
CROSS JOIN avg_week
LEFT JOIN platform_rank ON 1 = 1`

	var summary AnalyticsSummary
	var postsThisWeek sql.NullInt64
	var postsLastWeek sql.NullInt64
	var avgPosts sql.NullFloat64
	var platform sql.NullString
	if err := s.db.QueryRowContext(ctx, query).Scan(
		&summary.TotalUploaded,
		&summary.TotalPosted,
		&summary.TotalScheduled,
		&summary.TotalDraft,
		&summary.TotalPartial,
		&summary.TotalStorageBytes,
		&postsThisWeek,
		&postsLastWeek,
		&avgPosts,
		&platform,
	); err != nil {
		return AnalyticsSummary{}, fmt.Errorf("query analytics summary: %w", err)
	}

	summary.PostsThisWeek = int(postsThisWeek.Int64)
	summary.PostsLastWeek = int(postsLastWeek.Int64)
	summary.AvgPostsPerWeek = avgPosts.Float64
	summary.MostActivePlatform = stringPtr(platform)
	summary.WeeklyTrend = weeklyTrend(summary.PostsThisWeek, summary.PostsLastWeek)

	breakdown, err := s.platformBreakdown(ctx)
	if err != nil {
		return AnalyticsSummary{}, err
	}
	summary.PlatformBreakdown = breakdown

	return summary, nil
}

// GetHeatmap returns posted counts by day for the last 90 days.
func (s *AnalyticsService) GetHeatmap(ctx context.Context) ([]HeatmapEntry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT date(posted_at) AS date, COUNT(id) AS count
FROM videos
WHERE status = 'posted'
  AND posted_at IS NOT NULL
  AND posted_at >= date('now', '-90 days')
GROUP BY date(posted_at)
ORDER BY date ASC`)
	if err != nil {
		return nil, fmt.Errorf("query heatmap: %w", err)
	}
	defer rows.Close()

	entries := make([]HeatmapEntry, 0)
	for rows.Next() {
		var entry HeatmapEntry
		if err := rows.Scan(&entry.Date, &entry.Count); err != nil {
			return nil, fmt.Errorf("scan heatmap: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate heatmap: %w", err)
	}

	return entries, nil
}

func (s *AnalyticsService) platformBreakdown(ctx context.Context) (map[string]int, error) {
	breakdown := map[string]int{"instagram": 0, "tiktok": 0, "youtube": 0, "all": 0}

	rows, err := s.db.QueryContext(ctx, "SELECT platform, COUNT(id) FROM videos GROUP BY platform")
	if err != nil {
		return nil, fmt.Errorf("query platform breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var platform sql.NullString
		var count int
		if err := rows.Scan(&platform, &count); err != nil {
			return nil, fmt.Errorf("scan platform breakdown: %w", err)
		}
		if platform.Valid && platform.String != "" {
			breakdown[platform.String] = count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate platform breakdown: %w", err)
	}

	return breakdown, nil
}

func weeklyTrend(current, previous int) string {
	if previous == 0 {
		if current > 0 {
			return "+100%"
		}
		return "+0%"
	}

	trend := int(math.Round(float64(current-previous) / float64(previous) * 100))
	if trend >= 0 {
		return fmt.Sprintf("+%d%%", trend)
	}
	return fmt.Sprintf("%d%%", trend)
}
