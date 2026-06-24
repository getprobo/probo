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

package publish

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const publishMutation = `
mutation($input: PublishRiskListInput!) {
  publishRiskList(input: $input) {
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
	PublishRiskList struct {
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
	} `json:"publishRiskList"`
}

func NewCmdPublish(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagApprover []string
		flagMinor    bool
	)

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish the risk register as a document version",
		Example: `  # Publish the risk register
  prb risk publish --org ORG_ID

  # Publish with approvers
  prb risk publish --org ORG_ID --approver PROFILE_ID1 --approver PROFILE_ID2`,
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			input := map[string]any{
				"organizationId": flagOrg,
				"minor":          flagMinor,
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

			v := resp.PublishRiskList.DocumentVersionEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Published risk register %s (v%d.%d)\n",
				v.Title,
				v.Major,
				v.Minor,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringArrayVar(&flagApprover, "approver", nil, "Approver profile ID (can be repeated; ignored with --minor)")
	cmd.Flags().BoolVar(&flagMinor, "minor", false, "Publish as a minor version (no approval flow)")

	return cmd
}
