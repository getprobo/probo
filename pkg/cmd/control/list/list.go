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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: ControlOrder, $filter: ControlFilter) {
  node(id: $id) {
    __typename
    ... on Framework {
      controls(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            sectionTitle
            name
            description
            bestPractice
            maturityLevel
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

type control struct {
	ID            string  `json:"id"`
	SectionTitle  string  `json:"sectionTitle"`
	Name          string  `json:"name"`
	Description   *string `json:"description"`
	BestPractice  bool    `json:"bestPractice"`
	MaturityLevel string  `json:"maturityLevel"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagFramework string
		flagLimit     int
		flagAfter     string
		flagOrderBy   string
		flagOrderDir  string
		flagFilter    string
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List controls in a framework",
		Aliases: []string{"ls"},
		Example: `  # List controls in a framework
  prb control list --framework <framework-id>

  # Filter and output as JSON
  prb control ls --framework <framework-id> --filter "access" --json`,
		Args: cobra.NoArgs,
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
				"id": flagFramework,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "SECTION_TITLE"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			if flagFilter != "" {
				variables["filter"] = map[string]any{
					"query": flagFilter,
				}
			}

			controls, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[control]("controls", "Framework", fmt.Sprintf("framework %s not found", flagFramework)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				controls,
				totalCount,
				cmdutil.ListView[control]{
					Empty:      "No controls found.",
					Noun:       "controls",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "SECTION", "NAME", "BEST PRACTICE", "MATURITY"},
					Row: func(c control) ([]string, error) {
						bp := "No"
						if c.BestPractice {
							bp = "Yes"
						}

						return []string{
							c.ID,
							c.SectionTitle,
							c.Name,
							bp,
							docgen.MaturityLabel(coredata.ControlMaturityLevel(c.MaturityLevel)),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagFramework, "framework", "", "Framework ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of controls to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, SECTION_TITLE)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVarP(&flagFilter, "filter", "q", "", "Filter controls by search query")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("framework")

	return cmd
}
