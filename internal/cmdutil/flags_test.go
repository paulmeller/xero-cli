package cmdutil

import (
	"testing"

	"github.com/paulmeller/xero-cli/internal/config"
	"github.com/spf13/cobra"
)

func TestHasChangedFilterFlags_Default(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("where", "", "")
	cmd.Flags().String("order", "", "")
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().String("modified-since", "", "")

	if HasChangedFilterFlags(cmd) {
		t.Error("HasChangedFilterFlags should return false when no flags are changed")
	}
}

func TestHasChangedFilterFlags_WhereChanged(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("where", "", "")
	cmd.Flags().String("order", "", "")
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().String("modified-since", "", "")

	if err := cmd.Flags().Set("where", "Status==\"PAID\""); err != nil {
		t.Fatal(err)
	}

	if !HasChangedFilterFlags(cmd) {
		t.Error("HasChangedFilterFlags should return true when --where is set")
	}
}

func TestHasChangedFilterFlags_OrderChanged(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("where", "", "")
	cmd.Flags().String("order", "", "")
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().String("modified-since", "", "")

	if err := cmd.Flags().Set("order", "Date DESC"); err != nil {
		t.Fatal(err)
	}

	if !HasChangedFilterFlags(cmd) {
		t.Error("HasChangedFilterFlags should return true when --order is set")
	}
}

func TestHasChangedFilterFlags_PageChanged(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("where", "", "")
	cmd.Flags().String("order", "", "")
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().String("modified-since", "", "")

	if err := cmd.Flags().Set("page", "2"); err != nil {
		t.Fatal(err)
	}

	if !HasChangedFilterFlags(cmd) {
		t.Error("HasChangedFilterFlags should return true when --page is set")
	}
}

func TestHasChangedFilterFlags_ModifiedSinceChanged(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("where", "", "")
	cmd.Flags().String("order", "", "")
	cmd.Flags().Int("page", 0, "")
	cmd.Flags().String("modified-since", "", "")

	if err := cmd.Flags().Set("modified-since", "2024-01-01"); err != nil {
		t.Fatal(err)
	}

	if !HasChangedFilterFlags(cmd) {
		t.Error("HasChangedFilterFlags should return true when --modified-since is set")
	}
}

func TestGetPageSize_Default(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().Int("page-size", 0, "")
	cmd := &cobra.Command{Use: "list"}
	root.AddCommand(cmd)

	f := &Factory{
		Config: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
	}

	got := GetPageSize(cmd, f)
	if got != 100 {
		t.Errorf("GetPageSize = %d, want 100 (default)", got)
	}
}

func TestGetPageSize_FlagSet(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().Int("page-size", 0, "")
	cmd := &cobra.Command{Use: "list"}
	root.AddCommand(cmd)

	if err := root.PersistentFlags().Set("page-size", "50"); err != nil {
		t.Fatal(err)
	}

	f := &Factory{
		Config: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
	}

	got := GetPageSize(cmd, f)
	if got != 50 {
		t.Errorf("GetPageSize = %d, want 50", got)
	}
}

func TestGetPageSize_CappedAt100(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().Int("page-size", 0, "")
	cmd := &cobra.Command{Use: "list"}
	root.AddCommand(cmd)

	if err := root.PersistentFlags().Set("page-size", "500"); err != nil {
		t.Fatal(err)
	}

	f := &Factory{
		Config: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
	}

	got := GetPageSize(cmd, f)
	if got != 100 {
		t.Errorf("GetPageSize = %d, want 100 (capped)", got)
	}
}

func TestGetPageSize_FromConfig(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().Int("page-size", 0, "")
	cmd := &cobra.Command{Use: "list"}
	root.AddCommand(cmd)

	f := &Factory{
		Config: func() (*config.Config, error) {
			return &config.Config{
				Defaults: config.Defaults{PageSize: 25},
			}, nil
		},
	}

	got := GetPageSize(cmd, f)
	if got != 25 {
		t.Errorf("GetPageSize = %d, want 25 (from config)", got)
	}
}

func TestGetPageSize_ConfigCappedAt100(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().Int("page-size", 0, "")
	cmd := &cobra.Command{Use: "list"}
	root.AddCommand(cmd)

	f := &Factory{
		Config: func() (*config.Config, error) {
			return &config.Config{
				Defaults: config.Defaults{PageSize: 200},
			}, nil
		},
	}

	got := GetPageSize(cmd, f)
	if got != 100 {
		t.Errorf("GetPageSize = %d, want 100 (config capped)", got)
	}
}
