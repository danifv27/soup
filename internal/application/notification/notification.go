package notification

// Notification provides a struct to send messages via the Service
type Notification struct {
	Message string `json:"message"`
}

// Service sends Notification
type Notifier interface {
	Notify(notification Notification) error
}
