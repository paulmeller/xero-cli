package cmdutil

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
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
	Config             func() (*config.Config, error)
	APIClient          func() (*api.Client, error)
	APIClientFromToken func(token, tenantID string) *api.Client
	Formatter          func(format string) output.Formatter
	IO                 *IOStreams
	TenantID           func(cmd *cobra.Command) (string, error)
}

// BindTokenFlag should be called from the root command's PersistentPreRunE.
// When --token is set, it overrides f.APIClient to use the external token.
func BindTokenFlag(cmd *cobra.Command, f *Factory) {
	token, _ := cmd.Root().PersistentFlags().GetString("token")
	if token == "" {
		return
	}
	f.APIClient = func() (*api.Client, error) {
		tenantID, err := ResolveTenantID(cmd, f)
		if err != nil {
			return nil, err
		}
		return f.APIClientFromToken(token, tenantID), nil
	}
}

// ResolveTenantID returns the effective tenant ID from --tenant flag or config.
func ResolveTenantID(cmd *cobra.Command, f *Factory) (string, error) {
	if t, _ := cmd.Root().PersistentFlags().GetString("tenant"); t != "" {
		return t, nil
	}
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	if cfg.ActiveTenant == "" {
		return "", fmt.Errorf("no tenant configured; set --tenant or run 'xero tenants switch'")
	}
	return cfg.ActiveTenant, nil
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

	f.APIClientFromToken = func(token, tenantID string) *api.Client {
		return api.NewClientFromToken(token, tenantID)
	}

	f.TenantID = func(cmd *cobra.Command) (string, error) {
		return ResolveTenantID(cmd, f)
	}

	return f
}
