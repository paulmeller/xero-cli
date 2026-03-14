package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for xero.

To load completions:

Bash:
  $ source <(xero completion bash)
  # To load completions for each session, execute once:
  $ xero completion bash > /etc/bash_completion.d/xero

Zsh:
  $ source <(xero completion zsh)
  # To load completions for each session, execute once:
  $ xero completion zsh > "${fpath[1]}/_xero"

Fish:
  $ xero completion fish | source
  # To load completions for each session, execute once:
  $ xero completion fish > ~/.config/fish/completions/xero.fish
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
