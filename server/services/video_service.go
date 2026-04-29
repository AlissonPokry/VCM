package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	appmw "github.com/your-org/reelqueue-server/middleware"
	"github.com/your-org/reelqueue-server/models"
)

// VideoFilters contains all supported list query filters.
type VideoFilters struct {
	Status   string
	Platform string
	Tag      string
	Search   string
	Sort     string
	Order    string
	DateFrom string
	DateTo   string
}

// UploadedVideoInput contains metadata for a saved uploaded file.
type UploadedVideoInput struct {
	Title        string
	Description  string
	Filename     string
	OriginalName string
	FileSize     int64
	Platform     string
	ScheduledAt  string
	TagsRaw      string
}

// VideoUpdate contains optional video metadata changes.
type VideoUpdate struct {
	Title         *string
	Description   *string
	Platform      *string
	ScheduledAt   *string
	N8NWorkflowID *string
	Tags          *[]string
}

// N8NPostedInput contains the posted webhook payload.
type N8NPostedInput struct {
	VideoID     int64
	PostedAt    *string
	Platform    *string
	ExecutionID *string
}

// N8NFailedInput contains the failed webhook payload.
type N8NFailedInput struct {
	VideoID     int64
	Error       string
	ExecutionID *string
	Platform    *string
}

// VideoService owns video business logic and persistence.
type VideoService struct {
	db              *sql.DB
	tags            *TagService
	activity        *ActivityService
	uploadDir       string
	projectRoot     string
	dispatchTimeout time.Duration
}

// NewVideoService creates a video service.
func NewVideoService(db *sql.DB, tags *TagService, activity *ActivityService, uploadDir, projectRoot string, dispatchTimeout time.Duration) *VideoService {
	return &VideoService{
		db:              db,
		tags:            tags,
		activity:        activity,
		uploadDir:       uploadDir,
		projectRoot:     projectRoot,
		dispatchTimeout: dispatchTimeout,
	}
}

// List returns filtered videos and total count.
func (s *VideoService) List(ctx context.Context, filters VideoFilters) ([]models.Video, int, error) {
	where, args := buildVideoWhere(filters)
	sort := sanitizeSort(filters.Sort)
	order := sanitizeOrder(filters.Order)

	query := `SELECT v.id, v.title, v.description, v.filename, v.original_name, v.file_size,
v.duration, v.thumbnail, v.platform, v.status, v.scheduled_at, v.posted_at,
v.n8n_workflow_id, v.n8n_execution_log, v.created_at, v.updated_at
FROM videos v` + where + " ORDER BY v." + sort + " " + order

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list videos: %w", err)
	}
	defer rows.Close()

	videos, err := scanVideos(rows)
	if err != nil {
		return nil, 0, err
	}

	countQuery := "SELECT COUNT(DISTINCT v.id) FROM videos v" + where
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count videos: %w", err)
	}

	videos, err = s.tags.AttachTags(ctx, videos)
	if err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// GetByID returns one video with tags.
func (s *VideoService) GetByID(ctx context.Context, id int64) (models.Video, error) {
	video, err := s.getRawByID(ctx, id)
	if err != nil {
		return models.Video{}, err
	}

	withTags, err := s.tags.AttachTags(ctx, []models.Video{video})
	if err != nil {
		return models.Video{}, err
	}
	return withTags[0], nil
}

