package executor

import (
	"sync"

	"github.com/danifv27/soup/internal/application"
	señales "github.com/danifv27/soup/internal/application/signals"
)

//Loop Represents the signaled command execution
type Loop struct {
	Apps       application.Applications
	SigHandler señales.SignalHandler
}

//NewLoop version command main loop
func NewLoop(apps application.Applications, handler señales.SignalHandler) Loop {

	loop := Loop{
		Apps:       apps,
		SigHandler: handler,
	}

	return loop
}

func (l *Loop) Exec() {

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		l.SigHandler.Run()
		wg.Done()
	}()

	wg.Wait()
}
