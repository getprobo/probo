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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DocumentCollaborationHistorySessionOrder) {
  node(id: $id) {
    __typename
    ... on DocumentVersion {
      collaborationHistory(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            authorName
            startedAt
            endedAt
            changeCount
            heads
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

type historySession struct {
	ID          string   `json:"id"`
	AuthorName  *string  `json:"authorName"`
	StartedAt   string   `json:"startedAt"`
	EndedAt     string   `json:"endedAt"`
	ChangeCount int      `json:"changeCount"`
	Heads       []string `json:"heads"`
}

func NewCmdListHistory(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "list-history <document-version-id>",
		Short: "List collaboration history sessions for a document version",
		Example: `  # List collaboration edit sessions
  prb document list-history <document-version-id>`,
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
					[]string{"ENDED_AT"},
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

			sessions, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[historySession], error) {
					var resp struct {
						Node *struct {
							Typename             string                         `json:"__typename"`
							CollaborationHistory api.Connection[historySession] `json:"collaborationHistory"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("document version %s not found", args[0])
					}

					if resp.Node.Typename != "DocumentVersion" {
						return nil, fmt.Errorf("expected DocumentVersion node, got %s", resp.Node.Typename)
					}

					return &resp.Node.CollaborationHistory, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, sessions)
			}

			if len(sessions) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No collaboration history sessions found.")
				return nil
			}

			rows := make([][]string, 0, len(sessions))
			for _, session := range sessions {
				author := ""
				if session.AuthorName != nil {
					author = *session.AuthorName
				}

				rows = append(rows, []string{
					session.ID,
					author,
					session.StartedAt,
					session.EndedAt,
					fmt.Sprintf("%d", session.ChangeCount),
					strings.Join(session.Heads, ","),
				})
			}

			t := cmdutil.NewTable(
				"ID",
				"AUTHOR",
				"STARTED",
				"ENDED",
				"CHANGES",
				"HEADS",
			).Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(sessions) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d sessions\n",
					len(sessions),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of sessions to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (ENDED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