// CreateFromUpload stores a video row for an already-saved file.
func (s *VideoService) CreateFromUpload(ctx context.Context, input UploadedVideoInput) (models.Video, error) {
	if input.Filename == "" {
		return models.Video{}, appmw.NewAppError("Video file is required", 422, "FILE_REQUIRED")
	}
	if input.Platform == "" {
		input.Platform = "instagram"
	}
	if err := ensurePlatform(input.Platform); err != nil {
		return models.Video{}, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = strings.TrimSuffix(input.OriginalName, filepath.Ext(input.OriginalName))
	}
	if title == "" {
		return models.Video{}, appmw.NewAppError("Title is required", 422, "TITLE_REQUIRED")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Video{}, fmt.Errorf("begin create video transaction: %w", err)
	}
	defer rollbackQuietly(tx)

	result, err := tx.ExecContext(ctx, `INSERT INTO videos
(title, description, filename, original_name, file_size, platform, status, scheduled_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, 'scheduled', ?, ?, ?)`,
		title,
		nullableString(input.Description),
		input.Filename,
		input.OriginalName,
		input.FileSize,
		input.Platform,
		nullableString(input.ScheduledAt),
		now,
		now,
	)
	if err != nil {
		return models.Video{}, fmt.Errorf("insert video: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.Video{}, fmt.Errorf("read inserted video id: %w", err)
	}

	if err := s.tags.UpsertTags(ctx, tx, id, ParseTags(input.TagsRaw)); err != nil {
		return models.Video{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Video{}, fmt.Errorf("commit create video: %w", err)
	}

	s.activity.LogAsync(&id, "uploaded", fmt.Sprintf("Video %q uploaded (%dMB)", title, input.FileSize/1024/1024), "user")
	return s.GetByID(ctx, id)
}

// Update changes video metadata and tags.
func (s *VideoService) Update(ctx context.Context, id int64, update VideoUpdate) (models.Video, error) {
	current, err := s.GetByID(ctx, id)
	if err != nil {
		return models.Video{}, err
	}

	if update.Platform != nil {
		if err := ensurePlatform(*update.Platform); err != nil {
			return models.Video{}, err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Video{}, fmt.Errorf("begin update video transaction: %w", err)
	}
	defer rollbackQuietly(tx)

	set := make([]string, 0, 6)
	args := make([]any, 0, 8)

	if update.Title != nil {
		title := strings.TrimSpace(*update.Title)
		if title == "" {
			return models.Video{}, appmw.NewAppError("Title is required", 422, "TITLE_REQUIRED")
		}
		set = append(set, "title = ?")
		args = append(args, title)
	}
	if update.Description != nil {
		set = append(set, "description = ?")
		args = append(args, nullableString(*update.Description))
	}
	if update.Platform != nil {
		set = append(set, "platform = ?")
		args = append(args, *update.Platform)
	}
	if update.ScheduledAt != nil {
		set = append(set, "scheduled_at = ?")
		args = append(args, nullableString(*update.ScheduledAt))
	}
	if update.N8NWorkflowID != nil {
		set = append(set, "n8n_workflow_id = ?")
		args = append(args, nullableString(*update.N8NWorkflowID))
	}

	if len(set) > 0 {
		set = append(set, "updated_at = ?")
		args = append(args, time.Now().UTC().Format(time.RFC3339), id)

		query := "UPDATE videos SET " + strings.Join(set, ", ") + " WHERE id = ?"
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return models.Video{}, fmt.Errorf("update video: %w", err)
		}
	}

	if update.Tags != nil {
		if err := s.tags.UpsertTags(ctx, tx, id, *update.Tags); err != nil {
			return models.Video{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Video{}, fmt.Errorf("commit update video: %w", err)
	}

	title := current.Title
	if update.Title != nil && strings.TrimSpace(*update.Title) != "" {
		title = strings.TrimSpace(*update.Title)
	}
	s.activity.LogAsync(&id, "edited", fmt.Sprintf("Metadata updated for %q", title), "user")
	return s.GetByID(ctx, id)
}

// UpdateStatus changes a video status and posted timestamp.
func (s *VideoService) UpdateStatus(ctx context.Context, id int64, status string) (models.Video, error) {
	if err := ensureStatus(status); err != nil {
		return models.Video{}, err
	}
	if _, err := s.GetByID(ctx, id); err != nil {
		return models.Video{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var postedAt any
	if status == "posted" {
		postedAt = now
	}

	if _, err := s.db.ExecContext(ctx, "UPDATE videos SET status = ?, posted_at = ?, updated_at = ? WHERE id = ?", status, postedAt, now, id); err != nil {
		return models.Video{}, fmt.Errorf("update video status: %w", err)
	}

	s.activity.LogAsync(&id, "status_changed", fmt.Sprintf("Status changed to %s", status), "user")
	return s.GetByID(ctx, id)
}

// DeleteVideo deletes a video row and its local media files.
func (s *VideoService) DeleteVideo(ctx context.Context, id int64) (map[string]int64, error) {
	video, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin delete video transaction: %w", err)
	}
	defer rollbackQuietly(tx)

	if _, err := tx.ExecContext(ctx, "DELETE FROM video_tags WHERE video_id = ?", id); err != nil {
		return nil, fmt.Errorf("delete video tags: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM videos WHERE id = ?", id); err != nil {
		return nil, fmt.Errorf("delete video: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit delete video: %w", err)
	}

	removeQuietly(filepath.Join(s.uploadDir, video.Filename))
	if video.Thumbnail != nil {
		removeQuietly(s.resolveStoredPath(*video.Thumbnail))
	}

	s.activity.LogAsync(nil, "deleted", fmt.Sprintf("Video %q was permanently deleted", video.Title), "user")
	return map[string]int64{"id": id}, nil
}

// BulkAction performs supported bulk operations.
func (s *VideoService) BulkAction(ctx context.Context, ids []int64, action string, scheduledAt string) (any, error) {
	safeIDs := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			safeIDs = append(safeIDs, id)
		}
	}
	if len(safeIDs) == 0 {
		return nil, appmw.NewAppError("At least one video id is required", 422, "IDS_REQUIRED")
	}

	switch action {
	case "delete":
		results := make([]map[string]int64, 0, len(safeIDs))
		for _, id := range safeIDs {
			result, err := s.DeleteVideo(ctx, id)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
		return results, nil
	case "draft":
		if err := s.bulkStatus(ctx, safeIDs, "draft", "", nil); err != nil {
			return nil, err
		}
		for _, id := range safeIDs {
			videoID := id
			s.activity.LogAsync(&videoID, "status_changed", "Status changed to draft", "user")
		}
		return safeIDs, nil
	case "reschedule":
		if strings.TrimSpace(scheduledAt) == "" {
			return nil, appmw.NewAppError("scheduled_at is required for reschedule", 422, "SCHEDULED_AT_REQUIRED")
		}
		if err := s.bulkStatus(ctx, safeIDs, "scheduled", scheduledAt, &scheduledAt); err != nil {
			return nil, err
		}
		for _, id := range safeIDs {
			videoID := id
			s.activity.LogAsync(&videoID, "status_changed", fmt.Sprintf("Status changed to scheduled for %s", scheduledAt), "user")
		}
		return safeIDs, nil
	default:
		return nil, appmw.NewAppError("Invalid bulk action", 422, "INVALID_BULK_ACTION")
	}
}

// GetDueScheduled returns all due scheduled videos for n8n.
func (s *VideoService) GetDueScheduled(ctx context.Context) ([]models.Video, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT v.id, v.title, v.description, v.filename, v.original_name, v.file_size,
v.duration, v.thumbnail, v.platform, v.status, v.scheduled_at, v.posted_at,
v.n8n_workflow_id, v.n8n_execution_log, v.created_at, v.updated_at
FROM videos v
WHERE v.status = 'scheduled'
  AND v.scheduled_at IS NOT NULL
  AND datetime(v.scheduled_at) <= datetime('now')
ORDER BY v.scheduled_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("query due scheduled videos: %w", err)
	}
	defer rows.Close()

	videos, err := scanVideos(rows)
	if err != nil {
		return nil, err
	}
	videos, err = s.tags.AttachTags(ctx, videos)
	if err != nil {
		return nil, err
	}

	for _, video := range videos {
		videoID := video.ID
		s.activity.LogAsync(&videoID, "n8n_queued", fmt.Sprintf("Picked up by n8n for posting to %s", video.Platform), "n8n")
	}
	return videos, nil
}

// MarkPostedFromN8N handles n8n success callbacks.
func (s *VideoService) MarkPostedFromN8N(ctx context.Context, input N8NPostedInput) (models.Video, error) {
	if input.VideoID <= 0 {
		return models.Video{}, appmw.NewAppError("video_id is required", 422, "VIDEO_ID_REQUIRED")
	}

	video, err := s.GetByID(ctx, input.VideoID)
	if err != nil {
		return models.Video{}, err
	}

	if video.Platform == "all" {
		return s.markPostedAcrossPlatforms(ctx, video)
	}

	platform := video.Platform
	if input.Platform != nil && *input.Platform != "" {
		platform = *input.Platform
	}
	if err := ensurePlatform(platform); err != nil {
		return models.Video{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	postedAt := now
	if input.PostedAt != nil && strings.TrimSpace(*input.PostedAt) != "" {
		postedAt = *input.PostedAt
	}

	entry := models.N8NExecutionEntry{
		ExecutionID: input.ExecutionID,
		Timestamp:   now,
		Result:      "posted",
		Platform:    &platform,
		Error:       nil,
	}

	if err := s.updateExecutionAndStatus(ctx, video.ID, []models.N8NExecutionEntry{entry}, "posted", &postedAt, &platform); err != nil {
		return models.Video{}, err
	}

	s.activity.LogAsync(&video.ID, "n8n_posted", fmt.Sprintf("Successfully posted to %s at %s", platform, postedAt), "n8n")
	return s.GetByID(ctx, video.ID)
}

// MarkFailedFromN8N handles n8n failure callbacks without changing status.
func (s *VideoService) MarkFailedFromN8N(ctx context.Context, input N8NFailedInput) (models.Video, error) {
	if input.VideoID <= 0 {
		return models.Video{}, appmw.NewAppError("video_id is required", 422, "VIDEO_ID_REQUIRED")
	}

	video, err := s.GetByID(ctx, input.VideoID)
	if err != nil {
		return models.Video{}, err
	}

	message := strings.TrimSpace(input.Error)
	if message == "" {
		message = "Unknown error"
	}

	entry := models.N8NExecutionEntry{
		ExecutionID: input.ExecutionID,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Result:      "failed",
		Platform:    input.Platform,
		Error:       &message,
	}

	if err := s.appendExecutionEntries(ctx, video.ID, []models.N8NExecutionEntry{entry}); err != nil {
		return models.Video{}, err
	}

	s.activity.LogAsync(&video.ID, "n8n_failed", fmt.Sprintf("n8n execution failed: %s", message), "n8n")
	return s.GetByID(ctx, video.ID)
}

// ParseTags converts JSON or comma-delimited tag payloads to normalized tags.
func ParseTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}

	var parsed []string
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		return NormalizeTagNames(parsed)
	}

	cleaned := strings.Trim(raw, "[]")
	cleaned = strings.ReplaceAll(cleaned, `"`, "")
	cleaned = strings.ReplaceAll(cleaned, `'`, "")
	return NormalizeTagNames(strings.Split(cleaned, ","))
}

func (s *VideoService) markPostedAcrossPlatforms(ctx context.Context, video models.Video) (models.Video, error) {
	platforms := ResolvePlatforms(video.Platform)
	videoPath := filepath.Join(s.uploadDir, video.Filename)
	results := DispatchToPlatforms(video.ID, platforms, videoPath, s.dispatchTimeout)

	entries := make([]models.N8NExecutionEntry, 0, len(results))
	successes := 0
	for _, result := range results {
		timestamp := result.PostedAt
		if timestamp.IsZero() {
			timestamp = time.Now().UTC()
		}

		executionID := nullableString(result.ExecutionID)
		platform := result.Platform
		var errPtr *string
		outcome := "posted"
		if !result.Success {
			outcome = "failed"
			errPtr = nullableString(result.Error)
		} else {
			successes++
		}

		entries = append(entries, models.N8NExecutionEntry{
			ExecutionID: executionID,
			Timestamp:   timestamp.Format(time.RFC3339),
			Result:      outcome,
			Platform:    &platform,
			Error:       errPtr,
		})
	}

	status := "posted"
	if successes != len(results) {
		status = "partial"
	}
	postedAt := time.Now().UTC().Format(time.RFC3339)
	if err := s.updateExecutionAndStatus(ctx, video.ID, entries, status, &postedAt, nil); err != nil {
		return models.Video{}, err
	}

	s.activity.LogAsync(&video.ID, "n8n_posted", fmt.Sprintf("Posted to %d/%d platforms", successes, len(results)), "n8n")
	return s.GetByID(ctx, video.ID)
}

func (s *VideoService) getRawByID(ctx context.Context, id int64) (models.Video, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, title, description, filename, original_name, file_size,
duration, thumbnail, platform, status, scheduled_at, posted_at,
n8n_workflow_id, n8n_execution_log, created_at, updated_at
FROM videos
WHERE id = ?`, id)

	video, err := scanVideo(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Video{}, appmw.NewAppError("Video not found", 404, "VIDEO_NOT_FOUND")
	}
	if err != nil {
		return models.Video{}, err
	}
	return video, nil
}

func (s *VideoService) bulkStatus(ctx context.Context, ids []int64, status string, scheduledAt string, scheduledPtr *string) error {
	args := make([]any, 0, len(ids)+4)
	for _, id := range ids {
		args = append(args, id)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	query := "UPDATE videos SET status = ?, posted_at = NULL, updated_at = ?"
	prefixArgs := []any{status, now}
	if scheduledPtr != nil {
		query += ", scheduled_at = ?"
		prefixArgs = append(prefixArgs, scheduledAt)
	}
	query += " WHERE id IN (" + placeholders(len(ids)) + ")"
	args = append(prefixArgs, args...)

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("bulk update videos: %w", err)
	}
	return nil
}

func (s *VideoService) updateExecutionAndStatus(ctx context.Context, id int64, entries []models.N8NExecutionEntry, status string, postedAt *string, platform *string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin execution update transaction: %w", err)
	}
	defer rollbackQuietly(tx)

	if err := appendExecutionEntriesTx(ctx, tx, id, entries); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	set := []string{"status = ?", "posted_at = ?", "updated_at = ?"}
	args := []any{status, postedAt, now}
	if platform != nil {
		set = append(set, "platform = ?")
		args = append(args, *platform)
	}
	args = append(args, id)

	query := "UPDATE videos SET " + strings.Join(set, ", ") + " WHERE id = ?"
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("update execution status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit execution update: %w", err)
	}
	return nil
}

func (s *VideoService) appendExecutionEntries(ctx context.Context, id int64, entries []models.N8NExecutionEntry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin append execution transaction: %w", err)
	}
	defer rollbackQuietly(tx)

	if err := appendExecutionEntriesTx(ctx, tx, id, entries); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "UPDATE videos SET updated_at = ? WHERE id = ?", time.Now().UTC().Format(time.RFC3339), id); err != nil {
		return fmt.Errorf("touch video after execution append: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit append execution: %w", err)
	}
	return nil
}

func appendExecutionEntriesTx(ctx context.Context, tx *sql.Tx, id int64, entries []models.N8NExecutionEntry) error {
	var raw sql.NullString
	if err := tx.QueryRowContext(ctx, "SELECT n8n_execution_log FROM videos WHERE id = ?", id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appmw.NewAppError("Video not found", 404, "VIDEO_NOT_FOUND")
		}
		return fmt.Errorf("read execution log: %w", err)
	}

	var logEntries []models.N8NExecutionEntry
	if raw.Valid && strings.TrimSpace(raw.String) != "" {
		if err := json.Unmarshal([]byte(raw.String), &logEntries); err != nil {
			logEntries = []models.N8NExecutionEntry{}
		}
	}

	logEntries = append(logEntries, entries...)
	encoded, err := json.Marshal(logEntries)
	if err != nil {
		return fmt.Errorf("encode execution log: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "UPDATE videos SET n8n_execution_log = ?, updated_at = ? WHERE id = ?", string(encoded), time.Now().UTC().Format(time.RFC3339), id); err != nil {
		return fmt.Errorf("write execution log: %w", err)
	}
	return nil
}

func buildVideoWhere(filters VideoFilters) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	if filters.Status != "" && filters.Status != "all" {
		conditions = append(conditions, "v.status = ?")
		args = append(args, filters.Status)
	}
	if filters.Platform != "" && filters.Platform != "all" {
		conditions = append(conditions, "v.platform = ?")
		args = append(args, filters.Platform)
	}
	if filters.Search != "" {
		conditions = append(conditions, "(v.title LIKE ? OR v.description LIKE ?)")
		like := "%" + filters.Search + "%"
		args = append(args, like, like)
	}
	if filters.DateFrom != "" {
		conditions = append(conditions, "v.scheduled_at >= ?")
		args = append(args, filters.DateFrom)
	}
	if filters.DateTo != "" {
		conditions = append(conditions, "v.scheduled_at <= ?")
		args = append(args, filters.DateTo)
	}

	tags := NormalizeTagNames(strings.Split(filters.Tag, ","))
	for _, tag := range tags {
		conditions = append(conditions, `EXISTS (
SELECT 1
FROM video_tags ft
JOIN tags tt ON ft.tag_id = tt.id
WHERE ft.video_id = v.id AND tt.name = ? COLLATE NOCASE
)`)
		args = append(args, tag)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func scanVideos(rows *sql.Rows) ([]models.Video, error) {
	videos := make([]models.Video, 0)
	for rows.Next() {
		video, err := scanVideo(rows)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate videos: %w", err)
	}
	return videos, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanVideo(row scanner) (models.Video, error) {
	var video models.Video
	var description sql.NullString
	var duration sql.NullInt64
	var thumbnail sql.NullString
	var scheduledAt sql.NullString
	var postedAt sql.NullString
	var workflowID sql.NullString
	var executionLog sql.NullString
	var createdAt sql.NullString
	var updatedAt sql.NullString

	if err := row.Scan(
		&video.ID,
		&video.Title,
		&description,
		&video.Filename,
		&video.OriginalName,
		&video.FileSize,
		&duration,
		&thumbnail,
		&video.Platform,
		&video.Status,
		&scheduledAt,
		&postedAt,
		&workflowID,
		&executionLog,
		&createdAt,
		&updatedAt,
	); err != nil {
		return models.Video{}, fmt.Errorf("scan video: %w", err)
	}

	video.Description = stringPtr(description)
	video.Duration = int64Ptr(duration)
	video.Thumbnail = stringPtr(thumbnail)
	video.ScheduledAt = stringPtr(scheduledAt)
	video.PostedAt = stringPtr(postedAt)
	video.N8NWorkflowID = stringPtr(workflowID)
	if executionLog.Valid {
		video.N8NExecutionLog = executionLog.String
	} else {
		video.N8NExecutionLog = "[]"
	}
	if createdAt.Valid {
		video.CreatedAt = createdAt.String
	}
	if updatedAt.Valid {
		video.UpdatedAt = updatedAt.String
	}
	video.Tags = []string{}

	return video, nil
}

func ensurePlatform(platform string) error {
	switch platform {
	case "instagram", "tiktok", "youtube", "all":
		return nil
	default:
		return appmw.NewAppError("Invalid platform", 422, "INVALID_PLATFORM")
	}
}

func ensureStatus(status string) error {
	switch status {
	case "scheduled", "posted", "draft", "partial":
		return nil
	default:
		return appmw.NewAppError("Invalid status", 422, "INVALID_STATUS")
	}
}

func sanitizeSort(sort string) string {
	switch sort {
	case "created_at", "scheduled_at", "posted_at", "title":
		return sort
	default:
		return "created_at"
	}
}

func sanitizeOrder(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}
	return "DESC"
}

func nullableString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func rollbackQuietly(tx *sql.Tx) {
	_ = tx.Rollback()
}

func removeQuietly(path string) {
	if path == "" {
		return
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "remove %s failed: %v\n", path, err)
	}
}

func (s *VideoService) resolveStoredPath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(s.projectRoot, path))
}
