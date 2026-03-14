package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/paulmeller/xero-cli/internal/api"
	"github.com/paulmeller/xero-cli/internal/cmdutil"
	"github.com/paulmeller/xero-cli/internal/output"
)

var trackingColumns = []output.Column{
	{Header: "ID", Path: "TrackingCategoryID"},
	{Header: "NAME", Path: "Name"},
	{Header: "STATUS", Path: "Status", Format: "status"},
	{Header: "OPTIONS", Path: "Options.#"},
}

func newTrackingCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tracking",
		Aliases: []string{"tracking-categories"},
		Short:   "Manage tracking categories and options",
	}

	cmd.AddCommand(newTrackingListCmd(f))
	cmd.AddCommand(newTrackingGetCmd(f))
	cmd.AddCommand(newTrackingCreateCmd(f))
	cmd.AddCommand(newTrackingUpdateCmd(f))
	cmd.AddCommand(newTrackingDeleteCmd(f))
	cmd.AddCommand(newTrackingOptionsCmd(f))

	return cmd
}

func newTrackingListCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tracking categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			data, err := client.Get(cmd.Context(), api.PathTrackingCategories, nil)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			items := gjson.ParseBytes(data).Get("TrackingCategories")
			return formatter.FormatList(f.IO.Out, items, trackingColumns)
		},
	}
}

func newTrackingGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <category-id>",
		Short: "Get a tracking category with its options",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s", api.PathTrackingCategories, args[0])
			data, err := client.Get(cmd.Context(), path, nil)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			item := gjson.ParseBytes(data).Get("TrackingCategories.0")
			if !item.Exists() {
				item = gjson.ParseBytes(data)
			}
			return formatter.FormatOne(f.IO.Out, item, trackingColumns)
		},
	}
}

func newTrackingCreateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tracking category",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := api.TrackingCategory{Name: name}
			result, err := client.Post(cmd.Context(), api.PathTrackingCategories, body, "")
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result).Get("TrackingCategories.0"), trackingColumns)
		},
	}

	cmd.Flags().String("name", "", "Category name")
	return cmd
}

func newTrackingUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <category-id>",
		Short: "Update a tracking category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := api.TrackingCategory{Name: name}
			path := fmt.Sprintf("%s/%s", api.PathTrackingCategories, args[0])
			result, err := client.Post(cmd.Context(), path, body, "")
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result).Get("TrackingCategories.0"), trackingColumns)
		},
	}

	cmd.Flags().String("name", "", "Category name")
	return cmd
}

func newTrackingDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <category-id>",
		Short: "Delete a tracking category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmAction(f.IO, fmt.Sprintf("Delete tracking category %s?", args[0]), cmd) {
				return fmt.Errorf("aborted; use --force to skip confirmation")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s", api.PathTrackingCategories, args[0])
			_, err = client.Delete(cmd.Context(), path)
			if err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Deleted tracking category %s\n", args[0])
			}
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")
	return cmd
}

func newTrackingOptionsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options",
		Short: "Manage tracking options within a category",
	}

	cmd.AddCommand(newTrackingOptionAddCmd(f))
	cmd.AddCommand(newTrackingOptionUpdateCmd(f))
	cmd.AddCommand(newTrackingOptionDeleteCmd(f))

	return cmd
}

func newTrackingOptionAddCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <category-id>",
		Short: "Add an option to a tracking category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := api.TrackingOption{Name: name}
			path := fmt.Sprintf("%s/%s/Options", api.PathTrackingCategories, args[0])
			result, err := client.Put(cmd.Context(), path, body)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result), nil)
		},
	}

	cmd.Flags().String("name", "", "Option name")
	return cmd
}

func newTrackingOptionUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <category-id> <option-id>",
		Short: "Update a tracking option",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := api.TrackingOption{Name: name}
			path := fmt.Sprintf("%s/%s/Options/%s", api.PathTrackingCategories, args[0], args[1])
			result, err := client.Post(cmd.Context(), path, body, "")
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(cmd, f.IO)
			formatter := f.Formatter(format)
			return formatter.FormatOne(f.IO.Out, gjson.ParseBytes(result), nil)
		},
	}

	cmd.Flags().String("name", "", "Option name")
	return cmd
}

func newTrackingOptionDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <category-id> <option-id>",
		Short: "Delete a tracking option",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmAction(f.IO, "Delete this tracking option?", cmd) {
				return fmt.Errorf("aborted; use --force to skip confirmation")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}
			cmdutil.ApplyClientFlags(cmd, client, f)

			path := fmt.Sprintf("%s/%s/Options/%s", api.PathTrackingCategories, args[0], args[1])
			_, err = client.Delete(cmd.Context(), path)
			if err != nil {
				return err
			}

			quiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
			if !quiet {
				fmt.Fprintf(f.IO.ErrOut, "Deleted tracking option\n")
			}
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")
	return cmd
}

// Unused import guard
var _ json.RawMessage
