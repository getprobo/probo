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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation CorrectiveActionPlanUpdate($input: UpdateCorrectiveActionPlanInput!) {
  updateCorrectiveActionPlan(input: $input) {
    correctiveActionPlan {
      id
      title
    }
  }
}
`

type updateResponse struct {
	UpdateCorrectiveActionPlan struct {
		CorrectiveActionPlan struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"correctiveActionPlan"`
	} `json:"updateCorrectiveActionPlan"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTitle       string
		flagDescription string
		flagOwner       string
		flagDueDate     string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a corrective action plan",
		Example: `  # Rename a plan and clear its due date
  prb corrective-action-plan update <id> --title "Rotate credentials" --due-date ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("title") {
				input["title"] = flagTitle
			}

			if cmd.Flags().Changed("description") {
				if flagDescription == "" {
					input["description"] = nil
				} else {
					input["description"] = flagDescription
				}
			}

			if cmd.Flags().Changed("owner") {
				if flagOwner == "" {
					input["ownerId"] = nil
				} else {
					input["ownerId"] = flagOwner
				}
			}

			if cmd.Flags().Changed("due-date") {
				if flagDueDate == "" {
					input["dueDate"] = nil
				} else {
					input["dueDate"] = flagDueDate
				}
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			data, err := client.Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated corrective action plan %s\n",
				resp.UpdateCorrectiveActionPlan.CorrectiveActionPlan.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTitle, "title", "", "Plan title")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Plan description (empty to clear)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID (empty to clear)")
	cmd.Flags().StringVar(&flagDueDate, "due-date", "", "Due date in RFC3339 (empty to clear)")

	return cmd
}
