package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/your-org/reelqueue-server/config"
	appmw "github.com/your-org/reelqueue-server/middleware"
	"github.com/your-org/reelqueue-server/services"
)

// VideoHandler exposes video HTTP endpoints.
type VideoHandler struct {
	cfg        config.Config
	videos     *services.VideoService
	thumbnails *services.ThumbnailService
}

// NewVideoHandler creates a video handler.
func NewVideoHandler(cfg config.Config, videos *services.VideoService, thumbnails *services.ThumbnailService) *VideoHandler {
	return &VideoHandler{cfg: cfg, videos: videos, thumbnails: thumbnails}
}

// Routes registers video routes on the provided router.
func (h *VideoHandler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/upload", h.Upload)
	r.Post("/bulk", h.Bulk)
	r.Get("/{id}", h.Get)
	r.Patch("/{id}", h.Update)
	r.Patch("/{id}/status", h.UpdateStatus)
	r.Delete("/{id}", h.Delete)
	r.Get("/{id}/thumbnail", h.GetThumbnail)
	r.Get("/{id}/file", h.GetFile)
}

// List handles GET /api/videos.
func (h *VideoHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	videos, total, err := h.videos.List(r.Context(), services.VideoFilters{
		Status:   query.Get("status"),
		Platform: query.Get("platform"),
		Tag:      query.Get("tag"),
		Search:   query.Get("search"),
		Sort:     query.Get("sort"),
		Order:    query.Get("order"),
		DateFrom: query.Get("dateFrom"),
		DateTo:   query.Get("dateTo"),
	})
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, videos, appmw.Meta{Total: total})
}

// Get handles GET /api/videos/{id}.
func (h *VideoHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	video, err := h.videos.GetByID(r.Context(), id)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, video, map[string]any{})
}

// Upload handles POST /api/videos/upload.
func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	maxBytes := h.cfg.MaxFileSizeMB << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+1024*1024)

	if err := r.ParseMultipartForm(maxBytes); err != nil {
		appmw.RespondError(w, http.StatusRequestEntityTooLarge, "Video file is too large", "FILE_TOO_LARGE")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		appmw.RespondError(w, http.StatusUnprocessableEntity, "Video file is required", "FILE_REQUIRED")
		return
	}
	defer file.Close()

	if header.Size > maxBytes {
		appmw.RespondError(w, http.StatusRequestEntityTooLarge, "Video file is too large", "FILE_TOO_LARGE")
		return
	}

	buffer := make([]byte, 512)
	read, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		appmw.RespondError(w, http.StatusBadRequest, "Could not read uploaded file", "INVALID_UPLOAD")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		appmw.RespondError(w, http.StatusBadRequest, "Could not rewind uploaded file", "INVALID_UPLOAD")
		return
	}
	if !strings.HasPrefix(http.DetectContentType(buffer[:read]), "video/") {
		appmw.RespondError(w, http.StatusUnprocessableEntity, "Only video files are allowed", "INVALID_FILE_TYPE")
		return
	}

	if err := os.MkdirAll(h.cfg.UploadDir, 0o755); err != nil {
		appmw.HandleError(w, fmt.Errorf("create upload directory: %w", err))
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	filename := uuid.New().String() + ext
	target := filepath.Join(h.cfg.UploadDir, filename)
	dst, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		appmw.HandleError(w, fmt.Errorf("create upload file: %w", err))
		return
	}

	written, copyErr := io.Copy(dst, file)
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(target)
		appmw.HandleError(w, fmt.Errorf("save uploaded file: %w", copyErr))
		return
	}
	if closeErr != nil {
		_ = os.Remove(target)
		appmw.HandleError(w, fmt.Errorf("close uploaded file: %w", closeErr))
		return
	}
	if written > maxBytes {
		_ = os.Remove(target)
		appmw.RespondError(w, http.StatusRequestEntityTooLarge, "Video file is too large", "FILE_TOO_LARGE")
		return
	}

	video, err := h.videos.CreateFromUpload(r.Context(), services.UploadedVideoInput{
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		Filename:     filename,
		OriginalName: header.Filename,
		FileSize:     written,
		Platform:     r.FormValue("platform"),
		ScheduledAt:  r.FormValue("scheduled_at"),
		TagsRaw:      r.FormValue("tags"),
	})
	if err != nil {
		_ = os.Remove(target)
		appmw.HandleError(w, err)
		return
	}

	basename := strings.TrimSuffix(filename, filepath.Ext(filename))
	h.thumbnails.ProcessVideoAsync(video.ID, target, basename)
	appmw.Respond(w, http.StatusCreated, video, appmw.Meta{Total: 1})
}

// Update handles PATCH /api/videos/{id}.
func (h *VideoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	update, err := decodeVideoUpdate(r)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}

	video, err := h.videos.Update(r.Context(), id, update)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, video, map[string]any{})
}

