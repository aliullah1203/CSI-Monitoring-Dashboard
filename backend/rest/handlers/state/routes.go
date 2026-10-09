package state

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/state", h.GetState)
	mux.HandleFunc("GET /api/state/all", h.GetAll)
}
