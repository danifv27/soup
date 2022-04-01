package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/danifv27/soup/internal/infrastructure"
)

type SyncCmd struct {
	Repo     string `short:"r" help:"url of the repository."`
	Interval string `short:"i" help:"execution interval." default:"120"`
}

func (cmd *SyncCmd) Run(cli *CLI) error {

	// var err error
	// var apps application.Applications

	wg := &sync.WaitGroup{}
	wg.Add(1)

	infra := infrastructure.NewAdapters()
	infra.LoggerService.SetLevel(cli.Globals.LogLevel)
	infra.LoggerService.SetFormat(cli.Globals.LogFormat)
	infra.SigHandler.SetRunFunc(func() error {
		fmt.Println("Executing sync loop")

		return nil
	})
	infra.SigHandler.SetShutdownFunc(func(s os.Signal) error {
		fmt.Println("Shutting down sync loop")

		return nil
	})
	// apps = application.NewApplications(infra.LoggerService, infra.SigHandler, infra.VersionRepository)

	go func() {
		infra.SigHandler.Run()
		wg.Done()
	}()

	wg.Wait()

	return nil
}
