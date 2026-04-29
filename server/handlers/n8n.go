package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appmw "github.com/your-org/reelqueue-server/middleware"
	"github.com/your-org/reelqueue-server/services"
)

// N8NHandler exposes secret-protected n8n endpoints.
type N8NHandler struct {
	videos *services.VideoService
}

// NewN8NHandler creates an n8n handler.
func NewN8NHandler(videos *services.VideoService) *N8NHandler {
	return &N8NHandler{videos: videos}
}

// Routes registers n8n routes.
func (h *N8NHandler) Routes(r chi.Router) {
	r.Get("/queue", h.Queue)
	r.Post("/webhook/posted", h.Posted)
	r.Post("/webhook/failed", h.Failed)
}

// Queue handles GET /api/n8n/queue.
func (h *N8NHandler) Queue(w http.ResponseWriter, r *http.Request) {
	videos, err := h.videos.GetDueScheduled(r.Context())
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, videos, appmw.Meta{Total: len(videos)})
}

// Posted handles POST /api/n8n/webhook/posted.
func (h *N8NHandler) Posted(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		VideoID     int64   `json:"video_id"`
		PostedAt    *string `json:"posted_at"`
		Platform    *string `json:"platform"`
		ExecutionID *string `json:"execution_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		appmw.RespondError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
		return
	}

	video, err := h.videos.MarkPostedFromN8N(r.Context(), services.N8NPostedInput{
		VideoID:     payload.VideoID,
		PostedAt:    payload.PostedAt,
		Platform:    payload.Platform,
		ExecutionID: payload.ExecutionID,
	})
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, video, map[string]any{})
}

// Failed handles POST /api/n8n/webhook/failed.
func (h *N8NHandler) Failed(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		VideoID     int64   `json:"video_id"`
		Error       string  `json:"error"`
		ExecutionID *string `json:"execution_id"`
		Platform    *string `json:"platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		appmw.RespondError(w, http.StatusBadRequest, "Invalid JSON body", "BAD_REQUEST")
		return
	}

	video, err := h.videos.MarkFailedFromN8N(r.Context(), services.N8NFailedInput{
		VideoID:     payload.VideoID,
		Error:       payload.Error,
		ExecutionID: payload.ExecutionID,
		Platform:    payload.Platform,
	})
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, video, map[string]any{})
}
