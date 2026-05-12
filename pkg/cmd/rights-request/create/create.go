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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateRightsRequestInput!) {
  createRightsRequest(input: $input) {
    rightsRequestEdge {
      node {
        id
        requestType
        requestState
        dataSubject
      }
    }
  }
}
`

type createResponse struct {
	CreateRightsRequest struct {
		RightsRequestEdge struct {
			Node struct {
				ID           string `json:"id"`
				RequestType  string `json:"requestType"`
				RequestState string `json:"requestState"`
				DataSubject  string `json:"dataSubject"`
			} `json:"node"`
		} `json:"rightsRequestEdge"`
	} `json:"createRightsRequest"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagDataSubject string
		flagType        string
		flagState       string
		flagContact     string
		flagDetails     string
		flagDeadline    string
		flagActionTaken string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new rights request",
		Example: `  # Create a rights request interactively
  prb rights-request create

  # Create a rights request non-interactively
  prb rights-request create --data-subject "John Doe" --type ACCESS --state TODO`,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			if f.IOStreams.IsInteractive() {
				if flagDataSubject == "" {
					err := huh.NewInput().
						Title("Data subject").
						Value(&flagDataSubject).
						Run()
					if err != nil {
						return err
					}
				}

				if flagType == "" {
					err := huh.NewSelect[string]().
						Title("Request type").
						Options(
							huh.NewOption("Access", "ACCESS"),
							huh.NewOption("Deletion", "DELETION"),
							huh.NewOption("Portability", "PORTABILITY"),
						).
						Value(&flagType).
						Run()
					if err != nil {
						return err
					}
				}

				if flagState == "" {
					err := huh.NewSelect[string]().
						Title("Request state").
						Options(
							huh.NewOption("To Do", "TODO"),
							huh.NewOption("In Progress", "IN_PROGRESS"),
							huh.NewOption("Done", "DONE"),
						).
						Value(&flagState).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagDataSubject == "" {
				return fmt.Errorf("data subject is required; pass --data-subject or run interactively")
			}
			if flagType == "" {
				return fmt.Errorf("request type is required; pass --type or run interactively")
			}
			if flagState == "" {
				return fmt.Errorf("request state is required; pass --state or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"requestType":    flagType,
				"requestState":   flagState,
				"dataSubject":    flagDataSubject,
			}

			if flagContact != "" {
				input["contact"] = flagContact
			}
			if flagDetails != "" {
				input["details"] = flagDetails
			}
			if flagDeadline != "" {
				input["deadline"] = flagDeadline
			}
			if flagActionTaken != "" {
				input["actionTaken"] = flagActionTaken
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

			r := resp.CreateRightsRequest.RightsRequestEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created rights request %s\n",
				r.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagDataSubject, "data-subject", "", "Data subject name (required)")
	cmd.Flags().StringVar(&flagType, "type", "", "Request type: ACCESS, DELETION, PORTABILITY (required)")
	cmd.Flags().StringVar(&flagState, "state", "", "Request state: TODO, IN_PROGRESS, DONE (required)")
	cmd.Flags().StringVar(&flagContact, "contact", "", "Contact information")
	cmd.Flags().StringVar(&flagDetails, "details", "", "Request details")
	cmd.Flags().StringVar(&flagDeadline, "deadline", "", "Deadline")
	cmd.Flags().StringVar(&flagActionTaken, "action-taken", "", "Action taken")

	return cmd
}
