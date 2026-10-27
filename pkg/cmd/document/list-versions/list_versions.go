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

package listversions

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DocumentVersionOrder, $filter: DocumentVersionFilter) {
  node(id: $id) {
    __typename
    ... on Document {
      versions(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            title
            major
            minor
            status
            documentType
            classification
            publishedAt
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

type documentVersion struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Major          int     `json:"major"`
	Minor          int     `json:"minor"`
	Status         string  `json:"status"`
	DocumentType   string  `json:"documentType"`
	Classification string  `json:"classification"`
	PublishedAt    *string `json:"publishedAt"`
	CreatedAt      string  `json:"createdAt"`
}

func NewCmdListVersions(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit    int
		flagAfter    string
		flagOrderBy  string
		flagOrderDir string
		flagStatus   string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "list-versions <document-id>",
		Short: "List versions of a document",
		Example: `  # List all versions of a document
  prb document list-versions <document-id>

  # List only published versions
  prb document list-versions <document-id> --status PUBLISHED`,
		Args: cobra.ExactArgs(1),
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
				"id": args[0],
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			if flagStatus != "" {
				if err := cmdutil.ValidateEnum(
					"status",
					flagStatus,
					[]string{"PENDING_APPROVAL", "PUBLISHED", "REJECTED", "VOIDED"},
				); err != nil {
					return err
				}

				variables["filter"] = map[string]any{
					"statuses": []string{flagStatus},
				}
			}

			versions, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[documentVersion]("versions", "Document", fmt.Sprintf("document %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				versions,
				totalCount,
				cmdutil.ListView[documentVersion]{
					Empty:      "No versions found.",
					Noun:       "versions",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TITLE", "VERSION", "STATUS", "TYPE", "CLASSIFICATION"},
					Row: func(v documentVersion) ([]string, error) {
						return []string{
							v.ID,
							v.Title,
							fmt.Sprintf("%d.%d", v.Major, v.Minor),
							v.Status,
							v.DocumentType,
							v.Classification,
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of versions to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Filter by status (PENDING_APPROVAL, PUBLISHED, REJECTED, VOIDED)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
