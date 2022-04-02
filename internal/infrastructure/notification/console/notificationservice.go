package console

import (
	"fmt"

	"github.com/danifv27/soup/internal/application/notification"
)

// NotificationService provides a console implementation of the Service
type NotificationService struct{}

// NewNotificationService constructor for NotificationService
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// Notify prints out the notifications in console
func (NotificationService) Notify(notification notification.Notification) error {

	fmt.Printf(notification.Message)
	return nil
}
