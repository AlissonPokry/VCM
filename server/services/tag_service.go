package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/your-org/reelqueue-server/models"
)

// TagService manages tag normalization and video/tag joins.
type TagService struct {
	db *sql.DB
}

// NewTagService creates a tag service.
func NewTagService(db *sql.DB) *TagService {
	return &TagService{db: db}
}

// NormalizeTagNames trims, deduplicates, and drops empty tag names.
func NormalizeTagNames(tagNames []string) []string {
	seen := make(map[string]struct{}, len(tagNames))
	normalized := make([]string, 0, len(tagNames))

	for _, tag := range tagNames {
		name := strings.TrimSpace(tag)
		if name == "" {
			continue
		}

		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, name)
	}

	return normalized
}

// UpsertTags replaces all tags for a video inside the provided transaction.
func (s *TagService) UpsertTags(ctx context.Context, tx *sql.Tx, videoID int64, tagNames []string) error {
	normalized := NormalizeTagNames(tagNames)

	if _, err := tx.ExecContext(ctx, "DELETE FROM video_tags WHERE video_id = ?", videoID); err != nil {
		return fmt.Errorf("clear video tags: %w", err)
	}

	for _, name := range normalized {
		if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO tags (name) VALUES (?)", name); err != nil {
			return fmt.Errorf("insert tag %q: %w", name, err)
		}

		var tagID int64
		if err := tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = ? COLLATE NOCASE", name).Scan(&tagID); err != nil {
			return fmt.Errorf("lookup tag %q: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO video_tags (video_id, tag_id) VALUES (?, ?)", videoID, tagID); err != nil {
			return fmt.Errorf("attach tag %q: %w", name, err)
		}
	}

	return nil
}

// AssembleTagsForVideo returns tag names for one video in ascending order.
func (s *TagService) AssembleTagsForVideo(ctx context.Context, videoID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.name
FROM tags t
JOIN video_tags vt ON vt.tag_id = t.id
WHERE vt.video_id = ?
ORDER BY t.name ASC`, videoID)
	if err != nil {
		return nil, fmt.Errorf("query tags for video: %w", err)
	}
	defer rows.Close()

	tags := make([]string, 0)
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return tags, nil
}

// AttachTags adds tag arrays to each video.
func (s *TagService) AttachTags(ctx context.Context, videos []models.Video) ([]models.Video, error) {
	if len(videos) == 0 {
		return videos, nil
	}

	ids := make([]any, 0, len(videos))
	for _, video := range videos {
		ids = append(ids, video.ID)
	}

	query := `SELECT vt.video_id, t.name
FROM video_tags vt
JOIN tags t ON vt.tag_id = t.id
WHERE vt.video_id IN (` + placeholders(len(ids)) + `)
ORDER BY t.name ASC`

	rows, err := s.db.QueryContext(ctx, query, ids...)
	if err != nil {
		return nil, fmt.Errorf("query video tags: %w", err)
	}
	defer rows.Close()

	byVideo := make(map[int64][]string, len(videos))
	for rows.Next() {
		var videoID int64
		var tag string
		if err := rows.Scan(&videoID, &tag); err != nil {
			return nil, fmt.Errorf("scan video tag: %w", err)
		}
		byVideo[videoID] = append(byVideo[videoID], tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate video tags: %w", err)
	}

	for i := range videos {
		tags := byVideo[videos[i].ID]
		if tags == nil {
			tags = []string{}
		}
		videos[i].Tags = tags
	}

	return videos, nil
}

// ListTags returns tags currently attached to videos, sorted by usage.
func (s *TagService) ListTags(ctx context.Context) ([]models.Tag, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.id, t.name, COUNT(v.id) AS count
FROM tags t
LEFT JOIN video_tags vt ON t.id = vt.tag_id
LEFT JOIN videos v ON vt.video_id = v.id
GROUP BY t.id, t.name
HAVING COUNT(v.id) > 0
ORDER BY count DESC, t.name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	defer rows.Close()

	tags := make([]models.Tag, 0)
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Count); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return tags, nil
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	items := make([]string, count)
	for i := range items {
		items[i] = "?"
	}
	return strings.Join(items, ",")
}
