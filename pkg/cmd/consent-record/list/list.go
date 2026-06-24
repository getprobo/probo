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
	"encoding/json"
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
	CreatedAt        string  `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBannerID  string
		flagAction    string
		flagVisitorID string
		flagVersion   int
		flagLimit     int
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

			records, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[consentRecord], error) {
					var resp struct {
						Node *struct {
							Typename       string                        `json:"__typename"`
							ConsentRecords api.Connection[consentRecord] `json:"consentRecords"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("cookie banner %s not found", flagBannerID)
					}

					if resp.Node.Typename != "CookieBanner" {
						return nil, fmt.Errorf("expected CookieBanner node, got %s", resp.Node.Typename)
					}

					return &resp.Node.ConsentRecords, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, records)
			}

			if len(records) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No consent records found.")
				return nil
			}

			rows := make([][]string, 0, len(records))
			for _, r := range records {
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

				rows = append(rows, []string{r.ID, r.VisitorID, r.Action, r.SDKVersion, regulation, regulationSource, countryCode, r.CreatedAt})
			}

			t := cmdutil.NewTable("ID", "VISITOR ID", "ACTION", "SDK VERSION", "REGULATION", "SOURCE", "COUNTRY", "CREATED AT").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(records) {
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "\nShowing %d of %d consent records\n", len(records), totalCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagBannerID, "banner-id", "", "Cookie banner ID (required)")
	_ = cmd.MarkFlagRequired("banner-id")
	cmd.Flags().StringVar(&flagAction, "action", "", "Filter by action")
	cmd.Flags().StringVar(&flagVisitorID, "visitor-id", "", "Filter by visitor ID")
	cmd.Flags().IntVar(&flagVersion, "version", 0, "Filter by version")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
