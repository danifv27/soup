package bitbucket

import (
	"encoding/json"
	"net/http"

	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/soup/commands"
)

type Handler struct {
	apps application.Applications
	hook *bitbucketserver.Webhook
}

//NewHandler Constructor
func NewHandler(app application.Applications, secret string) *Handler {
	var hook *bitbucketserver.Webhook
	var err error

	if hook, err = bitbucketserver.New(bitbucketserver.Options.Secret(secret)); err != nil {
		return nil
	}

	return &Handler{
		apps: app,
		hook: hook,
	}
}

//WebhookEvent Returns liveness status
func (c Handler) WebhookEvent(w http.ResponseWriter, r *http.Request) {
	var payload interface{}
	var err error

	if payload, err = c.hook.Parse(r, bitbucketserver.RepositoryReferenceChangedEvent, bitbucketserver.DiagnosticsPingEvent); err != nil {
		if err == bitbucketserver.ErrEventNotFound {
			// ok event wasn;t one of the ones asked to be parsed
			c.apps.LoggerService.WithFields(logger.Fields{
				"payload": payload,
			}).Warn("Bitbucket event not found")
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	b, _ := json.MarshalIndent(payload, "", "  ")
	c.apps.LoggerService.WithFields(logger.Fields{
		"payload": string(b),
	}).Debug("Bitbucket event received")

	switch payload.(type) {
	case bitbucketserver.DiagnosticsPingPayload:
		ping := payload.(bitbucketserver.DiagnosticsPingPayload)
		c.apps.LoggerService.WithFields(logger.Fields{
			"payload": ping,
		}).Info("DiagnosticsPingEvent")
		w.WriteHeader(http.StatusOK)
		return
	case bitbucketserver.RepositoryReferenceChangedPayload:
		refChanged := payload.(bitbucketserver.RepositoryReferenceChangedPayload)
		c.apps.LoggerService.WithFields(logger.Fields{
			"payload": refChanged,
		}).Info("RepositoryReferenceChangedEvent")
		for _, r := range refChanged.Changes {
			c.apps.LoggerService.WithFields(logger.Fields{
				"branches": r.ReferenceID,
			}).Info("modified branches")
			err = c.apps.Commands.ProcessBranch.Handle(commands.ProcessBranchRequest{Branch: r.ReferenceID})
			if err == nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	return
}
