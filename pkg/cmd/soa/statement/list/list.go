// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: ApplicabilityStatementOrder) {
  node(id: $id) {
    __typename
    ... on StateOfApplicability {
      applicabilityStatements(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            applicability
            justification
            control {
              id
              sectionTitle
              name
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

type applicabilityStatement struct {
	ID            string `json:"id"`
	Applicability bool   `json:"applicability"`
	Justification string `json:"justification"`
	Control       struct {
		ID           string `json:"id"`
		SectionTitle string `json:"sectionTitle"`
		Name         string `json:"name"`
	} `json:"control"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list <soa-id>",
		Short:   "List applicability statements in a SoA",
		Aliases: []string{"ls"},
		Args:    cobra.ExactArgs(1),
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
			)

			variables := map[string]any{
				"id": args[0],
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "CONTROL_SECTION_TITLE"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			statements, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[applicabilityStatement], error) {
					var resp struct {
						Node *struct {
							Typename                string                                 `json:"__typename"`
							ApplicabilityStatements api.Connection[applicabilityStatement] `json:"applicabilityStatements"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("statement of applicability %s not found", args[0])
					}
					if resp.Node.Typename != "StateOfApplicability" {
						return nil, fmt.Errorf("expected StateOfApplicability node, got %s", resp.Node.Typename)
					}
					return &resp.Node.ApplicabilityStatements, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, statements)
			}

			if len(statements) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No applicability statements found.")
				return nil
			}

			rows := make([][]string, 0, len(statements))
			for _, s := range statements {
				applicable := "No"
				if s.Applicability {
					applicable = "Yes"
				}
				rows = append(rows, []string{
					s.ID,
					s.Control.SectionTitle,
					s.Control.Name,
					applicable,
					s.Justification,
				})
			}

			t := cmdutil.NewTable("ID", "SECTION", "CONTROL", "APPLICABLE", "JUSTIFICATION").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(statements) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d statements\n",
					len(statements),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of statements to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, CONTROL_SECTION_TITLE)")
	flagOutput = cmdutil.AddOutputFlag(cmd)
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")

	return cmd
}
