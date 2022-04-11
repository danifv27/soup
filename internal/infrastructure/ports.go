// The input ports provide the entry points of the application that receive input from the outside world.
//
// For example, an input port could be an HTTP handler handling synchronous calls or a Kafka consumer
// handling asynchronous messages.
//
package infrastructure

import (
	"github.com/danifv27/soup/internal/application"
	señales "github.com/danifv27/soup/internal/application/signals"
	"github.com/danifv27/soup/internal/infrastructure/executor"
	"github.com/danifv27/soup/internal/infrastructure/rest"
)

//Ports contains the ports services
type Ports struct {
	MainLoop  executor.Loop
	Actuators *rest.Server
}

//NewServices instantiates the services of input ports
func NewPorts(apps application.Applications, handler señales.SignalHandler) Ports {

	return Ports{
		MainLoop:  executor.NewLoop(apps, handler),
		Actuators: rest.NewServer(apps),
	}
}
