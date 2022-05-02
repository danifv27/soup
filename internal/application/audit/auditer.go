package audit

import "time"

type Event struct {
	// Namespace string     `json:"namespace"`
	// TargetID  string     `json:"target_id"`
	Action  string `json:"action"`
	Actor   string `json:"actor"`
	Message string `json:"message"`
	// State     int        `json:"state"`
	// ClientIP  string     `json:"client_ip"`
	CreatedAt *time.Time `json:"created_at"`
}

type ReadLogOption struct {
	// Namespace string
	// TargetID  string
	// Action    string
	// Actor     string
	// State     int
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}

type Auditer interface {
	Log(event *Event) error
	ReadLog(option *ReadLogOption) ([]Event, error)
	TotalCount(option *ReadLogOption) (int, error)
}
