// The application layer exposes all supported use cases of the application to the outside world.
// It consists of:

// 	Business logic/Use Cases
// 		Implementation of business requirements
// 		We can implement this with command/query separation. We cover this in our sample application.
// 	Application Services
// 		They provide isolated business logic/use cases functionality that is required. This functionality, is expressed by use cases.
// 		It can be an interface-only service if it is infrastructure-dependent

// The application layer code depends only on the domain layer.
package application

import (
	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/notification"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/application/soup/queries"
	"github.com/danifv27/soup/internal/domain/soup"
)

//Queries operations that request data
type Queries struct {
	GetVersionInfoHandler   queries.GetVersionInfoHandler
	GetLivenessInfoHandler  queries.GetLivenessInfoHandler
	GetReadinessInfoHandler queries.GetReadinessInfoHandler
}

//Commands operations that accept data to make a change or trigger an action
type Commands struct {
	PrintVersion  commands.PrintVersionRequestHandler
	LoopBranches  commands.LoopBranchesRequestHandler
	ProcessBranch commands.ProcessBranchRequestHandler
}

//Applications contains all exposed services of the application layer
type Applications struct {
	LoggerService logger.Logger
	Notifier      notification.Notifier
	Auditer       audit.Auditer
	Queries       Queries
	Commands      Commands
}

// NewApplications Bootstraps Application Layer dependencies
func NewApplications(logger logger.Logger,
	notifier notification.Notifier,
	auditer audit.Auditer,
	version soup.Version,
	git soup.Git,
	deploy soup.Deploy,
	config soup.Config,
	probes soup.Probe) Applications {

	return Applications{
		LoggerService: logger,
		Notifier:      notifier,
		Auditer:       auditer,
		Queries: Queries{
			GetVersionInfoHandler:   queries.NewGetVersionInfoHandler(version),
			GetLivenessInfoHandler:  queries.NewGetLivenessInfoHandler(probes),
			GetReadinessInfoHandler: queries.NewGetReadinessInfoHandler(probes, git),
		},
		Commands: Commands{
			PrintVersion:  commands.NewPrintVersionRequestHandler(version, notifier),
			LoopBranches:  commands.NewLoopBranchesRequestHandler(git, deploy, config, logger),
			ProcessBranch: commands.NewProcessBranchRequestHandler(git, deploy, config, logger),
		},
	}
}
