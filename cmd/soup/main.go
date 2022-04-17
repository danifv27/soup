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
	Port string `help:"actuator port" default:":8081" env:"SOUP_SYNC_ACTUATOR_PORT" optional:""`
	Root string `help:"actuator root" default:"/probe" env:"SOUP_SYNC_ACTUATOR_ROOT" optional:"" hidden:""`
}

type Alert struct {
	URL      string   `help:"the URL for the Alert API" env:"SOUP_ALERT_URL"`
	Apikey   string   `help:"token used for authenticating API requests" env:"SOUP_ALERT_APIKEY"`
	Priority string   `enum:"P1,P2,P3,P4" help:"The priority of alert" default:"P3" env:"SOUP_ALERT_PRIORITY" hidden:""`
	Tags     []string `help:"list of labels attached to the alert" env:"SOUP_ALERT_TAGS" hidden:""`
	Teams    []string `help:"list of teams for setting responders" env:"SOUP_ALERT_TEAMS" hidden:""`
}

type CLI struct {
	Logging  Log        `embed:"" prefix:"logging."`
	Actuator Actuator   `embed:"" prefix:"actuator."`
	Alert    Alert      `embed:"" prefix:"alert."`
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
	//config file has precedence over envars
	ctx := kong.Parse(&cli,
		kong.Bind(&setted),
		kong.Configuration(kong.JSON, fmt.Sprintf("/etc/%s.json", bin), fmt.Sprintf("~/.%s.json", bin), fmt.Sprintf("./%s.json", bin)),
		kong.Name(bin),
		kong.Description("GitOps tool for Kubernetes"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree: true,
		}),
	)
	infra, err := infrastructure.NewAdapters()
	ctx.FatalIfErrorf(err)
	err = infra.GitRepository.Init(cli.Sync.Repo.Repo,
		cli.Sync.Repo.As.Username.Username,
		cli.Sync.Repo.As.Username.Withtoken.Withtoken)
	ctx.FatalIfErrorf(err)

	if setted.contextWasSet {
		c := string(cli.Sync.Context)
		err = infra.DeployRepository.Init(cli.Sync.Path, &c)
	} else {
		err = infra.DeployRepository.Init(cli.Sync.Path, nil)
	}
	ctx.FatalIfErrorf(err)

	err = infra.NotificationService.Init(cli.Alert.URL, cli.Alert.Apikey)
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
