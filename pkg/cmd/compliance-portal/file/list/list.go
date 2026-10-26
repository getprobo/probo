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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: CompliancePortalFileOrder) {
  node(id: $id) {
    __typename
    ... on CompliancePortal {
      compliancePortalFiles(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            name
            category
            compliancePortalVisibility
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

type compliancePortalFile struct {
	ID                         string `json:"id"`
	Name                       string `json:"name"`
	Category                   string `json:"category"`
	CompliancePortalVisibility string `json:"compliancePortalVisibility"`
	CreatedAt                  string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagPortal   string
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List compliance portal files",
		Aliases: []string{"ls"},
		Example: `  # List files for a compliance portal
  prb compliance-portal file list --portal <compliance-portal-id>

  # List files sorted by name
  prb compliance-portal file ls --portal <id> --order-by NAME`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if flagPortal == "" {
				return fmt.Errorf("compliance portal is required; pass --portal")
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagPortal,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"NAME", "CREATED_AT", "UPDATED_AT"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			files, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				api.NodeConnection[compliancePortalFile]("compliancePortalFiles", "CompliancePortal", fmt.Sprintf("compliance portal %s not found", flagPortal)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				files,
				totalCount,
				cmdutil.ListView[compliancePortalFile]{
					Empty:   "No files found.",
					Noun:    "files",
					Headers: []string{"ID", "NAME", "CATEGORY", "VISIBILITY"},
					Row: func(file compliancePortalFile) ([]string, error) {
						return []string{
							file.ID,
							file.Name,
							file.Category,
							file.CompliancePortalVisibility,
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagPortal, "portal", "", "Compliance portal ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of files to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (NAME, CREATED_AT, UPDATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
