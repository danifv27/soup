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

type contextStr string

func (f contextStr) BeforeApply(set *WasSetted) error {

	set.contextWasSet = true

	return nil
}

type SyncCmd struct {
	Path    string     `help:"path to the kubeconfig file to use for requests or host url" env:"SOUP_SYNC_K8S_PATH" prefix:"sync."`
	Context contextStr `help:"the name of the kubeconfig context to use" env:"SOUP_SYNC_K8S_CONTEXT" prefix:"sync."`
	Repo    struct {
		Repo string `arg:"" help:"repo to sync"`
		As   struct {
			Username struct {
				Username  string `arg:"" help:"username" env:"SOUP_SYNC_USERNAME" optional:""`
				Withtoken struct {
					Withtoken string `arg:"" help:"personal access token" env:"SOUP_SYNC_TOKEN" optional:""`
				} `cmd:""`
			} `arg:""`
		} `cmd:""`
	} `arg:""`
}

func (cmd *SyncCmd) Run(cli *CLI, apps application.Applications, f *WasSetted) error {
	apps.LoggerService.SetLevel(cli.Logging.Level)
	apps.LoggerService.SetFormat(cli.Logging.Format)

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {
		var err error

		req := commands.LoopBranchesRequest{
			Path: cli.Sync.Path,
		}
		err = apps.Commands.LoopBranches.Handle(req)

		return err
	})
	h.SetShutdownFunc(func(s os.Signal) error {

		return nil
	})

	ports := infrastructure.NewPorts(apps, &h)
	wg := &sync.WaitGroup{}
	ports.Actuators.SetActuatorRoot(cli.Actuator.Root)
	ports.Actuators.Start(cli.Actuator.Port, wg)
	ports.MainLoop.Exec(wg)
	wg.Wait()

	return nil
}
