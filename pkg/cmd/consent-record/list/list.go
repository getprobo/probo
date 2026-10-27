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
query($id: ID!, $first: Int, $after: CursorKey, $filter: CookieConsentRecordFilter) {
  node(id: $id) {
    __typename
    ... on CookieBanner {
      consentRecords(first: $first, after: $after, filter: $filter) {
        totalCount
        edges {
          node {
            id
            visitorId
            action
            sdkVersion
            regulation
            regulationSource
            countryCode
            subdivisionCode
            origin
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

type consentRecord struct {
	ID               string  `json:"id"`
	VisitorID        string  `json:"visitorId"`
	Action           string  `json:"action"`
	SDKVersion       string  `json:"sdkVersion"`
	Regulation       *string `json:"regulation"`
	RegulationSource *string `json:"regulationSource"`
	CountryCode      *string `json:"countryCode"`
	SubdivisionCode  *string `json:"subdivisionCode"`
	Origin           *string `json:"origin"`
	CreatedAt        string  `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBannerID  string
		flagAction    string
		flagVisitorID string
		flagVersion   int
		flagLimit     int
		flagAfter     string
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List cookie consent records for a banner",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{"id": flagBannerID}

			filter := map[string]any{}
			if cmd.Flags().Changed("action") {
				filter["action"] = flagAction
			}

			if cmd.Flags().Changed("visitor-id") {
				filter["visitorId"] = flagVisitorID
			}

			if cmd.Flags().Changed("version") {
				filter["version"] = flagVersion
			}

			if len(filter) > 0 {
				variables["filter"] = filter
			}

			records, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[consentRecord]("consentRecords", "CookieBanner", fmt.Sprintf("cookie banner %s not found", flagBannerID)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				records,
				totalCount,
				cmdutil.ListView[consentRecord]{
					Empty:      "No consent records found.",
					Noun:       "consent records",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "VISITOR ID", "ACTION", "SDK VERSION", "REGULATION", "SOURCE", "COUNTRY", "SUBDIVISION", "ORIGIN", "CREATED AT"},
					Row: func(r consentRecord) ([]string, error) {
						regulation := "-"
						if r.Regulation != nil {
							regulation = *r.Regulation
						}

						regulationSource := "-"
						if r.RegulationSource != nil {
							regulationSource = *r.RegulationSource
						}

						countryCode := "-"
						if r.CountryCode != nil {
							countryCode = *r.CountryCode
						}

						subdivisionCode := "-"
						if r.SubdivisionCode != nil {
							subdivisionCode = *r.SubdivisionCode
						}

						origin := "-"
						if r.Origin != nil && *r.Origin != "" {
							origin = *r.Origin
						}

						return []string{r.ID, r.VisitorID, r.Action, r.SDKVersion, regulation, regulationSource, countryCode, subdivisionCode, origin, r.CreatedAt}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagBannerID, "banner-id", "", "Cookie banner ID (required)")
	_ = cmd.MarkFlagRequired("banner-id")
	cmd.Flags().StringVar(&flagAction, "action", "", "Filter by action")
	cmd.Flags().StringVar(&flagVisitorID, "visitor-id", "", "Filter by visitor ID")
	cmd.Flags().IntVar(&flagVersion, "version", 0, "Filter by version")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
