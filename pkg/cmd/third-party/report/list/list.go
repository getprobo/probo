// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package list

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: ThirdPartyComplianceReportOrder) {
  node(id: $id) {
    __typename
    ... on ThirdParty {
      complianceReports(first: $first, after: $after, orderBy: $orderBy) {
        edges {
          node {
            id
            reportName
            reportDate
            validUntil
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

type complianceReport struct {
	ID         string  `json:"id"`
	ReportName string  `json:"reportName"`
	ReportDate string  `json:"reportDate"`
	ValidUntil *string `json:"validUntil"`
	CreatedAt  string  `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list <third-party-id>",
		Short:   "List compliance reports for a third party",
		Aliases: []string{"ls"},
		Example: `  # List compliance reports for a third party
  prb third-party report list <third-party-id>

  # List compliance reports sorted by report date
  prb third-party report ls <third-party-id> --order-by REPORT_DATE --json`,
		Args: cobra.ExactArgs(1),
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

			variables := map[string]any{
				"id": args[0],
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"REPORT_DATE", "CREATED_AT"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			reports, _, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[complianceReport], error) {
					var resp struct {
						Node *struct {
							Typename          string                           `json:"__typename"`
							ComplianceReports api.Connection[complianceReport] `json:"complianceReports"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("third party %s not found", args[0])
					}

					if resp.Node.Typename != "ThirdParty" {
						return nil, fmt.Errorf("expected ThirdParty node, got %s", resp.Node.Typename)
					}

					return &resp.Node.ComplianceReports, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, reports)
			}

			if len(reports) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No compliance reports found.")
				return nil
			}

			rows := make([][]string, 0, len(reports))
			for _, r := range reports {
				validUntil := ""
				if r.ValidUntil != nil {
					validUntil = cmdutil.FormatTime(*r.ValidUntil)
				}

				rows = append(rows, []string{
					r.ID,
					r.ReportName,
					cmdutil.FormatTime(r.ReportDate),
					validUntil,
					cmdutil.FormatTime(r.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "NAME", "REPORT DATE", "VALID UNTIL", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			return nil
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of compliance reports to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (REPORT_DATE, CREATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
