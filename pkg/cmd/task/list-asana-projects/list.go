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

package listasanaprojects

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $teamId: String, $first: Int, $after: String, $query: String) {
  node(id: $id) {
    __typename
    ... on Organization {
      asanaProjects(teamId: $teamId, first: $first, after: $after, query: $query) {
        edges {
          node {
            id
            name
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

type asanaProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewCmdListAsanaProjects(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTeamID string
		flagQuery  string
		flagLimit  int
		flagAfter  string
	)

	cmd := &cobra.Command{
		Use:   "list-asana-projects <organization-id>",
		Short: "Search Asana projects for an organization",
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
				"id": args[0],
			}
			if flagTeamID != "" {
				variables["teamId"] = flagTeamID
			}

			if flagQuery != "" {
				variables["query"] = flagQuery
			}

			// AsanaProjectConnection has no totalCount. Pass len(projects)
			// as the total so the "Showing N of M" line never prints and
			// only Next cursor does.
			projects, _, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[asanaProject](
					"asanaProjects",
					"Organization",
					fmt.Sprintf("organization %s not found", args[0]),
				),
			)
			if err != nil {
				return fmt.Errorf("cannot list Asana projects: %w", err)
			}

			if len(projects) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No Asana projects found.")

				cmdutil.WritePageFooter(f.IOStreams, 0, 0, "projects", nextCursor)

				return nil
			}

			for _, project := range projects {
				_, _ = fmt.Fprintf(f.IOStreams.Out, "%s\t%s\n", project.ID, project.Name)
			}

			cmdutil.WritePageFooter(f.IOStreams, len(projects), len(projects), "projects", nextCursor)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTeamID, "team-id", "", "Asana team GID used to filter projects")
	cmd.Flags().StringVar(&flagQuery, "query", "", "Search projects by name")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of projects to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)

	return cmd
}
