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

package delete

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const deleteMutation = `
mutation($input: DeleteFrameworkInput!) {
  deleteFramework(input: $input) {
    deletedFrameworkId
  }
}
`

func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var flagYes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a framework",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flagYes {
				if !f.IOStreams.IsInteractive() {
					return fmt.Errorf("cannot delete framework: confirmation required, use --yes to confirm")
				}

				var confirmed bool
				err := huh.NewConfirm().
					Title(fmt.Sprintf("Delete framework %s?", args[0])).
					Value(&confirmed).
					Run()
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
			)

			_, err = client.Do(
				deleteMutation,
				map[string]any{
					"input": map[string]any{
						"frameworkId": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Deleted framework %s\n",
				args[0],
			)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
