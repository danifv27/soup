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
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/soup/queries"
	"github.com/danifv27/soup/internal/domain/soup"
)

//Queries operations that request data
type Queries struct {
	GetVersionInfoHandler queries.GetVersionInfoHandler
}

//Commands operations that accept data to make a change or trigger an action
type Commands struct {
}

//Applications contains all exposed services of the application layer
type Applications struct {
	LoggerService logger.Logger
	Queries       Queries
	Commands      Commands
}

// NewApplications Bootstraps Application Layer dependencies
func NewApplications(logger logger.Logger, versionRepo soup.VersionRepository) Applications {

	return Applications{
		LoggerService: logger,
		Queries: Queries{
			GetVersionInfoHandler: queries.NewGetVersionInfoHandler(versionRepo),
		},
		Commands: Commands{},
	}
}
