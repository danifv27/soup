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
	"github.com/danifv27/soup/internal/infrastructure"
	"github.com/danifv27/soup/internal/infrastructure/signals"
	"github.com/danifv27/soup/internal/infrastructure/watcher/kubernetes"
)

type Informer struct {
	URI        string        `help:"K8s Informer URI" env:"SOUP_INFORMER_URI" hidden:"" default:"informer:k8s?context=aws-dummy&path=kubeconfig-path&resync=45s&mode=diff"`
	Resources  []K8sResource `default:"v1/services,apps/v1/deployments" env:"SOUP_INFORMER_RESOURCES" help:"Resources to be watched"`
	Namespaces []string      `default:"all" env:"SOUP_INFORMER_NAMESPACES" help:"Namespace to watch"`
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

type DiffCmd struct {
	Actuator Actuator `embed:"" prefix:"diff.actuator."`
	Alert    Alert    `embed:"" prefix:"diff.alert."`
	K8s      K8s      `embed:"" prefix:"diff.k8s."`
	Informer Informer `embed:"" prefix:"diff.informer."`
}

func copyResources(resources []K8sResource) []kubernetes.Resource {
	var res []kubernetes.Resource

	for _, r := range resources {
		resource := kubernetes.Resource(r)
		res = append(res, resource)
	}

	return res
}

func initializeDiffCmd(cli *CLI, f *WasSetted) (application.Applications, error) {
	var apps application.Applications

	wArgs := infrastructure.WatcherArgs{
		URI: cli.Diff.Informer.URI,
		//Copy one struct to another where structs have same members and different types
		Resources:  copyResources(cli.Diff.Informer.Resources),
		Namespaces: cli.Diff.Informer.Namespaces,
	}

	gArgs := infrastructure.SVCArgs{
		URI: "svc:noop",
	}

	infra, err := infrastructure.NewAdapters(gArgs, cli.Audit.URI, cli.Diff.Alert.URI, wArgs)
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

func (cmd *DiffCmd) Run(cli *CLI, f *WasSetted) error {
	var err error
	var apps application.Applications

	if apps, err = initializeDiffCmd(cli, f); err != nil {
		return fmt.Errorf("Run: %w", err)
	}

	h := signals.NewSignalHandler([]os.Signal{syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM}, apps.LoggerService)
	h.SetRunFunc(func() error {
		var err error

		req := commands.LoopBranchesRequest{}
		if err = apps.Commands.LoopBranches.Handle(req); err != nil {
			n := notification.Notification{
				Message:     fmt.Sprintf("error deploying %s", cli.Sync.Repo.Path),
				Description: err.Error(),
				Priority:    cli.Diff.Alert.Priority,
				Tags:        append([]string(nil), cli.Diff.Alert.Tags...),
				Teams:       append([]string(nil), cli.Diff.Alert.Teams...),
			}
			apps.Notifier.Notify(n)
		}

		return err
	})
	h.SetShutdownFunc(func(s os.Signal) error {

		event := audit.Event{
			Action:  "DiffShutdown",
			Actor:   "system",
			Message: "diff command shutdown",
		}
		return apps.Auditer.Audit(&event)
	})

	ports := infrastructure.NewPorts(apps, &h)
	wg := &sync.WaitGroup{}
	ports.Actuators.SetActuatorRoot(cli.Diff.Actuator.Root)
	ports.Actuators.Start(cli.Diff.Actuator.Address, wg, false, cli.Audit.Enable, "")
	ports.MainLoop.Exec(wg, "diff cmd")
	wg.Wait()

	return nil
}
