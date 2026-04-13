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

package createdraft

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createDraftMutation = `
mutation($input: CreateDraftDocumentVersionInput!) {
  createDraftDocumentVersion(input: $input) {
    documentVersionEdge {
      node {
        id
        title
        major
        minor
        status
      }
    }
  }
}
`

type createDraftResponse struct {
	CreateDraftDocumentVersion struct {
		DocumentVersionEdge struct {
			Node struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Major  int    `json:"major"`
				Minor  int    `json:"minor"`
				Status string `json:"status"`
			} `json:"node"`
		} `json:"documentVersionEdge"`
	} `json:"createDraftDocumentVersion"`
}

func NewCmdCreateDraft(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-draft <document-id>",
		Short: "Create a new draft version of a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			data, err := client.Do(
				createDraftMutation,
				map[string]any{
					"input": map[string]any{
						"documentID": args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp createDraftResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			v := resp.CreateDraftDocumentVersion.DocumentVersionEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created draft version %s (%s v%d.%d)\n",
				v.ID,
				v.Title,
				v.Major,
				v.Minor,
			)

			return nil
		},
	}

	return cmd
}
