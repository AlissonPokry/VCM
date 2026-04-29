package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appmw "github.com/your-org/reelqueue-server/middleware"
	"github.com/your-org/reelqueue-server/services"
)

// ActivityHandler exposes activity-log endpoints.
type ActivityHandler struct {
	activity *services.ActivityService
}

// NewActivityHandler creates an activity handler.
func NewActivityHandler(activity *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{activity: activity}
}

// Routes registers activity routes.
func (h *ActivityHandler) Routes(r chi.Router) {
	r.Get("/", h.List)
}

// List handles GET /api/activity.
func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	var videoID *int64
	if raw := query.Get("video_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			appmw.RespondError(w, http.StatusBadRequest, "Invalid video id", "INVALID_VIDEO_ID")
			return
		}
		videoID = &parsed
	}

	entries, total, err := h.activity.ListActivity(r.Context(), limit, offset, videoID)
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, entries, appmw.Meta{Total: total})
}
