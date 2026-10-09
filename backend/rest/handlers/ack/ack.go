package ack

import (
	"encoding/json"
	"net/http"

	"backend/util"
)

type ackRequest struct {
	EventIDs []string `json:"event_ids"`
}

func (h *Handler) AckEvents(w http.ResponseWriter, r *http.Request) {
	var req ackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.EventIDs) == 0 {
		util.SendError(w, http.StatusBadRequest, "event_ids is required")
		return
	}

	results, err := h.ackRepo.AckEvents(req.EventIDs)
	if err != nil {
		util.SendError(w, http.StatusInternalServerError, "failed to acknowledge events")
		return
	}

	util.SendData(w, http.StatusOK, map[string]interface{}{"results": results})
}
