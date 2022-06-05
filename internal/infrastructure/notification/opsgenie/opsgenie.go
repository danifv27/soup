package opsgenie

import (
	"context"
	"fmt"
	"net/url"

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

func parseURI(uri string) (string, string, error) {
	var host string
	var apikey string

	u, err := url.Parse(uri)
	if err != nil {
		return "", "", err
	}
	if u.Scheme != "notifier" {
		return "", "", fmt.Errorf("ParseURI: invalid scheme %s", u.Scheme)
	}

	switch u.Opaque {
	case "opsgenie":
		host = u.Query().Get("host")
		if host == "" {
			return "", "", fmt.Errorf("ParseURI: host not defined")
		}
		apikey = u.Query().Get("apikey")
		if apikey == "" {
			return "", "", fmt.Errorf("ParseURI: apikey not defined")
		}
	default:
		return "", "", fmt.Errorf("ParseURI: unsupported notifier implementation %q", u.Opaque)
	}

	return host, apikey, nil
}

// NewOpsgenieService constructor for NotificationService
func NewOpsgenieService(uri string, l logger.Logger) (*OpsgenieService, error) {
	var err error
	var host, apikey string

	if host, apikey, err = parseURI(uri); err != nil {
		return nil, fmt.Errorf("NewOpsgenieService: %w", err)
	}

	svc := new(OpsgenieService)
	svc.logger = l
	svc.config = client.Config{
		ApiKey:         apikey,
		OpsGenieAPIURL: client.ApiUrl(host),
	}

	return svc, nil
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
