package mqttstatus

import (
	"backend/mqtt"
	"backend/util"
	"net/http"
)

func GetStatus(w http.ResponseWriter, r *http.Request) {
	util.SendData(w, http.StatusOK, mqtt.GetState())
}
