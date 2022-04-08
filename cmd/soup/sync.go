package main

import (
	"os"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type SyncCmd struct {
	Repo struct {
		Repo     string `arg help:"repo to sync"`
		Interval int    `short:"i" help:"synchronize every" default:"120"`
		As       struct {
			Username struct {
				Username  string `arg help:"username"`
				Withtoken struct {
					Withtoken string `arg help:"personal access token"`
				} `cmd`
			} `arg`
		} `cmd`
	} `arg`
}

func (cmd *SyncCmd) Run(cli *CLI) error {
	var apps application.Applications

	infra := infrastructure.NewAdapters()
	infra.LoggerService.SetLevel(cli.Globals.Logging.Level)
	infra.LoggerService.SetFormat(cli.Globals.Logging.Format)

	apps = application.NewApplications(infra.LoggerService, infra.NotificationService, infra.VersionRepository, infra.GitRepository)

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {
		var err error

		req := commands.LoopBranchesRequest{
			URL:      cli.Sync.Repo.Repo,
			Period:   cli.Sync.Repo.Interval,
			Token:    cli.Sync.Repo.As.Username.Withtoken.Withtoken,
			Username: cli.Sync.Repo.As.Username.Username,
		}
		err = apps.Commands.LoopBranches.Handle(req)

		return err
	})
	h.SetShutdownFunc(func(s os.Signal) error {

		return nil
	})

	ports := infrastructure.NewPorts(apps, &h)
	ports.MainLoop.Exec()

	return nil
}
