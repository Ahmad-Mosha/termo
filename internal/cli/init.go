package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// hooks run `termo remind check` before each prompt. It prints nothing
// unless a reminder fired, so it stays out of the way.
var hooks = map[string]string{
	"zsh": `_termo_check() { command termo remind check 2>/dev/null }
autoload -Uz add-zsh-hook
add-zsh-hook precmd _termo_check
`,
	"bash": `_termo_check() { command termo remind check 2>/dev/null; }
PROMPT_COMMAND="_termo_check${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
`,
	"fish": `function _termo_check --on-event fish_prompt
    command termo remind check 2>/dev/null
end
`,
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <zsh|bash|fish>",
		Short: "Show reminders in your terminal when they fire",
		Long: `Print a shell hook that shows fired reminders at your next prompt.
Add the line for your shell to its config file:

  zsh   ~/.zshrc                     eval "$(termo init zsh)"
  bash  ~/.bashrc                    eval "$(termo init bash)"
  fish  ~/.config/fish/config.fish   termo init fish | source`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"zsh", "bash", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			hook, ok := hooks[args[0]]
			if !ok {
				return fmt.Errorf("unsupported shell %q; use zsh, bash or fish", args[0])
			}
			fmt.Fprint(cmd.OutOrStdout(), hook)
			return nil
		},
	}
}
