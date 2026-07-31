// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package delete

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const deleteMutation = `
mutation($input: DeleteResourceTagInput!) {
  deleteResourceTag(input: $input) {
    deletedResourceTagId
  }
}
`

func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	var (
		flagID            string
		flagResourceTagID string
		flagYes           bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a resource tag",
		Example: `  # Delete a tag interactively
  prb resource-tag delete --id prbrtg_...

  # Delete without confirmation
  prb resource-tag delete --resource-tag-id prbrtg_... --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := flagID
			if id == "" {
				id = flagResourceTagID
			}

			if f.IOStreams.IsInteractive() {
				if id == "" {
					err := huh.NewInput().
						Title("Resource tag ID").
						Value(&id).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if id == "" {
				return fmt.Errorf("resource tag ID is required; pass --id or --resource-tag-id or run interactively")
			}

			if !flagYes {
				if !f.IOStreams.IsInteractive() {
					return fmt.Errorf("cannot delete resource tag: confirmation required, use --yes to confirm")
				}

				var confirmed bool

				err := huh.NewConfirm().
					Title(fmt.Sprintf("Delete resource tag %s?", id)).
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			_, err = client.Do(
				deleteMutation,
				map[string]any{
					"input": map[string]any{
						"resourceTagId": id,
					},
				},
			)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Deleted resource tag %s\n",
				id,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagID, "id", "", "Resource tag ID")
	cmd.Flags().StringVar(&flagResourceTagID, "resource-tag-id", "", "Resource tag ID")
	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
