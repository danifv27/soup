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
	"github.com/danifv27/soup/internal/application/notification"
	"github.com/danifv27/soup/internal/application/soup/commands"
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
)

type Resource struct {
	Kind string
}

func (r Resource) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	var value string
	err := ctx.Scan.PopValueInto("value", &value)
	if err != nil {
		return err
	}
	resources := strings.Split(value, ",")
	for _, resource := range resources {
		res := Resource{}
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

type DiffCmd struct {
	// Actuator Actuator `embed:"" prefix:"actuator."`
	K8s       K8s        `embed:"" prefix:"k8s."`
	Resources []Resource `prefix:"diff." default:"v1/services,apps/v1/deployments,apps/v1/statefulsets"`
}

func initializeDiffCmd(cli *CLI, f *WasSetted) (application.Applications, error) {
	var apps application.Applications

	infra, err := infrastructure.NewAdapters(cli.Audit.DbPath, cli.Alert.Enable)
	if err != nil {
		return application.Applications{}, err
	}

	err = infra.NotificationService.Init(cli.Alert.URL, cli.Alert.Apikey)
	if err != nil {
		return application.Applications{}, err
	}

	if f.contextWasSet {
		c := string(cli.Sync.K8s.Context)
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, &c)
	} else {
		err = infra.DeployRepository.Init(cli.Sync.K8s.Path, nil)
	}
	if err != nil {
		return application.Applications{}, err
	}

	apps = application.NewApplications(infra.LoggerService,
		infra.NotificationService,
		infra.AuditService,
		infra.VersionRepository,
		infra.GitRepository,
		infra.DeployRepository,
		infra.SoupRepository,
		infra.ProbeRepository)
	apps.LoggerService.SetLevel(cli.Logging.Level)
	apps.LoggerService.SetFormat(cli.Logging.Format)

	return apps, nil
}

func (cmd *DiffCmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications

	if apps, err = initializeDiffCmd(cli, f); err != nil {
		return err
	}

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {
		var err error

		req := commands.LoopBranchesRequest{}
		if err = apps.Commands.LoopBranches.Handle(req); err != nil {
			n := notification.Notification{
				Message:     fmt.Sprintf("error deploying %s", cli.Sync.Repo.Path),
				Description: err.Error(),
				Priority:    cli.Alert.Priority,
				Tags:        append([]string(nil), cli.Alert.Tags...),
				Teams:       append([]string(nil), cli.Alert.Teams...),
			}
			apps.Notifier.Notify(n)
		}

		return err
	})
	h.SetShutdownFunc(func(s os.Signal) error {

		return nil
	})

	ports := infrastructure.NewPorts(apps, &h)
	wg := &sync.WaitGroup{}
	ports.Actuators.SetActuatorRoot(cli.Sync.Actuator.Root)
	ports.Actuators.Start(cli.Sync.Actuator.Port, wg, false, cli.Audit.Enable, "")
	ports.MainLoop.Exec(wg)
	wg.Wait()

	return nil
}
