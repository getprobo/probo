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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TaskActivityOrder) {
  node(id: $id) {
    __typename
    ... on Task {
      activities(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            activityType
            field
            oldValue
            newValue
            createdAt
            actor {
              id
              fullName
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

type activity struct {
	ID           string  `json:"id"`
	ActivityType string  `json:"activityType"`
	Field        *string `json:"field"`
	OldValue     *string `json:"oldValue"`
	NewValue     *string `json:"newValue"`
	CreatedAt    string  `json:"createdAt"`
	Actor        *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"actor"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTask     string
		flagLimit    int
		flagAfter    string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List activities on a task",
		Aliases: []string{"ls"},
		Example: `  # List activities on a task
  prb task activity list --task <task-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			if flagTask == "" {
				return fmt.Errorf("task is required; pass --task")
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagTask,
				"orderBy": map[string]any{
					"field":     "CREATED_AT",
					"direction": flagOrderDir,
				},
			}

			activities, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[activity]("activities", "Task", fmt.Sprintf("task %s not found", flagTask)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				activities,
				totalCount,
				cmdutil.ListView[activity]{
					Empty:      "No activities found.",
					Noun:       "activities",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TYPE", "FIELD", "OLD", "NEW", "ACTOR", "CREATED"},
					Row: func(e activity) ([]string, error) {
						actorName := ""
						if e.Actor != nil {
							actorName = e.Actor.FullName
						}

						return []string{
							e.ID,
							e.ActivityType,
							stringOrEmpty(e.Field),
							stringOrEmpty(e.OldValue),
							stringOrEmpty(e.NewValue),
							actorName,
							cmdutil.FormatTime(e.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagTask, "task", "", "Task ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of activities to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func stringOrEmpty(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}
