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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: BusinessFunctionOrder, $filter: BusinessFunctionFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      businessFunctions(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            referenceId
            name
            classification
            mtdMinutes
            rtoMinutes
            rpoMinutes
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

type businessFunction struct {
	ID             string `json:"id"`
	ReferenceID    string `json:"referenceId"`
	Name           string `json:"name"`
	Classification string `json:"classification"`
	MTDMinutes     int    `json:"mtdMinutes"`
	RTOMinutes     int    `json:"rtoMinutes"`
	RPOMinutes     int    `json:"rpoMinutes"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg            string
		flagLimit          int
		flagOrderBy        string
		flagOrderDir       string
		flagClassification string
		flagOwner          string
		flagCIFOnly        bool
		flagOutput         *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List business functions in an organization",
		Aliases: []string{"ls"},
		Example: `  # List business functions in the default organization
  prb business-function list

  # List critical and important business functions only
  prb business-function ls --cif-only --json`,
		Args: cobra.NoArgs,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT", "REFERENCE_ID", "NAME", "CLASSIFICATION", "MTD_MINUTES", "RTO_MINUTES", "RPO_MINUTES"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			filter := map[string]any{}

			if flagClassification != "" {
				if err := cmdutil.ValidateEnum(
					"classification",
					flagClassification,
					[]string{"CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"},
				); err != nil {
					return err
				}

				filter["classification"] = flagClassification
			}

			if flagOwner != "" {
				filter["ownerId"] = flagOwner
			}

			if flagCIFOnly {
				filter["cifOnly"] = true
			}

			if len(filter) > 0 {
				variables["filter"] = filter
			}

			businessFunctions, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[businessFunction], error) {
					var resp struct {
						Node *struct {
							Typename          string                           `json:"__typename"`
							BusinessFunctions api.Connection[businessFunction] `json:"businessFunctions"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}

					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}

					return &resp.Node.BusinessFunctions, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, businessFunctions)
			}

			if len(businessFunctions) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No business functions found.")
				return nil
			}

			rows := make([][]string, 0, len(businessFunctions))
			for _, bf := range businessFunctions {
				rows = append(
					rows,
					[]string{
						bf.ID,
						bf.ReferenceID,
						bf.Name,
						bf.Classification,
						fmt.Sprintf("%d", bf.MTDMinutes),
						fmt.Sprintf("%d", bf.RTOMinutes),
						fmt.Sprintf("%d", bf.RPOMinutes),
					},
				)
			}

			t := cmdutil.NewTable(
				"ID",
				"REFERENCE ID",
				"NAME",
				"CLASSIFICATION",
				"MTD",
				"RTO",
				"RPO",
			).Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(businessFunctions) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d business functions\n",
					len(businessFunctions),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of business functions to list")
	cmd.Flags().StringVar(
		&flagOrderBy,
		"order-by",
		"",
		"Order by field (CREATED_AT, REFERENCE_ID, NAME, CLASSIFICATION, MTD_MINUTES, RTO_MINUTES, RPO_MINUTES)",
	)
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagClassification, "classification", "", "Filter by classification (CRITICAL, IMPORTANT, SECONDARY, STANDARD)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Filter by owner profile ID")
	cmd.Flags().BoolVar(&flagCIFOnly, "cif-only", false, "Return only critical and important business functions")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
