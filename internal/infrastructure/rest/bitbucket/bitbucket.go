package bitbucket

import (
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

//WebhookEvent Returns liveness status
func (c Handler) WebhookEvent(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(http.StatusOK)
	c.apps.LoggerService.Info("WebhookEvent")
}
