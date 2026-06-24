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

package deletedraft

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const deleteDraftMutation = `
mutation($input: DeleteDocumentDraftInput!) {
  deleteDocumentDraft(input: $input) {
    document {
      id
    }
  }
}
`

func NewCmdDeleteDraft(f *cmdutil.Factory) *cobra.Command {
	var flagYes bool

	cmd := &cobra.Command{
		Use:   "delete-draft <document-id>",
		Short: "Delete the draft version of a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flagYes {
				if !f.IOStreams.IsInteractive() {
					return fmt.Errorf("cannot delete draft: confirmation required, use --yes to confirm")
				}

				var confirmed bool

				err := huh.NewConfirm().
					Title(fmt.Sprintf("Delete draft for document %s?", args[0])).
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
				deleteDraftMutation,
				map[string]any{
					"input": map[string]any{
						"documentId": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Deleted draft for document %s\n",
				args[0],
			)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
