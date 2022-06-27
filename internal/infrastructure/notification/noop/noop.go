package noop

import (
	"github.com/danifv27/soup/internal/application/notification"
)

// NoopNotifier
type NoopNotifier struct{}

// NewNotifier constructor for NotificationService
func NewNotifier() *NoopNotifier {

	return &NoopNotifier{}
}

// Notify prints out the notifications in console
func (n *NoopNotifier) Notify(notification notification.Notification) error {

	return nil
}
