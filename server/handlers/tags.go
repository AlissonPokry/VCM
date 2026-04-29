package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	appmw "github.com/your-org/reelqueue-server/middleware"
	"github.com/your-org/reelqueue-server/services"
)

// TagHandler exposes tag HTTP endpoints.
type TagHandler struct {
	tags *services.TagService
}

// NewTagHandler creates a tag handler.
func NewTagHandler(tags *services.TagService) *TagHandler {
	return &TagHandler{tags: tags}
}

// Routes registers tag routes.
func (h *TagHandler) Routes(r chi.Router) {
	r.Get("/", h.List)
}

// List handles GET /api/tags.
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.tags.ListTags(r.Context())
	if err != nil {
		appmw.HandleError(w, err)
		return
	}
	appmw.Respond(w, http.StatusOK, tags, appmw.Meta{Total: len(tags)})
}
