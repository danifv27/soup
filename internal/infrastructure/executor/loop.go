package executor

import (
	"sync"

	"github.com/danifv27/soup/internal/application"
	señales "github.com/danifv27/soup/internal/application/signals"
)

//Loop Represents the signaled command execution
type Loop struct {
	apps       application.Applications
	sigHandler señales.SignalHandler
}

//NewLoop version command main loop
func NewLoop(apps application.Applications, handler señales.SignalHandler) Loop {

	loop := Loop{
		apps:       apps,
		sigHandler: handler,
	}

	return loop
}

func (l *Loop) Exec(wg *sync.WaitGroup) {

	// wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		l.apps.LoggerService.Debug("starting executor")
		l.sigHandler.Run()
		wg.Done()
	}()

	// wg.Wait()
}
