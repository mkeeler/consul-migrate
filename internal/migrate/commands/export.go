package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
	"github.com/mkeeler/consul-migrate/internal/migrate"
)

type exportCommand struct {
	ui    cli.Ui
	flags *flag.FlagSet
	http  *httpFlags

	output  string
	verbose bool
	silent  bool
}

func NewExport(ui cli.Ui) (cli.Command, error) {
	c := &exportCommand{
		ui:    ui,
		http:  &httpFlags{},
		flags: flag.NewFlagSet("", flag.ContinueOnError),
	}

	c.flags.BoolVar(&c.silent, "silent", false, "Disables all normal log output")
	c.flags.BoolVar(&c.verbose, "verbose", false, "Enable verbose debugging output")
	c.flags.StringVar(&c.output, "output", "", "File path to output the data to. Defaults to stdout")

	flagMerge(c.flags, c.http.flags())

	return c, nil
}

func (c *exportCommand) Help() string {
	return usage(exportHelp, c.flags)
}

func (c *exportCommand) Synopsis() string {
	return "Export Consul data"
}

func (c *exportCommand) Run(args []string) int {
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

	hclog.L().Info("starting data export")
	data, err := migrate.Export(client)
	if err != nil {
		hclog.L().Error("error exporting data", "error", err)
		return 1
	}

	serialized, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		hclog.L().Error("error serializing exported data", "error", err)
		return 1
	}

	if c.output == "" {
		c.ui.Output(string(serialized))
	} else {
		if err := ioutil.WriteFile(c.output, serialized, 0600); err != nil {
			hclog.L().Error("failed to write data to file", "file", c.output, "error", err)
			return 1
		}
		hclog.L().Info("data written to file", "file", c.output)
	}

	return 0
}

const exportHelp = `
Usage: consul-migrate export [options] <output>

  Exports Consul data.
  
  <output> can be - to indicate writing to stdout
  or a file path where to write the data to.
`
