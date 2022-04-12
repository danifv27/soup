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
	Format string `short:"f" help:"Format the output (pretty|json)." enum:"pretty,json" default:"pretty"`
}

func (cmd *VersionCmd) Run(cli *CLI, apps application.Applications) error {

	apps.LoggerService.SetLevel(cli.Globals.Logging.Level)
	apps.LoggerService.SetFormat(cli.Globals.Logging.Format)

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {
		var err error

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
