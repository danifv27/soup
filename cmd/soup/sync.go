package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/notification"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type VCS struct {
	URI string `help:"version control URI" env:"SOUP_VCS_URI" hidden:"" default:"svc:git?username=dummy&token=dummy-personal-access-token"`
	// Username  string `help:"username" env:"SOUP_VCS_USERNAME"`
	// Withtoken string `help:"personal access token" env:"SOUP_VCS_TOKEN"`
	Secret string `help:"webhook secret" env:"SOUP_VCS_WEBHOOK_SECRET"`
}
type RepoSubcmd struct {
	Path string `arg:"" help:"repo to sync"`
	VCS  VCS    `embed:"" prefix:"sync.repo.vcs."`
}

type ServeSubcmd struct {
	Path string `arg:"" help:"repo to sync"`
	VCS  VCS    `embed:"" prefix:"sync.serve.vcs."`
}
type SyncCmd struct {
	Alert    Alert       `embed:"" prefix:"sync.alert."`
	Actuator Actuator    `embed:"" prefix:"sync.actuator."`
	K8s      K8s         `embed:"" prefix:"sync.k8s."`
	Repo     RepoSubcmd  `cmd:"" help:"One-shot reconciliation"`
	Serve    ServeSubcmd `cmd:"" help:"Serve reconciliation bitbucket webhook"`
}

func initializeSyncCmd(cli *CLI, path string, vcs VCS, f *WasSetted) (application.Applications, error) {
	var apps application.Applications

	wArgs := infrastructure.WatcherArgs{
		URI: "informer:noop",
	}
	gArgs := infrastructure.SVCArgs{
		URI:     vcs.URI,
		Address: path,
	}
	infra, err := infrastructure.NewAdapters(gArgs, cli.Audit.URI, cli.Sync.Alert.URI, wArgs)
	if err != nil {
		return application.Applications{}, fmt.Errorf("initializeSyncCmd: %w", err)
	}
	// err = infra.GitRepository.Init(path,
	// 	vcs.Username,
	// 	vcs.Withtoken)
	// if err != nil {
	// 	return application.Applications{}, fmt.Errorf("initializeSyncCmd: %w", err)
	// }

	if f.contextWasSet {
		c := string(cli.Sync.K8s.Context)
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, &c)
	} else {
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, nil)
	}
	if err != nil {
		return application.Applications{}, fmt.Errorf("initializeSyncCmd: %w", err)
	}

	apps = application.NewApplications(infra.LoggerService,
		infra.NotificationService,
		infra.AuditService,
		infra.VersionRepository,
		infra.GitRepository,
		infra.DeployRepository,
		infra.SoupRepository,
		infra.ProbeRepository,
		infra.InformerService)
	apps.LoggerService.SetLevel(cli.Logging.Level)
	apps.LoggerService.SetFormat(cli.Logging.Format)

	return apps, nil
}

func (cmd *RepoSubcmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications

	if apps, err = initializeSyncCmd(cli, cmd.Path, cmd.VCS, f); err != nil {
		return fmt.Errorf("Run: %w", err)
	}

	event := audit.Event{
		Action:  "SyncRun",
		Actor:   "system",
		Message: "sync command run",
	}
	if err = apps.Auditer.Audit(&event); err != nil {
		apps.LoggerService.WithFields(logger.Fields{
			"err": fmt.Errorf("Run: %w", err),
		}).Info("check audit subsystem")
	}

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {
		var err error

		req := commands.LoopBranchesRequest{}
		if err = apps.Commands.LoopBranches.Handle(req); err != nil {
			n := notification.Notification{
				Message:     fmt.Sprintf("error deploying %s", cli.Sync.Repo.Path),
				Description: err.Error(),
				Priority:    cli.Sync.Alert.Priority,
				Tags:        append([]string(nil), cli.Sync.Alert.Tags...),
				Teams:       append([]string(nil), cli.Sync.Alert.Teams...),
			}
			apps.Notifier.Notify(n)
		}

		return err
	})
	h.SetShutdownFunc(func(s os.Signal) error {

		event := audit.Event{
			Action:  "SyncShutdown",
			Actor:   "system",
			Message: "diff command shutdown",
		}

		return apps.Auditer.Audit(&event)
	})

	ports := infrastructure.NewPorts(apps, &h)
	wg := &sync.WaitGroup{}
	ports.Actuators.SetActuatorRoot(cli.Sync.Actuator.Root)
	ports.Actuators.Start(cli.Sync.Actuator.Address, wg, false, cli.Audit.Enable, cmd.VCS.Secret)
	ports.MainLoop.Exec(wg, "sync cmd")
	wg.Wait()

	return nil
}

func (cmd *ServeSubcmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications

	if apps, err = initializeSyncCmd(cli, cmd.Path, cmd.VCS, f); err != nil {
		return fmt.Errorf("Run: %w", err)
	}

	event := audit.Event{
		Action:  "ServeRun",
		Actor:   "system",
		Message: "serve command run",
	}
	if err = apps.Auditer.Audit(&event); err != nil {
		apps.LoggerService.WithFields(logger.Fields{
			"err": fmt.Errorf("Run: %w", err),
		}).Info("check audit subsystem")
	}

	ports := infrastructure.NewPorts(apps, nil)
	wg := &sync.WaitGroup{}
	ports.Actuators.SetActuatorRoot(cli.Sync.Actuator.Root)
	ports.Actuators.Start(cli.Sync.Actuator.Address, wg, true, cli.Audit.Enable, cmd.VCS.Secret)
	wg.Wait()

	return nil
}
