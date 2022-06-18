package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type KubeWatchCmd struct {
	Actuator Actuator `embed:"" prefix:"kubewatch.actuator."`
	Alert    Alert    `embed:"" prefix:"kubewatch.alert."`
	Informer Informer `embed:"" prefix:"kubewatch.informer."`
}

func initializeWatchCmd(cli *CLI, f *WasSetted) (application.Applications, error) {
	var apps application.Applications

	wArgs := infrastructure.WatcherArgs{
		URI: cli.Kubewatch.Informer.URI,
		//Copy one struct to another where structs have same members and different types
		Resources:  copyResources(cli.Kubewatch.Informer.Resources),
		Namespaces: cli.Kubewatch.Informer.Namespaces,
	}

	gArgs := infrastructure.SVCArgs{
		URI: "svc:noop",
	}

	infra, err := infrastructure.NewAdapters(gArgs, cli.Audit.URI, cli.Kubewatch.Alert.URI, wArgs)
	if err != nil {
		return application.Applications{}, fmt.Errorf("initializeWatchCmd: %w", err)
	}

	if f.contextWasSet {
		c := string(cli.Sync.K8s.Context)
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, &c)
	} else {
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, nil)
	}
	if err != nil {
		return application.Applications{}, fmt.Errorf("initializeWatchCmd: %w", err)
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

func (cmd *KubeWatchCmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications
	var stopCh chan struct{}

	if apps, err = initializeWatchCmd(cli, f); err != nil {
		return fmt.Errorf("Run: %w", err)
	}

	wgLoop := &sync.WaitGroup{}

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {
		wgLoop.Add(1)
		stopCh = make(chan struct{})

		apps.Informer.Start(stopCh)
		wgLoop.Wait()

		return nil
	})
	h.SetShutdownFunc(func(s os.Signal) error {
		apps.LoggerService.WithFields(logger.Fields{
			"signal": s,
		}).Debug("kubewatch cmd shutting down")
		close(stopCh)
		wgLoop.Done()
		event := audit.Event{
			Action:  "KubeWatchShutdown",
			Actor:   "system",
			Message: "kubewatch command shutdown",
		}
		return apps.Auditer.Audit(&event)
	})

	ports := infrastructure.NewPorts(apps, &h)
	ports.Actuators.SetActuatorRoot(cli.Kubewatch.Actuator.Root)
	wg := &sync.WaitGroup{}
	ports.Actuators.Start(cli.Kubewatch.Actuator.Address, wg, false, cli.Audit.Enable, "")
	ports.MainLoop.Exec(wg, "kubewatch cmd")
	wg.Wait()

	return nil
}
