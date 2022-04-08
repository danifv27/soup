package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Globals `envprefix:"SOUP_"`
	Version VersionCmd `cmd:"" help:"Show the version information" envprefix:"SOUP_VERSION_"`
	Sync    SyncCmd    `cmd:"" help:"Reconcile kubernetes and vcs" envprefix:"SOUP_SYNC_"`
}

type Globals struct {
	Logging struct {
		Level  string `enum:"debug,info,warn,error,fatal" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info" env:"LEVEL"`
		Format string `help:"The log target and format. Example: logger:syslog?appname=bob&local=7 or logger:stdout?json=true" default:"logger:stdout?json=false" env:"FORMAT"`
	} `embed:"" prefix:"logging." envprefix:"LOG_"`
}

func main() {
	cli := CLI{}
	bin := filepath.Base(os.Args[0])

	ctx := kong.Parse(&cli,
		kong.Name(bin),
		kong.Description("GitOps operator for Kubernetes"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree: true,
		}),
		kong.Vars{
			"config_file": fmt.Sprintf("~/.%s", bin),
		},
	)
	err := ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
