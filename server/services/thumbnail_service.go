package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// ThumbnailService extracts thumbnails and probes duration with FFmpeg.
type ThumbnailService struct {
	db           *sql.DB
	thumbnailDir string
	projectRoot  string
	enabled      bool
}

// NewThumbnailService creates a thumbnail service.
func NewThumbnailService(db *sql.DB, thumbnailDir, projectRoot string, enabled bool) *ThumbnailService {
	return &ThumbnailService{db: db, thumbnailDir: thumbnailDir, projectRoot: projectRoot, enabled: enabled}
}

// ProcessVideoAsync runs video processing in a recover-protected goroutine.
func (s *ThumbnailService) ProcessVideoAsync(videoID int64, inputPath, outputBasename string) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("thumbnail panic recovered: %v", recovered)
				s.updateVideoMetadata(videoID, nil, nil)
			}
		}()
		s.ProcessVideo(videoID, inputPath, outputBasename)
	}()
}

// ProcessVideo extracts a thumbnail and stores duration metadata.
func (s *ThumbnailService) ProcessVideo(videoID int64, inputPath, outputBasename string) {
	if !s.enabled {
		log.Printf("thumbnail processing disabled: ffmpeg not found")
		s.updateVideoMetadata(videoID, nil, nil)
		return
	}

	if err := os.MkdirAll(s.thumbnailDir, 0o755); err != nil {
		log.Printf("create thumbnail directory failed: %v", err)
		s.updateVideoMetadata(videoID, nil, nil)
		return
	}

	duration, err := probeDuration(inputPath)
	if err != nil {
		log.Printf("probe duration failed: %v", err)
		s.updateVideoMetadata(videoID, nil, nil)
		return
	}

	outputPath := filepath.Join(s.thumbnailDir, outputBasename+".jpg")
	if err := extractFrame(inputPath, outputPath, duration); err != nil {
		log.Printf("extract thumbnail failed: %v", err)
		s.updateVideoMetadata(videoID, nil, duration)
		return
	}

	relative, err := filepath.Rel(s.projectRoot, outputPath)
	if err != nil {
		log.Printf("resolve thumbnail relative path failed: %v", err)
		s.updateVideoMetadata(videoID, nil, duration)
		return
	}

	thumbnail := filepath.ToSlash(relative)
	s.updateVideoMetadata(videoID, &thumbnail, duration)
}

func probeDuration(inputPath string) (*int64, error) {
	raw, err := ffmpeg.Probe(inputPath)
	if err != nil {
		return nil, fmt.Errorf("ffprobe %s: %w", inputPath, err)
	}

	var payload struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}
	if payload.Format.Duration == "" {
		return nil, nil
	}

	seconds, err := strconv.ParseFloat(payload.Format.Duration, 64)
	if err != nil {
		return nil, fmt.Errorf("parse duration: %w", err)
	}

	rounded := int64(math.Round(seconds))
	return &rounded, nil
}

func extractFrame(inputPath, outputPath string, duration *int64) error {
	seek := "1"
	if duration != nil && *duration <= 1 {
		seek = "0"
	}

	err := ffmpeg.Input(inputPath, ffmpeg.KwArgs{"ss": seek}).
		Output(outputPath, ffmpeg.KwArgs{
			"frames:v": 1,
			"vf":       "scale=640:360:force_original_aspect_ratio=decrease,pad=640:360:(ow-iw)/2:(oh-ih)/2",
			"q:v":      2,
		}).
		OverWriteOutput().
		Run()
	if err != nil {
		return fmt.Errorf("run ffmpeg frame extraction: %w", err)
	}
	return nil
}

func (s *ThumbnailService) updateVideoMetadata(videoID int64, thumbnail *string, duration *int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := s.db.ExecContext(
		ctx,
		"UPDATE videos SET thumbnail = ?, duration = ?, updated_at = datetime('now') WHERE id = ?",
		thumbnail,
		duration,
		videoID,
	); err != nil {
		log.Printf("update thumbnail metadata failed: %v", err)
	}
}
