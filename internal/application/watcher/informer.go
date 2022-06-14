package watcher

import (
	"github.com/danifv27/soup/internal/application/audit"
	"k8s.io/client-go/tools/cache"
)

const (
	EventAdd    string = "EventAdd"
	EventUpdate string = "EventUpdate"
	EventDelete string = "EventDelete"
)

type EventHandler func(resourceType string, auditer audit.Auditer) cache.ResourceEventHandlerFuncs

type Informer interface {
	AddEventHandler(handler EventHandler, auditer audit.Auditer)
	HasSynced() bool
	Start(ch <-chan struct{})
}
