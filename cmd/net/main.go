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

type NetCmd struct {
}

type CLI struct {
	Logging Log        `embed:"" prefix:"logging."`
	Version VersionCmd `cmd:"" help:"Show the version information"`
	Net     NetCmd     `cmd:"" help:"utility and GO package to wait for port to open (TCP, UDP)."`
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

// flagSet.StringVar(&conf.proto, "proto", "tcp", "tcp")
// flagSet.StringVar(&conf.addrs, "addrs", "", "address:port(,address:port,address:port,...)")
// flagSet.UintVar(&conf.deadlineMS, "deadline", 10000, "deadline in milliseconds")
// flagSet.UintVar(&conf.delayMS, "wait", 100, "delay of single request in milliseconds")
// flagSet.UintVar(&conf.breakMS, "delay", 50, "break between requests in milliseconds")
// flagSet.BoolVar(&conf.debug, "debug", false, "debug messages toggler")
// flagSet.StringVar(&conf.packetBase64, "packet", "", "UDP packet to be sent")
