package main

import (
	"os"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type VersionCmd struct {
	Format string `short:"f" help:"Format the output (pretty|json)." enum:"pretty,json" default:"pretty"`
}

func (cmd *VersionCmd) Run(cli *CLI) error {
	// var info *queries.GetVersionInfoResult
	// var err error
	// var out []byte
	var apps application.Applications
	// var notif notification.Notification

	infra := infrastructure.NewAdapters()
	infra.LoggerService.SetLevel(cli.Globals.LogLevel)
	infra.LoggerService.SetFormat(cli.Globals.LogFormat)

	apps = application.NewApplications(infra.LoggerService, infra.NotificationService, infra.VersionRepository, infra.GitRepository)

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {

		req := commands.PrintVersionRequest{
			Format: cli.Version.Format,
		}
		return apps.Commands.PrintVersion.Handle(req)
	})

	h.SetShutdownFunc(func(s os.Signal) error {

		return nil
	})

	ports := infrastructure.NewPorts(apps, &h)
	// infra.LoggerService.Debug("debug message test")

	ports.MainLoop.Exec()

	return nil
}
