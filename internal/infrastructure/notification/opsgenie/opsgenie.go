package opsgenie

import (
	"context"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/notification"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

// NotificationService provides a console implementation of the Service
type OpsgenieService struct {
	logger logger.Logger
	config client.Config
}

// NewOpsgenieService constructor for NotificationService
func NewOpsgenieService(l logger.Logger) (*OpsgenieService, error) {

	svc := new(OpsgenieService)
	svc.logger = l

	return svc, nil
}

func (svc *OpsgenieService) Init(url string, token string) error {

	svc.config = client.Config{
		ApiKey:         token,
		OpsGenieAPIURL: client.ApiUrl(url),
	}

	return nil
}

func generateResponders(responderNames []string, responderType alert.ResponderType) []alert.Responder {
	var responders []alert.Responder

	if len(responderNames) == 0 {
		return nil
	}

	for _, name := range responderNames {
		responders = append(responders, alert.Responder{
			Name:     name,
			Username: name,
			Type:     responderType,
		})
	}

	return responders
}

// Notify prints out the notifications in console
func (svc *OpsgenieService) Notify(n notification.Notification) error {
	var err error
	var alertClient *alert.Client
	var resp *alert.AsyncAlertResult

	if alertClient, err = alert.NewClient(&svc.config); err != nil {
		return err
	}

	responders := generateResponders(n.Teams, alert.TeamResponder)
	// responders = append(responders, generateResponders(n.USers, alert.UserResponder)...)
	// responders = append(responders, generateResponders(n.Escalations, alert.EscalationResponder)...)
	// responders = append(responders, generateResponders(n.Schedules, alert.ScheduleResponder)...)

	req := alert.CreateAlertRequest{
		Message:     n.Message,
		Description: n.Description,
		Priority:    alert.Priority(n.Priority),
		Tags:        append([]string(nil), n.Tags...),
		Responders:  responders,
	}
	if len(n.Message) > 130 {
		req.Message = n.Message[:130]
	} else {
		req.Message = n.Message
	}
	if resp, err = alertClient.Create(context.TODO(), &req); err != nil {
		return err
	}

	svc.logger.WithFields(logger.Fields{
		"request":  req,
		"response": resp,
	}).Debug("sended notification")

	return nil
}
