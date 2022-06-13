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
	"net/url"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/notification"
	"github.com/danifv27/soup/internal/domain/soup"
	"github.com/danifv27/soup/internal/infrastructure/audit/clover"
	"github.com/danifv27/soup/internal/infrastructure/deployment"
	"github.com/danifv27/soup/internal/infrastructure/git"
	"github.com/danifv27/soup/internal/infrastructure/logger/logrus"
	"github.com/danifv27/soup/internal/infrastructure/notification/console"
	"github.com/danifv27/soup/internal/infrastructure/notification/opsgenie"
	"github.com/danifv27/soup/internal/infrastructure/status"
	"github.com/danifv27/soup/internal/infrastructure/storage/config"
	"github.com/danifv27/soup/internal/infrastructure/storage/embed"
)

//Adapters contains the exposed adapters of interface adapters
type Adapters struct {
	LoggerService       logger.Logger
	NotificationService notification.Notifier
	AuditService        audit.Auditer
	VersionRepository   soup.Version
	GitRepository       soup.Git
	DeployRepository    soup.Deploy
	SoupRepository      soup.Config
	ProbeRepository     soup.Probe
}

func getSchemaFromURI(uri string) (string, error) {
func getOpaqueFromURI(uri string) (string, error) {

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if u.Scheme == "audit" || u.Scheme == "notifier" || u.Scheme == "informer" {
		return u.Opaque, nil
	}

	return "", fmt.Errorf("getOpaqueFromURI: '%s' unsupported schema", u.Scheme)
}

func NewAdapters(auditerURI string, notifierURI string) (Adapters, error) {
	var err error
	var opaque string
	var n notification.Notifier
	var a audit.Auditer

	l := logrus.NewLoggerService()
	if a, err = clover.NewCloverAuditer(auditerURI); err != nil {
		return Adapters{}, err
	}
	r := git.NewGitRepo(l, a)
	c := config.NewSoupConfig(".")
	d := deployment.NewDeployHandler(l)

	if opaque, err = getOpaqueFromURI(notifierURI); err != nil {
		return Adapters{}, err
	}
	switch {
	case opaque == "opsgenie":
		if n, err = opsgenie.NewOpsgenieService(notifierURI, l); err != nil {
			return Adapters{}, err
		}
	case opaque == "noop":
		n = notifNoop.NewNotifier()
	case opaque == "console":
		n = console.NewNotificationService()
	default:
		return Adapters{}, fmt.Errorf("NewAdapters: unsopported notifier schema %s", opaque)
	}
	}

	return Adapters{
		LoggerService:       l,
		NotificationService: n,
		AuditService:        a,
		VersionRepository:   embed.NewVersionRepo(),
		GitRepository:       &r,
		DeployRepository:    &d,
		SoupRepository:      c,
		ProbeRepository:     status.NewProbeRepo(&r, &d),
	}, nil
}
