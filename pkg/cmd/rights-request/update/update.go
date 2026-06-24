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
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateRightsRequestInput!) {
  updateRightsRequest(input: $input) {
    rightsRequest {
      id
      requestType
      requestState
      dataSubject
    }
  }
}
`

type updateResponse struct {
	UpdateRightsRequest struct {
		RightsRequest struct {
			ID           string `json:"id"`
			RequestType  string `json:"requestType"`
			RequestState string `json:"requestState"`
			DataSubject  string `json:"dataSubject"`
		} `json:"rightsRequest"`
	} `json:"updateRightsRequest"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagType        string
		flagState       string
		flagDataSubject string
		flagContact     string
		flagDetails     string
		flagDeadline    string
		flagActionTaken string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a rights request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("type") {
				input["requestType"] = flagType
			}

			if cmd.Flags().Changed("state") {
				input["requestState"] = flagState
			}

			if cmd.Flags().Changed("data-subject") {
				input["dataSubject"] = flagDataSubject
			}

			if cmd.Flags().Changed("contact") {
				input["contact"] = flagContact
			}

			if cmd.Flags().Changed("details") {
				input["details"] = flagDetails
			}

			if cmd.Flags().Changed("deadline") {
				input["deadline"] = flagDeadline
			}

			if cmd.Flags().Changed("action-taken") {
				input["actionTaken"] = flagActionTaken
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
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

			r := resp.UpdateRightsRequest.RightsRequest
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated rights request %s\n",
				r.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagType, "type", "", "Request type: ACCESS, DELETION, PORTABILITY")
	cmd.Flags().StringVar(&flagState, "state", "", "Request state: TODO, IN_PROGRESS, DONE")
	cmd.Flags().StringVar(&flagDataSubject, "data-subject", "", "Data subject name")
	cmd.Flags().StringVar(&flagContact, "contact", "", "Contact information")
	cmd.Flags().StringVar(&flagDetails, "details", "", "Request details")
	cmd.Flags().StringVar(&flagDeadline, "deadline", "", "Deadline")
	cmd.Flags().StringVar(&flagActionTaken, "action-taken", "", "Action taken")

	return cmd
}
