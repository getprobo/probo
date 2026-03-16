// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package completion

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdCompletion(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for prb.

To load completions:

Bash:
  $ source <(prb completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ prb completion bash > /etc/bash_completion.d/prb
  # macOS:
  $ prb completion bash > $(brew --prefix)/etc/bash_completion.d/prb

Zsh:
  $ source <(prb completion zsh)
  # To load completions for each session, execute once:
  $ prb completion zsh > "${fpath[1]}/_prb"

Fish:
  $ prb completion fish | source
  # To load completions for each session, execute once:
  $ prb completion fish > ~/.config/fish/completions/prb.fish

PowerShell:
  PS> prb completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, execute once:
  PS> prb completion powershell > prb.ps1`,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := f.IOStreams.Out
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(out, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(out)
			case "fish":
				return cmd.Root().GenFishCompletion(out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(out)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return cmd
}
