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

type SyncCmd struct {
	Path     string `help:"path to the kubeconfig file to use for requests or host url" env:"SOUP_SYNC_PATH"`
	Actuator string `help:"actuator port" default:":8081" env:"SOUP_SYNC_ACTUATOR" optional:"" hidden:""`
	Repo     struct {
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
			Token:    cli.Sync.Repo.As.Username.Withtoken.Withtoken,
			Username: cli.Sync.Repo.As.Username.Username,
			Path:     cli.Sync.Path,
		}
		err = apps.Commands.LoopBranches.Handle(req)

		return err
	})
	h.SetShutdownFunc(func(s os.Signal) error {

		return nil
	})

	ports := infrastructure.NewPorts(apps, &h)
	wg := &sync.WaitGroup{}
	ports.Actuators.Start(cli.Sync.Actuator, wg)
	ports.MainLoop.Exec(wg)
	wg.Wait()

	return nil
}
