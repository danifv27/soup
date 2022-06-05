package notification

// Notification provides a struct to send messages via the Service
type Notification struct {
	//Alert text
	Message string `json:"message"`
	//Alert text in long form
	Description string `json:"description"`
	//List of labels attached to the alert.
	Tags []string `json:"tags"`
	//Priority of alert
	Priority string `json:"priority"`
	//list of teams for setting responders
	Teams []string `json:"teams"`
}

// Service sends Notification
type Notifier interface {
	Notify(notification Notification) error
}
