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
		flagAfter  string
	)

	cmd := &cobra.Command{
		Use:   "list-linear-issues <organization-id>",
		Short: "Search Linear issues for a team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id":     args[0],
				"teamId": flagTeamID,
			}
			if flagQuery != "" {
				variables["query"] = flagQuery
			}

			issues, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[linearIssue](
					"linearIssues",
					"Organization",
					fmt.Sprintf("organization %s not found", args[0]),
				),
			)
			if err != nil {
				return fmt.Errorf("cannot list Linear issues: %w", err)
			}

			if len(issues) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No Linear issues found.")

				cmdutil.WritePageFooter(f.IOStreams, 0, totalCount, "issues", nextCursor)

				return nil
			}

			for _, issue := range issues {
				_, _ = fmt.Fprintf(f.IOStreams.Out, "%s\t%s\t%s\n", issue.ID, issue.Identifier, issue.Title)
			}

			cmdutil.WritePageFooter(f.IOStreams, len(issues), totalCount, "issues", nextCursor)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTeamID, "team-id", "", "Linear team ID")
	cmd.Flags().StringVar(&flagQuery, "query", "", "Search issues in the team")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of issues to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	_ = cmd.MarkFlagRequired("team-id")

	return cmd
}
