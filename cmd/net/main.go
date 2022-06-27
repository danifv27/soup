package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
)

type Log struct {
	Level  string `enum:"debug,info,warn,error,fatal" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info" env:"SOUP_LOGGING_LEVEL"`
	Format string `help:"The log target and format. Example: logger:syslog?appname=bob&local=7 or logger:stdout?json=true" default:"logger:stdout?json=false" env:"SOUP_LOGGING_FORMAT"`
}

type CLI struct {
	Logging Log        `embed:"" prefix:"logging."`
	Version VersionCmd `cmd:"" help:"Show the version information"`
	Wait    WaitCmd    `cmd:"" help:"Wait for port to open (TCP, UDP)."`
}

type WaitCmd struct {
	Flags WaitFlags `embed:""`
}

type VersionCmd struct {
	Flags VersionFlags `embed:""`
}

func main() {
	var err error

	cli := CLI{
		Logging: Log{},
		Version: VersionCmd{},
	}

	// setted := WasSetted{
	// 	contextWasSet: false,
	// }

	bin := filepath.Base(os.Args[0])
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	exBin := filepath.Base(ex)
	//config file has precedence over envars
	ctx := kong.Parse(&cli,
		// kong.Bind(&setted),
		kong.Name(bin),
		kong.Description("Utility to wait for port to open (TCP, UDP)"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Tree: true,
		}),
		kong.Configuration(kong.JSON, fmt.Sprintf("/etc/%s.json", bin), fmt.Sprintf("~/.%s.json", bin), fmt.Sprintf("%s/.%s.json", exPath, exBin)),
	)
	err = ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
