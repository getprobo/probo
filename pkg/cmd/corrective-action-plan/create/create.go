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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation CorrectiveActionPlanCreate($input: CreateCorrectiveActionPlanInput!) {
  createCorrectiveActionPlan(input: $input) {
    correctiveActionPlanEdge {
      node {
        id
        title
      }
    }
  }
}
`

type createResponse struct {
	CreateCorrectiveActionPlan struct {
		CorrectiveActionPlanEdge struct {
			Node struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"node"`
		} `json:"correctiveActionPlanEdge"`
	} `json:"createCorrectiveActionPlan"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagFinding     string
		flagTitle       string
		flagDescription string
		flagOwner       string
		flagDueDate     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a corrective action plan for a finding",
		Example: `  # Create a corrective action plan interactively
  prb corrective-action-plan create

  # Create a corrective action plan non-interactively
  prb corrective-action-plan create --finding <id> --title "Restore access reviews" --owner <id> --due-date 2027-01-15T00:00:00Z`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			if f.IOStreams.IsInteractive() {
				if flagFinding == "" {
					err := huh.NewInput().
						Title("Finding ID").
						Value(&flagFinding).
						Run()
					if err != nil {
						return err
					}
				}

				if flagTitle == "" {
					err := huh.NewInput().
						Title("Title").
						Value(&flagTitle).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagFinding == "" {
				return fmt.Errorf("finding is required; pass --finding or run interactively")
			}

			if flagTitle == "" {
				return fmt.Errorf("title is required; pass --title or run interactively")
			}

			input := map[string]any{
				"findingId": flagFinding,
				"title":     flagTitle,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}

			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}

			if flagDueDate != "" {
				input["dueDate"] = flagDueDate
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			p := resp.CreateCorrectiveActionPlan.CorrectiveActionPlanEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created corrective action plan %s\n",
				p.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagFinding, "finding", "", "Finding ID (required)")
	cmd.Flags().StringVar(&flagTitle, "title", "", "Plan title (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Plan description")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagDueDate, "due-date", "", "Due date (RFC3339)")

	return cmd
}
