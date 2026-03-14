package cmdutil

import (
	"io"
	"os"

	"golang.org/x/term"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/config"
	"github.com/paulmeller/xero-cli/internal/output"
)

// IOStreams holds the standard I/O streams.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
	IsTTY  bool
}

// Factory provides access to shared dependencies for commands.
type Factory struct {
	Config    func() (*config.Config, error)
	APIClient func() (*api.Client, error)
	Formatter func(format string) output.Formatter
	IO        *IOStreams
}

// NewFactory creates a Factory with default implementations.
func NewFactory() *Factory {
	ios := &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	ios.IsTTY = term.IsTerminal(int(os.Stdin.Fd()))

	configFunc := func() (*config.Config, error) {
		return config.Load("")
	}

	f := &Factory{
		Config: configFunc,
		IO:     ios,
	}

	f.Formatter = func(format string) output.Formatter {
		switch format {
		case "json":
			return &output.JSONFormatter{}
		case "csv":
			return &output.CSVFormatter{}
		case "tsv":
			return &output.TSVFormatter{}
		default:
			return output.NewTableFormatter(ios.Out, ios.IsTTY)
		}
	}

	f.APIClient = func() (*api.Client, error) {
		cfg, err := configFunc()
		if err != nil {
			return nil, err
		}
		return api.NewClientFromConfig(cfg)
	}

	return f
}
