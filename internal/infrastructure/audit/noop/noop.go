package noop

import (
	"github.com/danifv27/soup/internal/application/audit"
)

// NoopAuditer
type NoopAuditer struct{}

// NewAuditer
func NewAuditer() *NoopAuditer {

	return &NoopAuditer{}
}

func (n *NoopAuditer) Audit(event *audit.Event) error {

	return nil
}

func (n *NoopAuditer) GetEvents(option *audit.GetEventOption) ([]audit.Event, error) {

	events := make([]audit.Event, 0, 0)

	return events, nil
}

func (n *NoopAuditer) GetNumberOfEvents(option *audit.GetEventOption) (int, error) {

	return 0, nil
}
