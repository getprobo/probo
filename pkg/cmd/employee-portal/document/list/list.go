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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: EmployeePortalDocumentOrder) {
  node(id: $id) {
    __typename
    ... on EmployeePortal {
      documents(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        pageInfo {
          hasNextPage
          endCursor
        }
        edges {
          node {
            id
            documentId
            title
            documentType
          }
        }
      }
    }
  }
}
`

type sharedDocument struct {
	ID           string `json:"id"`
	DocumentID   string `json:"documentId"`
	Title        string `json:"title"`
	DocumentType string `json:"documentType"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagPortal   string
		flagLimit    int
		flagAfter    string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List documents shared with an employee portal",
		Example: `  prb employee-portal document list --portal <employee-portal-id>

  prb employee-portal document ls --portal <employee-portal-id> --output json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
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
				"id": flagPortal,
				"orderBy": map[string]any{
					"field":     "CREATED_AT",
					"direction": flagOrderDir,
				},
			}

			documents, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnectionRequired[sharedDocument](
					"documents",
					"EmployeePortal",
					fmt.Sprintf("employee portal %s not found", flagPortal),
					"employee portal document list is unavailable",
				),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				documents,
				totalCount,
				cmdutil.ListView[sharedDocument]{
					Empty:      "No shared documents found.",
					Noun:       "shared documents",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "DOCUMENT", "TITLE", "TYPE"},
					Row: func(document sharedDocument) ([]string, error) {
						return []string{
							document.ID,
							document.DocumentID,
							document.Title,
							document.DocumentType,
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagPortal, "portal", "", "Employee portal ID")
	_ = cmd.MarkFlagRequired("portal")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of documents to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
