package probes

import (
	"encoding/json"
	"net/http"

	"github.com/danifv27/soup/internal/application"
)

type ProbeHandler struct {
	apps application.Applications
}

//NewHandler implements probe interface
func NewHandler(app application.Applications) *ProbeHandler {

	return &ProbeHandler{apps: app}
}

//GetLiveness Returns liveness status
func (c ProbeHandler) GetLiveness(w http.ResponseWriter, _ *http.Request) {

	enc := json.NewEncoder(w)
	info, err := c.apps.Queries.GetLivenessInfoHandler.Handle()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.apps.LoggerService.With("err", err).Warn("liveness probe failed")
	}
	c.apps.LoggerService.Debug("liveness probe responded")
	enc.Encode(info)
}

//GetReadiness Returns readiness status
func (c ProbeHandler) GetReadiness(w http.ResponseWriter, _ *http.Request) {

	enc := json.NewEncoder(w)
	info, err := c.apps.Queries.GetReadinessInfoHandler.Handle()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.apps.LoggerService.With("err", err).Warn("readiness probe failed")
	}
	c.apps.LoggerService.Debug("readiness probe responded")
	enc.Encode(info)
}
