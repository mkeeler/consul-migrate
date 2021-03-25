package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mkeeler/consul-migrate/internal/migrate"
	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
)

type importCommand struct {
	ui    cli.Ui
	flags *flag.FlagSet
	http  *httpFlags

	input   string
	verbose bool
	silent  bool
}

func NewImport(ui cli.Ui) (cli.Command, error) {
	c := &importCommand{
		ui:    ui,
		http:  &httpFlags{},
		flags: flag.NewFlagSet("", flag.ContinueOnError),
	}

	c.flags.BoolVar(&c.silent, "silent", false, "Disables all normal log output")
	c.flags.BoolVar(&c.verbose, "verbose", false, "Enable verbose debugging output")
	c.flags.StringVar(&c.input, "input", "", "File path to read data from. Defaults to stdin")

	flagMerge(c.flags, c.http.flags())
	return c, nil
}

func (c *importCommand) Help() string {
	return usage(importHelp, c.flags)
}

func (c *importCommand) Synopsis() string {
	return "Import Consul data"
}

func (c *importCommand) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		c.ui.Error(fmt.Sprintf("Failed to parse flags: %v", err))
		return 1
	}

	if c.verbose && c.silent {
		c.ui.Error(fmt.Sprintf("Cannot specify both -silent and -verbose"))
		return 1
	}

	level := hclog.Info
	if c.verbose {
		level = hclog.Debug
	} else if c.silent {
		level = hclog.Off
	}

	initLogging(c.ui, level)

	client, err := c.http.apiClient()
	if err != nil {
		hclog.L().Error("error connecting to Consul agent", "error", err)
		return 1
	}

	var dataBytes []byte
	if c.input != "" {
		dataBytes, err = ioutil.ReadFile(c.input)
		if err != nil {
			hclog.L().Error("error reading input file", "error", err)
			return 1
		}
	} else {
		dataBytes, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			hclog.L().Error("error reading data from stdin", "error", err)
			return 1
		}
	}

	var data migrate.Data
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		hclog.L().Error("error deserializing JSON data", "error", err)
		return 1
	}

	err = migrate.Import(client, &data)
	if err != nil {
		hclog.L().Error("error importing data", "error", err)
		return 1
	}

	hclog.L().Info("successfully imported data")
	return 0
}

const importHelp = `
Usage: consul-migrate import [options]

  Imports Consul data from the output of consul-migrate export
`
