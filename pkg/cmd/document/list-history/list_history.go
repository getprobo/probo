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

package listhistory

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query DocumentListHistory($id: ID!, $first: Int, $after: CursorKey, $orderBy: DocumentRevisionOrder, $filter: DocumentRevisionFilter) {
  node(id: $id) {
    __typename
    ... on Document {
      history(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            authorName
            startedAt
            endedAt
            changeCount
            heads
            synthetic
            versions {
              id
              major
              minor
              status
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

type revision struct {
	ID          string   `json:"id"`
	AuthorName  *string  `json:"authorName"`
	StartedAt   string   `json:"startedAt"`
	EndedAt     string   `json:"endedAt"`
	ChangeCount int      `json:"changeCount"`
	Heads       []string `json:"heads"`
	Synthetic   bool     `json:"synthetic"`
	Versions    []struct {
		ID     string `json:"id"`
		Major  int    `json:"major"`
		Minor  int    `json:"minor"`
		Status string `json:"status"`
	} `json:"versions"`
}

func NewCmdListHistory(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit     int
		flagAfter     string
		flagOrderBy   string
		flagOrderDir  string
		flagVersioned bool
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:   "list-history <document-id>",
		Short: "List the history of a document",
		Long: `List the history of a document, newest first.

Every entry is a revision of the document. A published version is a tag on
the revision it was taken at and shows in the VERSIONS column as
<major>.<minor> (<status>); read its content with "prb document view-version
<version-id>". A "*" marks a version that cannot be tied to a revision of the
history.`,
		Example: `  # List the history
  prb document list-history <document-id>

  # List only the revisions tagged by a version
  prb document list-history <document-id> --versioned`,
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
					[]string{"ENDED_AT"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			if flagVersioned {
				variables["filter"] = map[string]any{
					"versioned": true,
				}
			}

			sessions, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[revision]("history", "Document", fmt.Sprintf("document %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				sessions,
				totalCount,
				cmdutil.ListView[revision]{
					Empty:      "No revisions found.",
					Noun:       "sessions",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "AUTHOR", "STARTED", "ENDED", "CHANGES", "VERSIONS", "HEADS"},
					Row: func(session revision) ([]string, error) {
						author := ""
						if session.AuthorName != nil {
							author = *session.AuthorName
						}

						tags := make([]string, len(session.Versions))
						for i, v := range session.Versions {
							tags[i] = fmt.Sprintf("%d.%d (%s)", v.Major, v.Minor, v.Status)
							if session.Synthetic {
								tags[i] += " *"
							}
						}

						return []string{
							session.ID,
							author,
							session.StartedAt,
							session.EndedAt,
							fmt.Sprintf("%d", session.ChangeCount),
							strings.Join(tags, ", "),
							strings.Join(session.Heads, ","),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of sessions to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (ENDED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	cmd.Flags().BoolVar(&flagVersioned, "versioned", false, "Only list revisions tagged by a version")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
