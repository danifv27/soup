package noop

import (
	"github.com/danifv27/soup/internal/application/notification"
)

// NotificationService provides a console implementation of the Service
type NoopNotifier struct{}

// NewNotifier constructor for NotificationService
func NewNotifier() *NoopNotifier {

	return &NoopNotifier{}
}

// func (svc *NotificationService) Init(url string, token string) error {

// 	return nil
// }

// Notify prints out the notifications in console
func (svc *NoopNotifier) Notify(notification notification.Notification) error {

	return nil
}
