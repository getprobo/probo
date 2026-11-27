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

package list

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query AssetSnapshotList($id: ID!, $first: Int, $after: CursorKey, $orderBy: AssetSnapshotOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      assetSnapshots(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            capturedAt
            name
            note
            actor {
              id
              fullName
            }
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

type snapshotRow struct {
	ID         string  `json:"id"`
	CapturedAt string  `json:"capturedAt"`
	Name       *string `json:"name"`
	Note       *string `json:"note"`
	Actor      *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"actor"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagLimit    int
		flagAfter    string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List asset snapshots",
		Aliases: []string{"ls"},
		Example: `  # List snapshots for an organization
  prb asset snapshot list --org ORG_ID`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			_, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagOrg,
				"orderBy": map[string]any{
					"field":     "CAPTURED_AT",
					"direction": flagOrderDir,
				},
			}

			snapshots, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[snapshotRow]("assetSnapshots", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				snapshots,
				totalCount,
				cmdutil.ListView[snapshotRow]{
					Empty:      "No asset snapshots found.",
					Noun:       "asset snapshots",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "CAPTURED", "NAME", "ACTOR"},
					Row: func(snapshot snapshotRow) ([]string, error) {
						name := ""
						if snapshot.Name != nil {
							name = *snapshot.Name
						}

						actorName := ""
						if snapshot.Actor != nil {
							actorName = snapshot.Actor.FullName
						}

						return []string{
							snapshot.ID,
							cmdutil.FormatTime(snapshot.CapturedAt),
							name,
							actorName,
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of snapshots to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
