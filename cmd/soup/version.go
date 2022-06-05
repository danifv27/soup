package main

import (
	"os"
	"sync"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type VersionCmd struct {
	Format string `prefix:"version." short:"f" help:"Format the output (pretty|json)." enum:"pretty,json" default:"pretty"`
}

func initializeVersionCmd(cli *CLI, f *WasSetted) (application.Applications, error) {
	var apps application.Applications

	infra, err := infrastructure.NewAdapters(cli.Audit.URL, "notifier:console") //Version command does not need to talk with opsgenie
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

func (cmd *VersionCmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications

	if apps, err = initializeVersionCmd(cli, f); err != nil {
		return err
	}

	apps.LoggerService.SetLevel(cli.Logging.Level)
	apps.LoggerService.SetFormat(cli.Logging.Format)

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {

		req := commands.PrintVersionRequest{
			Format: cli.Version.Format,
		}
		err = apps.Commands.PrintVersion.Handle(req)

		return err
	})

	h.SetShutdownFunc(func(s os.Signal) error {

		return nil
	})

	ports := infrastructure.NewPorts(apps, &h)
	wg := &sync.WaitGroup{}
	ports.MainLoop.Exec(wg)
	wg.Wait()

	return nil
}
