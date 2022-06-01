package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/notification"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type VCS struct {
	Username  string `help:"username" env:"SOUP_VCS_USERNAME"`
	Withtoken string `help:"personal access token" env:"SOUP_VCS_TOKEN"`
	Secret    string `help:"Webhook secret" env:"SOUP_VCS_WEBHOOK_SECRET"`
}
type RepoSubcmd struct {
	Path string `arg:"" help:"repo to sync"`
	VCS  VCS    `embed:"" prefix:"vcs."`
}

type ServeSubcmd struct {
	Path string `arg:"" help:"repo to sync"`
	VCS  VCS    `embed:"" prefix:"vcs."`
}
type SyncCmd struct {
	Actuator Actuator    `embed:"" prefix:"actuator."`
	K8s      K8s         `embed:"" prefix:"k8s."`
	Repo     RepoSubcmd  `cmd:"" help:"One-shot reconciliation"`
	Serve    ServeSubcmd `cmd:"" help:"Serve reconciliation bitbucket webhook"`
}

func initializeSyncCmd(cli *CLI, path string, vcs VCS, f *WasSetted) (application.Applications, error) {
	var apps application.Applications

	infra, err := infrastructure.NewAdapters(cli.Audit.DbPath, cli.Alert.Enable)
	if err != nil {
		return application.Applications{}, err
	}
	err = infra.GitRepository.Init(path,
		vcs.Username,
		vcs.Withtoken)
	if err != nil {
		return application.Applications{}, err
	}

	err = infra.NotificationService.Init(cli.Alert.URL, cli.Alert.Apikey)
	if err != nil {
		return application.Applications{}, err
	}

	if f.contextWasSet {
		c := string(cli.Sync.K8s.Context)
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, &c)
	} else {
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, nil)
	}
	if err != nil {
		return application.Applications{}, err
	}

	apps = application.NewApplications(infra.LoggerService,
		infra.NotificationService,
		infra.AuditService,
		infra.VersionRepository,
		infra.GitRepository,
		infra.DeployRepository,
		infra.SoupRepository,
		infra.ProbeRepository)
	apps.LoggerService.SetLevel(cli.Logging.Level)
	apps.LoggerService.SetFormat(cli.Logging.Format)

	return apps, nil
}

func (cmd *RepoSubcmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications

	if apps, err = initializeSyncCmd(cli, cmd.Path, cmd.VCS, f); err != nil {
		return err
	}

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {
		var err error

		req := commands.LoopBranchesRequest{}
		if err = apps.Commands.LoopBranches.Handle(req); err != nil {
			n := notification.Notification{
				Message:     fmt.Sprintf("error deploying %s", cli.Sync.Repo.Path),
				Description: err.Error(),
				Priority:    cli.Alert.Priority,
				Tags:        append([]string(nil), cli.Alert.Tags...),
				Teams:       append([]string(nil), cli.Alert.Teams...),
			}
			apps.Notifier.Notify(n)
		}

		return err
	})
	h.SetShutdownFunc(func(s os.Signal) error {

		return nil
	})

	ports := infrastructure.NewPorts(apps, &h)
	wg := &sync.WaitGroup{}
	ports.Actuators.SetActuatorRoot(cli.Sync.Actuator.Root)
	ports.Actuators.Start(cli.Sync.Actuator.Port, wg, false, cli.Audit.Enable, cmd.VCS.Secret)
	ports.MainLoop.Exec(wg)
	wg.Wait()

	return nil
}

func (cmd *ServeSubcmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications

	if apps, err = initializeSyncCmd(cli, cmd.Path, cmd.VCS, f); err != nil {
		return err
	}

	ports := infrastructure.NewPorts(apps, nil)
	wg := &sync.WaitGroup{}
	ports.Actuators.SetActuatorRoot(cli.Sync.Actuator.Root)
	ports.Actuators.Start(cli.Sync.Actuator.Port, wg, true, cli.Audit.Enable, cmd.VCS.Secret)
	wg.Wait()

	return nil
}
