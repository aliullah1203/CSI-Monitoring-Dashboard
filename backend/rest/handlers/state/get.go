package state

import (
	"net/http"

	"backend/util"
)

// GetAll returns summary + pending + exceptions in one DB round-trip
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	sourceID := r.URL.Query().Get("source_id")
	summary, err := h.stateRepo.GetSummary(sourceID)
	if err != nil {
		util.SendError(w, http.StatusInternalServerError, "failed to get summary")
		return
	}
	pending, err := h.stateRepo.GetPending(sourceID)
	if err != nil {
		util.SendError(w, http.StatusInternalServerError, "failed to get pending")
		return
	}
	exceptions, err := h.stateRepo.GetExceptions(sourceID)
	if err != nil {
		util.SendError(w, http.StatusInternalServerError, "failed to get exceptions")
		return
	}
	util.SendData(w, http.StatusOK, map[string]interface{}{
		"summary":    summary,
		"pending":    pending,
		"exceptions": exceptions,
	})
}

func (h *Handler) GetState(w http.ResponseWriter, r *http.Request) {
	sourceID := r.URL.Query().Get("source_id")
	view := r.URL.Query().Get("view")
	if view == "" {
		view = "summary"
	}

	switch view {
	case "summary":
		summary, err := h.stateRepo.GetSummary(sourceID)
		if err != nil {
			util.SendError(w, http.StatusInternalServerError, "failed to get summary")
			return
		}
		util.SendData(w, http.StatusOK, summary)

	case "pending":
		events, err := h.stateRepo.GetPending(sourceID)
		if err != nil {
			util.SendError(w, http.StatusInternalServerError, "failed to get pending events")
			return
		}
		util.SendData(w, http.StatusOK, map[string]interface{}{"events": events})

	case "exceptions":
		events, err := h.stateRepo.GetExceptions(sourceID)
		if err != nil {
			util.SendError(w, http.StatusInternalServerError, "failed to get exceptions")
			return
		}
		util.SendData(w, http.StatusOK, map[string]interface{}{"events": events})

	default:
		util.SendError(w, http.StatusBadRequest, "view must be summary, pending, or exceptions")
	}
}
