package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/alecthomas/kong"
)

type Log struct {
	Level  string `enum:"debug,info,warn,error,fatal" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info" env:"SOUP_LOGGING_LEVEL"`
	Format string `help:"The log target and format. Example: logger:syslog?appname=bob&local=7 or logger:stdout?json=true" default:"logger:stdout?json=false" env:"SOUP_LOGGING_FORMAT"`
}

type Auditer struct {
	DbPath string `help:"Path to audit database" env:"SOUP_AUDIT_PATH" hidden:"" default:"audit:clover?path=/tmp/soup-audit&collection=audit"`
	Enable bool   `help:"Enable audit endpoint" env:"SOUP_AUDIT_ENABLE_ENDPOINT" hidden:"" default:"true"`
}
type Actuator struct {
	Port string `help:"actuator port" default:":8081" env:"SOUP_SYNC_ACTUATOR_PORT" optional:""`
	Root string `help:"actuator root" default:"/probe" env:"SOUP_SYNC_ACTUATOR_ROOT" optional:"" hidden:""`
}

type K8s struct {
	Path    string     `help:"path to the kubeconfig file to use for requests or host url" env:"SOUP_K8S_PATH"`
	Context contextStr `help:"the name of the kubeconfig context to use" env:"SOUP_K8S_CONTEXT"`
}

type Alert struct {
	URL      string   `help:"the URL for the Alert API" env:"SOUP_ALERT_URL"`
	Apikey   string   `help:"token used for authenticating API requests" env:"SOUP_ALERT_APIKEY"`
	Priority string   `enum:"P1,P2,P3,P4" help:"The priority of alert" default:"P3" env:"SOUP_ALERT_PRIORITY" hidden:""`
	Tags     []string `help:"list of labels attached to the alert" env:"SOUP_ALERT_TAGS" hidden:""`
	Teams    []string `help:"list of teams for setting responders" env:"SOUP_ALERT_TEAMS" hidden:""`
}

type CLI struct {
	Logging Log        `embed:"" prefix:"logging."`
	Audit   Auditer    `embed:"" prefix:"audit."`
	Alert   Alert      `embed:"" prefix:"alert."`
	Version VersionCmd `cmd:"" help:"Show the version information"`
	Sync    SyncCmd    `cmd:"" help:"Sync kubernetes with VCS contents"`
	Diff    DiffCmd    `cmd:"" help:"Kubernetes resource diff"`
}

func main() {
	var err error

	cli := CLI{
		Logging: Log{},
		Version: VersionCmd{},
		Sync:    SyncCmd{},
		Diff:    DiffCmd{},
	}

	setted := WasSetted{
		contextWasSet: false,
	}

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
		kong.TypeMapper(reflect.TypeOf([]Resource{}), Resource{}),
	)
	err = ctx.Run(&cli)
	ctx.FatalIfErrorf(err)
}
