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

// func (svc *NotificationService) Init(url string, token string) error {

// 	return nil
// }

// Notify prints out the notifications in console
func (svc *NotificationService) Notify(notification notification.Notification) error {

	fmt.Println(notification.Message)
	return nil
}
