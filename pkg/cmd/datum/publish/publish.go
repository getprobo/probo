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

package publish

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const publishMutation = `
mutation($input: PublishDataListInput!) {
  publishDataList(input: $input) {
    documentEdge {
      node {
        id
        status
        createdAt
      }
    }
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

type publishResponse struct {
	PublishDataList struct {
		DocumentEdge struct {
			Node struct {
				ID        string `json:"id"`
				Status    string `json:"status"`
				CreatedAt string `json:"createdAt"`
			} `json:"node"`
		} `json:"documentEdge"`
		DocumentVersionEdge struct {
			Node struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Major  int    `json:"major"`
				Minor  int    `json:"minor"`
				Status string `json:"status"`
			} `json:"node"`
		} `json:"documentVersionEdge"`
	} `json:"publishDataList"`
}

func NewCmdPublish(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagApprover []string
	)

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish the data list as a document version",
		Example: `  # Publish the data list
  prb datum publish --org ORG_ID

  # Publish with approvers
  prb datum publish --org ORG_ID --approver PROFILE_ID1 --approver PROFILE_ID2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			if flagOrg == "" {
				flagOrg = hc.Organization
			}
			if flagOrg == "" {
				return fmt.Errorf("organization is required: pass --org or run `prb auth login`")
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
			)

			input := map[string]any{
				"organizationId": flagOrg,
			}

			if len(flagApprover) > 0 {
				input["approverIds"] = flagApprover
			}

			data, err := client.Do(
				publishMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp publishResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			v := resp.PublishDataList.DocumentVersionEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Published data list %s (v%d.%d)\n",
				v.Title,
				v.Major,
				v.Minor,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringArrayVar(&flagApprover, "approver", nil, "Approver profile ID (can be repeated)")

	return cmd
}
