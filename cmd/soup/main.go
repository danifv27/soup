package main

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/infrastructure"
)

type CLI struct {
	Globals `envprefix:"SOUP_"`
	Version VersionCmd `cmd:"" help:"Show the version information" envprefix:"SOUP_VERSION_"`
	Sync    SyncCmd    `cmd:"" help:"Reconcile kubernetes with vcs contents" envprefix:"SOUP_SYNC_"`
}

type Globals struct {
	Logging struct {
		Level  string `enum:"debug,info,warn,error,fatal" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info" env:"LEVEL"`
		Format string `help:"The log target and format. Example: logger:syslog?appname=bob&local=7 or logger:stdout?json=true" default:"logger:stdout?json=false" env:"FORMAT"`
	} `embed:"" prefix:"logging." envprefix:"LOG_"`
}

type WasSetted struct {
	contextWasSet bool
}

func main() {
	var err error

	cli := CLI{
		Globals: Globals{},
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
		kong.Name(bin),
		kong.Description("GitOps operator for Kubernetes"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree: true,
		}),
		// kong.Vars{
		// 	"config_file": fmt.Sprintf("~/.%s", bin),
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
