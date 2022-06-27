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
	"fmt"
	"net/url"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/notification"
	"github.com/danifv27/soup/internal/application/watcher"
	"github.com/danifv27/soup/internal/domain/soup"
	"github.com/danifv27/soup/internal/infrastructure/audit/clover"
	auditNoop "github.com/danifv27/soup/internal/infrastructure/audit/noop"
	"github.com/danifv27/soup/internal/infrastructure/deployment"
	"github.com/danifv27/soup/internal/infrastructure/logger/logrus"
	"github.com/danifv27/soup/internal/infrastructure/notification/console"
	notifNoop "github.com/danifv27/soup/internal/infrastructure/notification/noop"
	"github.com/danifv27/soup/internal/infrastructure/notification/opsgenie"
	"github.com/danifv27/soup/internal/infrastructure/status"
	"github.com/danifv27/soup/internal/infrastructure/storage/config"
	"github.com/danifv27/soup/internal/infrastructure/storage/embed"
	"github.com/danifv27/soup/internal/infrastructure/svc/git"
	svcNoop "github.com/danifv27/soup/internal/infrastructure/svc/noop"
	k8s "github.com/danifv27/soup/internal/infrastructure/watcher/kubernetes"
	"github.com/danifv27/soup/internal/infrastructure/watcher/noop"
)

//Adapters contains the exposed adapters of interface adapters
type Adapters struct {
	LoggerService       logger.Logger
	NotificationService notification.Notifier
	AuditService        audit.Auditer
	InformerService     watcher.Informer
	VersionRepository   soup.Version
	GitRepository       soup.Git
	DeployRepository    soup.Deploy
	SoupRepository      soup.Config
	ProbeRepository     soup.Probe
}

type WatcherArgs struct {
	URI        string
	Resources  []k8s.Resource
	Namespaces []string
}

type SVCArgs struct {
	URI     string
	Address string
}

func getOpaqueFromURI(uri string) (string, error) {

	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	if u.Scheme == "audit" || u.Scheme == "notifier" || u.Scheme == "informer" || u.Scheme == "svc" {
		return u.Opaque, nil
	}

	return "", fmt.Errorf("getOpaqueFromURI: '%s' unsupported schema", u.Scheme)
}

func NewAdapters(gArgs SVCArgs, auditerURI string, notifierURI string, wArgs WatcherArgs) (Adapters, error) {
	var err error
	var r soup.Git
	var opaque string
	var n notification.Notifier
	var a audit.Auditer
	var w watcher.Informer

	l := logrus.NewLoggerService()

	if opaque, err = getOpaqueFromURI(auditerURI); err != nil {
		return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
	}
	switch {
	case opaque == "clover":
		if a, err = clover.NewCloverAuditer(auditerURI); err != nil {
			return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
		}
	case opaque == "noop":
		a = auditNoop.NewAuditer()
	}
	if opaque, err = getOpaqueFromURI(gArgs.URI); err != nil {
		return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
	}
	switch {
	case opaque == "git":
		var repo git.GitRepo

		if repo, err = git.NewGitRepo(gArgs.URI, gArgs.Address, l, a); err != nil {
			return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
		}
		r = &repo
	case opaque == "noop":
		r = svcNoop.NewGit()
	default:
		return Adapters{}, fmt.Errorf("NewAdapters: unsopported svc schema %s", opaque)
	}

	c := config.NewSoupConfig(".")
	d := deployment.NewDeployHandler(l)

	if opaque, err = getOpaqueFromURI(notifierURI); err != nil {
		return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
	}
	switch {
	case opaque == "opsgenie":
		if n, err = opsgenie.NewOpsgenieService(notifierURI, l); err != nil {
			return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
		}
	case opaque == "noop":
		n = notifNoop.NewNotifier()
	case opaque == "console":
		n = console.NewNotificationService()
	default:
		return Adapters{}, fmt.Errorf("NewAdapters: unsopported notifier schema %s", opaque)
	}

	if opaque, err = getOpaqueFromURI(wArgs.URI); err != nil {
		return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
	}
	switch {
	case opaque == "k8s":
		if w, err = k8s.NewWatcher(wArgs.URI, wArgs.Resources, wArgs.Namespaces, l, a); err != nil {
			return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
		}
	case opaque == "noop":
		w, _ = noop.NewWatcher()
	default:
		return Adapters{}, fmt.Errorf("NewAdapters: unsopported informer schema %s", opaque)
	}

	return Adapters{
		LoggerService:       l,
		NotificationService: n,
		AuditService:        a,
		InformerService:     w,
		VersionRepository:   embed.NewVersionRepo(),
		GitRepository:       r,
		DeployRepository:    &d,
		SoupRepository:      c,
		ProbeRepository:     status.NewProbeRepo(r, &d),
	}, nil
}
