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

package listlinearissues

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $teamId: String!, $first: Int, $after: String, $query: String) {
  node(id: $id) {
    __typename
    ... on Organization {
      linearIssues(teamId: $teamId, first: $first, after: $after, query: $query) {
        edges {
          node {
            id
            identifier
            title
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
`

type linearIssue struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
}

func NewCmdListLinearIssues(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTeamID string
		flagQuery  string
		flagLimit  int
	)

	cmd := &cobra.Command{
		Use:   "list-linear-issues <organization-id>",
		Short: "Search Linear issues for a team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
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

			variables := map[string]any{
				"id":     args[0],
				"teamId": flagTeamID,
			}
			if flagQuery != "" {
				variables["query"] = flagQuery
			}

			issues, _, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[linearIssue], error) {
					var resp struct {
						Node *struct {
							Typename     string                      `json:"__typename"`
							LinearIssues api.Connection[linearIssue] `json:"linearIssues"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, fmt.Errorf("cannot parse Linear issues: %w", err)
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", args[0])
					}

					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}

					return &resp.Node.LinearIssues, nil
				},
			)
			if err != nil {
				return fmt.Errorf("cannot list Linear issues: %w", err)
			}

			if len(issues) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No Linear issues found.")

				return nil
			}

			for _, issue := range issues {
				_, _ = fmt.Fprintf(f.IOStreams.Out, "%s\t%s\t%s\n", issue.ID, issue.Identifier, issue.Title)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTeamID, "team-id", "", "Linear team ID")
	cmd.Flags().StringVar(&flagQuery, "query", "", "Search issues in the team")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of issues to list")
	_ = cmd.MarkFlagRequired("team-id")

	return cmd
}
