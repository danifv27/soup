package noop

import (
	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/watcher"
)

type WatcherHandler struct{}

func NewWatcher() (*WatcherHandler, error) {

	return &WatcherHandler{}, nil
}

func (w WatcherHandler) AddEventHandler(handler watcher.EventHandler, auditer audit.Auditer) {}

func (w WatcherHandler) HasSynced() bool {
	return false
}
func (w WatcherHandler) Start(ch <-chan struct{}) {}
