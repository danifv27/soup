// Package adapters provides different layers to interact with the external world.
// Typically, software applications have the following behavior:
// 		1. Receive input to initiate an operation
// 		2. Interact with infrastructure services to complete a function or produce output.
// The entry points from which we receive information (i.e., requests) are called input ports.
// The gateways through which we integrate with external services are called interface adapters.

// Input Ports and interface adapters depend on frameworks/platforms and external services that are not part of the business or domain logic.
// For this reason, they belong to this separate layer named infrastructure. This layer is sometimes referred to as Ports & Adapters.
// Interface Adapters
// 		The interface adapters are responsible for implementing domain and application services(interfaces) by integrating with specific frameworks/providers.
//		For example, we can use a SQL provider to implement a domain repository or integrate with an email/SMS provider to implement a Notification service.
// Input ports
// 		The input ports provide the entry points of the application that receive input from the outside world.
// 		For example, an input port could be an HTTP handler handling synchronous calls or a Kafka consumer handling asynchronous messages.
//
// The infrastructure layer interacts with the application layer only.
package infrastructure

import (
	"os"
	"syscall"

	"github.com/danifv27/soup/internal/application/logger"
	señales "github.com/danifv27/soup/internal/application/signals"
	"github.com/danifv27/soup/internal/domain/soup"
	"github.com/danifv27/soup/internal/infrastructure/logger/logrus"
	"github.com/danifv27/soup/internal/infrastructure/signals"
	"github.com/danifv27/soup/internal/infrastructure/storage/embed"
)

//Adapters contains the exposed adapters of interface adapters
type Adapters struct {
	LoggerService     logger.Logger
	SigHandler        señales.SignalHandler
	VersionRepository soup.VersionRepository
}

func NewAdapters() Adapters {
	l := logrus.NewLoggerService()
	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, l)
	return Adapters{
		LoggerService:     l,
		SigHandler:        &h,
		VersionRepository: embed.NewVersionRepo(),
	}
}
