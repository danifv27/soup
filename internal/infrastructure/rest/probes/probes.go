package probes

import (
	"encoding/json"
	"fmt"
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
	info, err := c.apps.Queries.GetLivenessInfoHandler.Handle()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(info)
	if err != nil {
		return
	}
}
