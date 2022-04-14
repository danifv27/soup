package probes

import (
	"encoding/json"
	"net/http"

	"github.com/danifv27/soup/internal/application"
)

type Handler struct {
	apps application.Applications
}

//NewHandler Constructor
func NewHandler(app application.Applications) *Handler {

	return &Handler{apps: app}
}

//GetLiveness Returns liveness status
func (c Handler) GetLiveness(w http.ResponseWriter, _ *http.Request) {

	enc := json.NewEncoder(w)
	info, err := c.apps.Queries.GetLivenessInfoHandler.Handle()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	enc.Encode(info)
}

//GetReadiness Returns liveness status
func (c Handler) GetReadiness(w http.ResponseWriter, _ *http.Request) {

	enc := json.NewEncoder(w)
	info, err := c.apps.Queries.GetReadinessInfoHandler.Handle()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	enc.Encode(info)
}
