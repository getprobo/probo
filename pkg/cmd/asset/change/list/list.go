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
query AssetChangeEventList($id: ID!, $first: Int, $after: CursorKey, $orderBy: AssetChangeEventOrder) {
  node(id: $id) {
    __typename
    ... on Asset {
      changeEvents(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            changeType
            before
            after
            observedAt
            assetSyncId
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

type changeEvent struct {
	ID          string  `json:"id"`
	ChangeType  string  `json:"changeType"`
	Before      *string `json:"before"`
	After       *string `json:"after"`
	ObservedAt  string  `json:"observedAt"`
	AssetSyncID *string `json:"assetSyncId"`
	Actor       *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"actor"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagAsset    string
		flagLimit    int
		flagAfter    string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List change events on an asset",
		Aliases: []string{"ls"},
		Example: `  # List change events on an asset
  prb asset change list --asset <asset-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			if flagAsset == "" {
				return fmt.Errorf("asset is required; pass --asset")
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagAsset,
				"orderBy": map[string]any{
					"field":     "OBSERVED_AT",
					"direction": flagOrderDir,
				},
			}

			events, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[changeEvent]("changeEvents", "Asset", fmt.Sprintf("asset %s not found", flagAsset)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				events,
				totalCount,
				cmdutil.ListView[changeEvent]{
					Empty:      "No change events found.",
					Noun:       "change events",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TYPE", "ACTOR", "SYNC", "OBSERVED"},
					Row: func(event changeEvent) ([]string, error) {
						actorName := ""
						if event.Actor != nil {
							actorName = event.Actor.FullName
						}

						syncID := ""
						if event.AssetSyncID != nil {
							syncID = *event.AssetSyncID
						}

						return []string{
							event.ID,
							event.ChangeType,
							actorName,
							syncID,
							cmdutil.FormatTime(event.ObservedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagAsset, "asset", "", "Asset ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of change events to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
