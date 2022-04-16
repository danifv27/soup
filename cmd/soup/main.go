package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/infrastructure"
)

type Log struct {
	Level  string `enum:"debug,info,warn,error,fatal" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info" env:"SOUP_LOGGING_LEVEL"`
	Format string `help:"The log target and format. Example: logger:syslog?appname=bob&local=7 or logger:stdout?json=true" default:"logger:stdout?json=false" env:"SOUP_LOGGING_FORMAT"`
}

type Actuator struct {
	Port string `help:"actuator port" default:":8081" env:"SOUP_SYNC_ACTUATOR_PORT" optional:"" hidden:""`
	Root string `help:"actuator root" default:"/probe" env:"SOUP_SYNC_ACTUATOR_ROOT" optional:"" hidden:""`
}
type CLI struct {
	Logging  Log        `embed:"" prefix:"logging."`
	Actuator Actuator   `embed:"" prefix:"actuator."`
	Version  VersionCmd `cmd:"" help:"Show the version information"`
	Sync     SyncCmd    `cmd:"" help:"Reconcile kubernetes with vcs contents"`
}

type WasSetted struct {
	contextWasSet bool
}

func main() {
	var err error

	cli := CLI{
		Logging: Log{},
		// Globals: Globals{},
		Version: VersionCmd{},
		Sync:    SyncCmd{},
	}
	setted := WasSetted{
		contextWasSet: false,
	}
	// cli.Globals.Flags.ContextWasSet = false
	bin := filepath.Base(os.Args[0])
	ctx := kong.Parse(&cli,
		kong.Bind(&setted),
		kong.Configuration(kong.JSON, fmt.Sprintf("/etc/%s.json", bin), fmt.Sprintf("~/.%s.json", bin), fmt.Sprintf("./%s.json", bin)),
		kong.Name(bin),
		kong.Description("GitOps operator for Kubernetes"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree: true,
		}),
		// kong.Vars{
		// 	"config_file": fmt.Sprintf("~/.%s.json", bin),
		// },
	)
	infra := infrastructure.NewAdapters()
	infra.GitRepository.Init(cli.Sync.Repo.Repo,
		cli.Sync.Repo.As.Username.Username,
		cli.Sync.Repo.As.Username.Withtoken.Withtoken)
	if setted.contextWasSet {
		c := string(cli.Sync.Context)
		err = infra.DeployRepository.Init(cli.Sync.Path, &c)
	} else {
		err = infra.DeployRepository.Init(cli.Sync.Path, nil)
	}
	ctx.FatalIfErrorf(err)
	apps := application.NewApplications(infra.LoggerService,
		infra.NotificationService,
		infra.VersionRepository,
		infra.GitRepository,
		infra.DeployRepository,
		infra.SoupRepository,
		infra.ProbeRepository)

	err = ctx.Run(&cli, apps)
	ctx.FatalIfErrorf(err)
}
