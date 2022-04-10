package main

import (
	"os"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

// type hostFlag string

// func (d hostFlag) BeforeApply(logger *logger.Logger) error {
// 	logger.SetOutput(os.Stdout)
// 	return nil
// }

type SyncCmd struct {
	Host       string `help:"if not set, kubeconfig auth will be used. If using kubectl proxy - set it to http://localhost:8001"`
	Kubeconfig string `help:"path to the kubeconfig file to use for requests"`
	Repo       struct {
		Repo string `arg:"" help:"repo to sync"`
		// Interval int    `short:"i" help:"synchronize every" default:"120" env:"SOUP_SYNC_INTERVAL"`
		As struct {
			Username struct {
				Username  string `arg:"" help:"username" env:"SOUP_SYNC_USERNAME" optional:""`
				Withtoken struct {
					Withtoken string `arg:"" help:"personal access token" env:"SOUP_SYNC_TOKEN" optional:""`
				} `cmd:""`
			} `arg:""`
		} `cmd:""`
	} `arg:""`
}

func (cmd *SyncCmd) Run(cli *CLI, apps application.Applications) error {
	apps.LoggerService.SetLevel(cli.Globals.Logging.Level)
	apps.LoggerService.SetFormat(cli.Globals.Logging.Format)

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
