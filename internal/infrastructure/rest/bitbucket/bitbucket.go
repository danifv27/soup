package bitbucket

import (
	"net/http"

	bitbucketserver "github.com/go-playground/webhooks/v6/bitbucket-server"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/logger"
)

type Handler struct {
	apps application.Applications
	hook *bitbucketserver.Webhook
}

//NewHandler Constructor
func NewHandler(app application.Applications) *Handler {
	var hook *bitbucketserver.Webhook
	var err error

	if hook, err = bitbucketserver.New(bitbucketserver.Options.Secret("elsecretomasoscuro")); err != nil {
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

	if payload, err = c.hook.Parse(r, bitbucketserver.RepositoryReferenceChangedEvent, bitbucketserver.PullRequestOpenedEvent); err != nil {
		if err == bitbucketserver.ErrEventNotFound {
			// ok event wasn;t one of the ones asked to be parsed
			c.apps.LoggerService.WithFields(logger.Fields{
				"payload": payload,
			}).Warn("Bitbucket event not found")
			return
		}
	}

	c.apps.LoggerService.WithFields(logger.Fields{
		"payload": payload,
	}).Debug("Bitbucket event received")

	// switch payload.(type) {
	// case bitbucketserver.RepositoryModifiedEvent:
	// 	push := payload.(bitbucketserver.RepoPushPayload)
	// 	c.apps.LoggerService.WithFields(logger.Fields{
	// 		"payload": push,
	// 	}).Info("WebhookEvent")
	// }
}
