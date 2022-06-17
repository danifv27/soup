package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type VersionCmd struct {
	Format string `prefix:"version." short:"f" help:"Format the output (pretty|json)." enum:"pretty,json" default:"pretty"`
}

func initializeVersionCmd(cli *CLI, f *WasSetted) (application.Applications, error) {
	var apps application.Applications

	wArgs := infrastructure.WatcherArgs{
		URI: "informer:noop",
	}

	gArgs := infrastructure.SVCArgs{
		URI: "svc:noop",
	}
	infra, err := infrastructure.NewAdapters(gArgs, cli.Audit.URI, "notifier:console", wArgs) //Version command does not need to talk with opsgenie
	if err != nil {
		return application.Applications{}, fmt.Errorf("initializeVersionCmd: %w", err)
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

func (cmd *VersionCmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications

	if apps, err = initializeVersionCmd(cli, f); err != nil {
		return fmt.Errorf("Run: %w", err)
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

		event := audit.Event{
			Action:  "VersionShutdown",
			Actor:   "system",
			Message: "version command shutdown",
		}
		return apps.Auditer.Audit(&event)
	})

	ports := infrastructure.NewPorts(apps, &h)
	wg := &sync.WaitGroup{}
	ports.MainLoop.Exec(wg)
	wg.Wait()

	return nil
}
