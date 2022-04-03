package main

import (
	"os"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type SyncCmd struct {
	Repo     string `short:"r" help:"url of the repository."`
	Interval string `short:"i" help:"execution interval." default:"120"`
}

func (cmd *SyncCmd) Run(cli *CLI) error {
	var apps application.Applications

	infra := infrastructure.NewAdapters()
	infra.LoggerService.SetLevel(cli.Globals.LogLevel)
	infra.LoggerService.SetFormat(cli.Globals.LogFormat)

	apps = application.NewApplications(infra.LoggerService, infra.NotificationService, infra.VersionRepository, infra.GitRepository)

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {

		return nil
	})
	h.SetShutdownFunc(func(s os.Signal) error {

		return nil
	})

	ports := infrastructure.NewPorts(apps, &h)
	ports.MainLoop.Exec()

	return nil
}
