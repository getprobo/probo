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

package view

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query AssetSnapshotView($id: ID!) {
  node(id: $id) {
    __typename
    ... on AssetSnapshot {
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
}
`

const assetsQuery = `
query AssetSnapshotAssets($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    __typename
    ... on AssetSnapshot {
      assets(first: $first, after: $after, orderBy: { field: NAME, direction: ASC }) {
        totalCount
        edges {
          node {
            id
            assetId
            name
            amount
            assetType
            status
            kind
            externalId
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

type (
	snapshotView struct {
		Typename   string  `json:"__typename"`
		ID         string  `json:"id"`
		CapturedAt string  `json:"capturedAt"`
		Name       *string `json:"name"`
		Note       *string `json:"note"`
		Actor      *struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"actor"`
	}

	copiedAsset struct {
		ID         string  `json:"id"`
		AssetID    string  `json:"assetId"`
		Name       string  `json:"name"`
		Amount     int     `json:"amount"`
		AssetType  string  `json:"assetType"`
		Status     string  `json:"status"`
		Kind       *string `json:"kind"`
		ExternalID *string `json:"externalId"`
	}

	viewResponse struct {
		Node *snapshotView `json:"node"`
	}
)

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit  int
		flagAfter  string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View an asset snapshot and its copied rows",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			data, err := client.Do(viewQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil || resp.Node.Typename != "AssetSnapshot" {
				return fmt.Errorf("asset snapshot %s not found", args[0])
			}

			if *flagOutput == cmdutil.OutputJSON {
				assets, _, _, err := api.Paginate(
					client,
					assetsQuery,
					map[string]any{"id": args[0]},
					flagLimit,
					flagAfter,
					api.NodeConnection[copiedAsset]("assets", "AssetSnapshot", fmt.Sprintf("asset snapshot %s not found", args[0])),
				)
				if err != nil {
					return err
				}

				return cmdutil.PrintJSON(f.IOStreams.Out, map[string]any{
					"snapshot": resp.Node,
					"assets":   assets,
				})
			}

			name := ""
			if resp.Node.Name != nil {
				name = *resp.Node.Name
			}

			note := ""
			if resp.Node.Note != nil {
				note = *resp.Node.Note
			}

			actorName := ""
			if resp.Node.Actor != nil {
				actorName = resp.Node.Actor.FullName
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"ID: %s\nCaptured: %s\nName: %s\nNote: %s\nActor: %s\n",
				resp.Node.ID,
				cmdutil.FormatTime(resp.Node.CapturedAt),
				name,
				note,
				actorName,
			)

			assets, totalCount, nextCursor, err := api.Paginate(
				client,
				assetsQuery,
				map[string]any{"id": args[0]},
				flagLimit,
				flagAfter,
				api.NodeConnection[copiedAsset]("assets", "AssetSnapshot", fmt.Sprintf("asset snapshot %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				assets,
				totalCount,
				cmdutil.ListView[copiedAsset]{
					Empty:      "No copied assets.",
					Noun:       "copied assets",
					NextCursor: nextCursor,
					Headers:    []string{"NAME", "AMOUNT", "TYPE", "KIND", "ASSET"},
					Row: func(asset copiedAsset) ([]string, error) {
						kind := ""
						if asset.Kind != nil {
							kind = *asset.Kind
						}

						return []string{
							asset.Name,
							fmt.Sprintf("%d", asset.Amount),
							asset.AssetType,
							kind,
							asset.AssetID,
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of copied rows to show")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
