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

package listscantrackers

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listScanTrackersQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $direction: OrderDirection!, $trackerTypes: [TrackerType!]) {
  node(id: $id) {
    __typename
    ... on CookieBannerScan {
      detectedTrackers(
        first: $first
        after: $after
        orderBy: { field: CREATED_AT, direction: $direction }
        filter: { trackerTypes: $trackerTypes }
      ) {
        totalCount
        edges {
          node {
            id
            identifier
            observedBy
            initiatorUrl
            createdAt
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

type tracker struct {
	ID           string  `json:"id"`
	Identifier   string  `json:"identifier"`
	ObservedBy   string  `json:"observedBy"`
	InitiatorURL *string `json:"initiatorUrl"`
	CreatedAt    string  `json:"createdAt"`
}

func NewCmdListScanTrackers(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit     int
		flagAfter     string
		flagDirection string
		flagTypes     []string
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:     "list-scan-trackers <scan-id>",
		Short:   "List detections from a cookie banner scan",
		Aliases: []string{"ls-scan-trackers"},
		Args:    cobra.ExactArgs(1),
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

			variables := map[string]any{
				"id":        args[0],
				"direction": flagDirection,
			}
			if len(flagTypes) > 0 {
				variables["trackerTypes"] = flagTypes
			}

			trackers, totalCount, nextCursor, err := api.Paginate(
				client,
				listScanTrackersQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnectionOrMissing[tracker]("detectedTrackers", "CookieBannerScan", fmt.Sprintf("cookie banner scan %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				trackers,
				totalCount,
				cmdutil.ListView[tracker]{
					Empty:      "No detected trackers found.",
					Noun:       "detected trackers",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "IDENTIFIER", "OBSERVED BY", "INITIATOR", "FIRST SEEN"},
					Row: func(item tracker) ([]string, error) {
						initiator := ""
						if item.InitiatorURL != nil {
							initiator = *item.InitiatorURL
						}

						return []string{
							item.ID,
							item.Identifier,
							item.ObservedBy,
							initiator,
							cmdutil.FormatTime(item.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	cmd.Flags().StringArrayVar(&flagTypes, "type", nil, "Tracker type (COOKIE, LOCAL_STORAGE, SESSION_STORAGE, INDEXED_DB, CACHE_STORAGE)")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmdutil.AddOrderDirectionFlag(cmd, &flagDirection, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
