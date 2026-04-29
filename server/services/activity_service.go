package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/your-org/reelqueue-server/models"
)

// ActivityService writes and reads audit-log entries.
type ActivityService struct {
	db *sql.DB
}

// NewActivityService creates an activity service.
func NewActivityService(db *sql.DB) *ActivityService {
	return &ActivityService{db: db}
}

// Log writes one activity entry. Failures are logged and never propagated.
func (s *ActivityService) Log(videoID *int64, action, detail, source string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("activity log panic recovered: %v", recovered)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if source == "" {
		source = "user"
	}

	if _, err := s.db.ExecContext(
		ctx,
		"INSERT INTO activity_log (video_id, action, detail, source) VALUES (?, ?, ?, ?)",
		videoID,
		action,
		detail,
		source,
	); err != nil {
		log.Printf("activity log insert failed: %v", err)
	}
}

// LogAsync writes one activity entry in a recover-protected goroutine.
func (s *ActivityService) LogAsync(videoID *int64, action, detail, source string) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("activity async panic recovered: %v", recovered)
			}
		}()
		s.Log(videoID, action, detail, source)
	}()
}

// ListActivity returns paginated activity entries and total count.
func (s *ActivityService) ListActivity(ctx context.Context, limit, offset int, videoID *int64) ([]models.ActivityEntry, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	where := ""
	args := make([]any, 0, 3)
	if videoID != nil {
		where = " WHERE a.video_id = ?"
		args = append(args, *videoID)
	}

	var total int
	countQuery := "SELECT COUNT(a.id) FROM activity_log a" + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count activity: %w", err)
	}

	query := `SELECT a.id, a.video_id, a.action, a.detail, a.source, a.created_at, v.title
FROM activity_log a
LEFT JOIN videos v ON a.video_id = v.id` + where + `
ORDER BY a.created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list activity: %w", err)
	}
	defer rows.Close()

	entries := make([]models.ActivityEntry, 0)
	for rows.Next() {
		var entry models.ActivityEntry
		var videoIDValue sql.NullInt64
		var detail sql.NullString
		var title sql.NullString

		if err := rows.Scan(&entry.ID, &videoIDValue, &entry.Action, &detail, &entry.Source, &entry.CreatedAt, &title); err != nil {
			return nil, 0, fmt.Errorf("scan activity: %w", err)
		}
		entry.VideoID = int64Ptr(videoIDValue)
		entry.Detail = stringPtr(detail)
		entry.VideoTitle = stringPtr(title)
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate activity: %w", err)
	}

	return entries, total, nil
}

func int64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func stringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}
