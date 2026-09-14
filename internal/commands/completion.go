package commands

import "github.com/spf13/cobra"

func newCompletionCommand() *cobra.Command {
	command := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion script",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(command *cobra.Command, args []string) error {
			root := command.Root()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(command.OutOrStdout(), true)
			case "zsh":
				return root.GenZshCompletion(command.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(command.OutOrStdout(), true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(command.OutOrStdout())
			}
			return nil
		},
	}
	return command
}
