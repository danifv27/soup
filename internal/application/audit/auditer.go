package audit

import "time"

type Event struct {
	Action    string     `json:"action"`
	Actor     string     `json:"actor"`
	Message   string     `json:"message"`
	CreatedAt *time.Time `json:"created_at"`
	// State     int        `json:"state"`
	// ClientIP  string     `json:"client_ip"`
	// Namespace string     `json:"namespace"`
	// TargetID  string     `json:"target_id"`
}

type GetEventOption struct {
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	// Namespace string
	// TargetID  string
	// Action    string
	// Actor     string
	// State     int
	// Offset    int
}

type Auditer interface {
	Audit(event *Event) error
	GetEvents(option *GetEventOption) ([]Event, error)
	GetNumberOfEvents(option *GetEventOption) (int, error)
}