// UpdateStatus handles PATCH /api/videos/{id}/status.
func (h *VideoHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		appmw.RespondError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
		return
	}

	video, err := h.videos.UpdateStatus(r.Context(), id, payload.Status)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, video, map[string]any{})
}

// Bulk handles POST /api/videos/bulk.
func (h *VideoHandler) Bulk(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		IDs         []int64 `json:"ids"`
		Action      string  `json:"action"`
		ScheduledAt string  `json:"scheduled_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		appmw.RespondError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
		return
	}

	result, err := h.videos.BulkAction(r.Context(), payload.IDs, payload.Action, payload.ScheduledAt)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, result, map[string]any{})
}

// Delete handles DELETE /api/videos/{id}.
func (h *VideoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	result, err := h.videos.DeleteVideo(r.Context(), id)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, result, map[string]any{})
}

// GetThumbnail serves the stored JPEG thumbnail.
func (h *VideoHandler) GetThumbnail(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	video, err := h.videos.GetByID(r.Context(), id)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	if video.Thumbnail == nil {
		appmw.RespondError(w, http.StatusNotFound, "Thumbnail not found", "THUMBNAIL_NOT_FOUND")
		return
	}

	path := storedPath(h.cfg.ProjectRoot, *video.Thumbnail)
	if !withinDir(h.cfg.ThumbnailDir, path) {
		appmw.RespondError(w, http.StatusNotFound, "Thumbnail not found", "THUMBNAIL_NOT_FOUND")
		return
	}
	serveContentFile(w, r, path, "THUMBNAIL_NOT_FOUND")
}

// GetFile streams the stored source video with range request support.
func (h *VideoHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	video, err := h.videos.GetByID(r.Context(), id)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}

	path := filepath.Join(h.cfg.UploadDir, filepath.Base(video.Filename))
	serveContentFile(w, r, path, "VIDEO_FILE_NOT_FOUND")
}

func parseIDParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		appmw.RespondError(w, http.StatusBadRequest, "Invalid video id", "INVALID_VIDEO_ID")
		return 0, false
	}
	return id, true
}

func decodeVideoUpdate(r *http.Request) (services.VideoUpdate, error) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return services.VideoUpdate{}, appmw.NewAppError("Invalid JSON body", http.StatusBadRequest, "BAD_REQUEST")
	}

	var update services.VideoUpdate
	if value, ok := raw["title"]; ok {
		parsed, err := rawString(value)
		if err != nil {
			return services.VideoUpdate{}, err
		}
		update.Title = &parsed
	}
	if value, ok := raw["description"]; ok {
		parsed, err := rawString(value)
		if err != nil {
			return services.VideoUpdate{}, err
		}
		update.Description = &parsed
	}
	if value, ok := raw["platform"]; ok {
		parsed, err := rawString(value)
		if err != nil {
			return services.VideoUpdate{}, err
		}
		update.Platform = &parsed
	}
	if value, ok := raw["scheduled_at"]; ok {
		parsed, err := rawString(value)
		if err != nil {
			return services.VideoUpdate{}, err
		}
		update.ScheduledAt = &parsed
	}
	if value, ok := raw["n8n_workflow_id"]; ok {
		parsed, err := rawString(value)
		if err != nil {
			return services.VideoUpdate{}, err
		}
		update.N8NWorkflowID = &parsed
	}
	if value, ok := raw["tags"]; ok {
		tags, err := rawTags(value)
		if err != nil {
			return services.VideoUpdate{}, err
		}
		update.Tags = &tags
	}

	return update, nil
}

func rawString(raw json.RawMessage) (string, error) {
	if string(raw) == "null" {
		return "", nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", appmw.NewAppError("Invalid string field", http.StatusBadRequest, "BAD_REQUEST")
	}
	return value, nil
}

func rawTags(raw json.RawMessage) ([]string, error) {
	if string(raw) == "null" {
		return []string{}, nil
	}
	var tags []string
	if err := json.Unmarshal(raw, &tags); err == nil {
		return services.NormalizeTagNames(tags), nil
	}
	var tagString string
	if err := json.Unmarshal(raw, &tagString); err == nil {
		return services.ParseTags(tagString), nil
	}
	return nil, appmw.NewAppError("Invalid tags field", http.StatusBadRequest, "BAD_REQUEST")
}

func storedPath(projectRoot, value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(projectRoot, value))
}

func withinDir(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func serveContentFile(w http.ResponseWriter, r *http.Request, path string, missingCode string) {
	file, err := os.Open(path)
	if err != nil {
		appmw.RespondError(w, http.StatusNotFound, missingMessage(missingCode), missingCode)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		appmw.RespondError(w, http.StatusNotFound, missingMessage(missingCode), missingCode)
		return
	}

	// http.ServeContent preserves range requests, which browser video seeking needs.
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func missingMessage(code string) string {
	switch code {
	case "THUMBNAIL_NOT_FOUND":
		return "Thumbnail not found"
	case "VIDEO_FILE_NOT_FOUND":
		return "Video file not found"
	default:
		return "File not found"
	}
}
