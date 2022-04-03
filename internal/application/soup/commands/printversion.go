package commands

import (
	"encoding/json"
	"fmt"

	"github.com/danifv27/soup/internal/application/notification"
	"github.com/danifv27/soup/internal/domain/soup"
)

type PrintVersionRequest struct {
	Format string
}

type PrintVersionRequestHandler interface {
	Handle(command PrintVersionRequest) error
}

type printVersionRequestHandler struct {
	repo     soup.Version
	notifier notification.Notifier
}

//NewUpdateCragRequestHandler Constructor
func NewPrintVersionRequestHandler(version soup.Version, notifier notification.Notifier) PrintVersionRequestHandler {

	return printVersionRequestHandler{
		repo:     version,
		notifier: notifier,
	}
}

//Handle Handles the update request
func (h printVersionRequestHandler) Handle(command PrintVersionRequest) error {
	var info *soup.VersionInfo
	var err error
	var out []byte
	var notif notification.Notification

	if info, err = h.repo.GetVersionInfo(); err != nil {
		return err
	}
	if command.Format == "json" {
		if out, err = json.MarshalIndent(info, "", "    "); err != nil {
			return err
		}
		notif = notification.Notification{
			Message: string(out),
		}
		// fmt.Println(string(out))
	} else {
		// fmt.Println(info)
		notif = notification.Notification{
			Message: fmt.Sprint(info),
		}
	}
	h.notifier.Notify(notif)

	return nil
}
