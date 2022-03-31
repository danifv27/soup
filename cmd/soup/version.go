package main

import (
	"encoding/json"
	"fmt"

	"github.com/danifv27/soup/internal/application"
	"github.com/danifv27/soup/internal/application/soup/queries"
	"github.com/danifv27/soup/internal/infrastructure"
)

type VersionCmd struct {
	Format string `short:"f" help:"Format the output (pretty|json)." enum:"pretty,json" default:"pretty"`
}

func (cmd *VersionCmd) Run(cli *CLI) error {
	var err error
	var info *queries.GetVersionInfoResult
	var out []byte

	infra := infrastructure.NewServices()
	infra.LoggerService.SetLevel(cli.Globals.LogLevel)
	infra.LoggerService.SetFormat(cli.Globals.LogFormat)
	infra.LoggerService.Debug("debug message test")
	apps := application.NewApplications(infra.LoggerService, infra.VersionRepository)
	if info, err = apps.Queries.GetVersionInfoHandler.Handle(); err != nil {
		return err
	}
	if cli.Version.Format == "json" {
		if out, err = json.MarshalIndent(info, "", "    "); err != nil {
			return err
		}
		fmt.Println(string(out))
	} else {
		fmt.Println(info)
	}

	return nil
}
