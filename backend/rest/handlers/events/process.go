package events

import (
	"encoding/json"
	"net/http"

	"backend/repo"
	"backend/util"
)

func (h *Handler) ProcessEvents(w http.ResponseWriter, r *http.Request) {
	var body json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		util.SendError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var inputs []repo.EventInput

	// Try array first
	if err := json.Unmarshal(body, &inputs); err != nil {
		// Try single object
		var single repo.EventInput
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			util.SendError(w, http.StatusBadRequest, "body must be an event object or array")
			return
		}
		inputs = []repo.EventInput{single}
	}

	results, err := h.eventsRepo.ProcessEvents(inputs, nil)
	if err != nil {
		util.SendError(w, http.StatusInternalServerError, "failed to process events")
		return
	}

	util.SendData(w, http.StatusOK, map[string]interface{}{"results": results})
}
