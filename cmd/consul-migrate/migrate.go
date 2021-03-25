package main

import (
	"os"

	"github.com/hashicorp/consul-migrate/internal/migrate"
	"github.com/hashicorp/consul-migrate/internal/migrate/commands"
	"github.com/mitchellh/cli"
)

func main() {
	ui := &cli.ColoredUi{
		Ui: &cli.BasicUi{
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
		OutputColor: cli.UiColorNone,
		ErrorColor:  cli.UiColorRed,
		InfoColor:   cli.UiColorBlue,
		WarnColor:   cli.UiColorYellow,
	}

	app := cli.NewCLI("consul-migrate", migrate.Version)
	app.Args = os.Args[1:]
	app.Commands = map[string]cli.CommandFactory{
		"export": func() (cli.Command, error) { return commands.NewExport(ui) },
		"import": func() (cli.Command, error) { return commands.NewImport(ui) },
	}

	exitStatus, err := app.Run()
	if err != nil {
		ui.Error(err.Error())
	}

	os.Exit(exitStatus)
}
