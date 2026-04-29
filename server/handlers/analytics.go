package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	appmw "github.com/your-org/reelqueue-server/middleware"
	"github.com/your-org/reelqueue-server/services"
)

// AnalyticsHandler exposes analytics endpoints.
type AnalyticsHandler struct {
	analytics *services.AnalyticsService
}

// NewAnalyticsHandler creates an analytics handler.
func NewAnalyticsHandler(analytics *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

// Routes registers analytics routes.
func (h *AnalyticsHandler) Routes(r chi.Router) {
	r.Get("/summary", h.Summary)
	r.Get("/heatmap", h.Heatmap)
}

// Summary handles GET /api/analytics/summary.
func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.analytics.GetSummary(r.Context())
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, summary, map[string]any{})
}

// Heatmap handles GET /api/analytics/heatmap.
func (h *AnalyticsHandler) Heatmap(w http.ResponseWriter, r *http.Request) {
	entries, err := h.analytics.GetHeatmap(r.Context())
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, entries, appmw.Meta{Total: len(entries)})
}
