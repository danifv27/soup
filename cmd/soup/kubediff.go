package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type Informer struct {
	URI        string        `help:"K8s Informer URI" env:"SOUP_INFORMER_URI" hidden:"" default:"informer:k8s?context=aws-dummy&path=kubeconfig-path&resync=45s&mode=diff"`
	Resources  []K8sResource `default:"v1/services,apps/v1/deployments" env:"SOUP_INFORMER_RESOURCES" help:"Resources to be watched"`
	Namespaces []string      `default:"all" env:"SOUP_INFORMER_NAMESPACES" help:"Namespaces to watch"`
}

type K8sResource struct {
	Kind string
}

func (r K8sResource) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	var value string
	err := ctx.Scan.PopValueInto("value", &value)
	if err != nil {
		return err
	}
	resources := strings.Split(value, ",")
	for _, resource := range resources {
		res := K8sResource{}
		res.Kind = resource
		target.Set(reflect.Append(target, reflect.ValueOf(res)))
	}
	// If v represents a struct
	// v := target.FieldByName("Kind")
	// if v.IsValid() {
	// 	v.SetString(value)
	// }

	return nil
}

type KubeDiffCmd struct {
	Actuator Actuator `embed:"" prefix:"kubediff.actuator."`
	Alert    Alert    `embed:"" prefix:"kubediff.alert."`
	Informer Informer `embed:"" prefix:"kubediff.informer."`
}

func initializeDiffCmd(cli *CLI, f *WasSetted) (application.Applications, error) {
	var apps application.Applications

	wArgs := infrastructure.WatcherArgs{
		URI: cli.Kubediff.Informer.URI,
		//Copy one struct to another where structs have same members and different types
		Resources:  copyResources(cli.Kubediff.Informer.Resources),
		Namespaces: cli.Kubediff.Informer.Namespaces,
	}

	gArgs := infrastructure.SVCArgs{
		URI: "svc:noop",
	}

	infra, err := infrastructure.NewAdapters(gArgs, cli.Audit.URI, cli.Kubediff.Alert.URI, wArgs)
	if err != nil {
		return application.Applications{}, fmt.Errorf("initializeDiffCmd: %w", err)
	}

	if f.contextWasSet {
		c := string(cli.Sync.K8s.Context)
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, &c)
	} else {
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, nil)
	}
	if err != nil {
		return application.Applications{}, fmt.Errorf("initializeDiffCmd: %w", err)
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

func (cmd *KubeDiffCmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications
	var stopCh chan struct{}

	if apps, err = initializeDiffCmd(cli, f); err != nil {
		return fmt.Errorf("Run: %w", err)
	}

	event := audit.Event{
		Action:  "KubeDiffRun",
		Actor:   "system",
		Message: "kubediff command run",
	}
	if err = apps.Auditer.Audit(&event); err != nil {
		apps.LoggerService.WithFields(logger.Fields{
			"err": fmt.Errorf("Run: %w", err),
		}).Info("check audit subsystem")
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
		}).Debug("Kubediff cmd shutting down")
		close(stopCh)
		wgLoop.Done()
		event := audit.Event{
			Action:  "KubeDiffShutdown",
			Actor:   "system",
			Message: "kubediff command shutdown",
		}
		return apps.Auditer.Audit(&event)
	})

	ports := infrastructure.NewPorts(apps, &h)
	ports.Actuators.SetActuatorRoot(cli.Kubediff.Actuator.Root)
	wg := &sync.WaitGroup{}
	ports.Actuators.Start(cli.Kubediff.Actuator.Address, wg, false, cli.Audit.Enable, "")
	ports.MainLoop.Exec(wg, "kubediff cmd")
	wg.Wait()

	return nil
}
