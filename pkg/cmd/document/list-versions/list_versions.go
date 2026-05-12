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

package listversions

import (
	"encoding/json"
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
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT"},
				); err != nil {
					return err
				}
				if err := cmdutil.ValidateEnum(
					"order-direction",
					flagOrderDir,
					[]string{"ASC", "DESC"},
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
					[]string{"DRAFT", "PENDING_APPROVAL", "PUBLISHED"},
				); err != nil {
					return err
				}
				variables["filter"] = map[string]any{
					"statuses": []string{flagStatus},
				}
			}

			versions, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[documentVersion], error) {
					var resp struct {
						Node *struct {
							Typename string                          `json:"__typename"`
							Versions api.Connection[documentVersion] `json:"versions"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("document %s not found", args[0])
					}
					if resp.Node.Typename != "Document" {
						return nil, fmt.Errorf("expected Document node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Versions, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, versions)
			}

			if len(versions) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No versions found.")
				return nil
			}

			rows := make([][]string, 0, len(versions))
			for _, v := range versions {
				rows = append(rows, []string{
					v.ID,
					v.Title,
					fmt.Sprintf("%d.%d", v.Major, v.Minor),
					v.Status,
					v.DocumentType,
					v.Classification,
				})
			}

			t := cmdutil.NewTable("ID", "TITLE", "VERSION", "STATUS", "TYPE", "CLASSIFICATION").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(versions) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d versions\n",
					len(versions),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of versions to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Filter by status (DRAFT, PENDING_APPROVAL, PUBLISHED)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
