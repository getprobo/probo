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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateDocumentInput!) {
  updateDocument(input: $input) {
    document {
      id
      trustCenterVisibility
      versions(first: 1) {
        edges {
          node {
            title
          }
        }
      }
    }
  }
}
`

type updateResponse struct {
	UpdateDocument struct {
		Document struct {
			ID                    string `json:"id"`
			TrustCenterVisibility string `json:"trustCenterVisibility"`
			Versions              struct {
				Edges []struct {
					Node struct {
						Title string `json:"title"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versions"`
		} `json:"document"`
	} `json:"updateDocument"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var flagTrustCenterVisibility string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a document",
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

			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("trust-center-visibility") {
				if err := cmdutil.ValidateEnum(
					"trust-center-visibility",
					flagTrustCenterVisibility,
					[]string{"NONE", "PRIVATE", "PUBLIC"},
				); err != nil {
					return err
				}
				input["trustCenterVisibility"] = flagTrustCenterVisibility
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			data, err := client.Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			doc := resp.UpdateDocument.Document
			title := doc.ID
			if len(doc.Versions.Edges) > 0 {
				title = doc.Versions.Edges[0].Node.Title
			}
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated document %s (%s)\n",
				doc.ID,
				title,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTrustCenterVisibility, "trust-center-visibility", "", "Trust center visibility: NONE, PRIVATE, PUBLIC")

	return cmd
}
