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
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateAuditInput!) {
  createAudit(input: $input) {
    auditEdge {
      node {
        id
        name
        state
        validFrom
        validUntil
      }
    }
  }
}
`

type createResponse struct {
	CreateAudit struct {
		AuditEdge struct {
			Node struct {
				ID         string  `json:"id"`
				Name       string  `json:"name"`
				State      string  `json:"state"`
				ValidFrom  *string `json:"validFrom"`
				ValidUntil *string `json:"validUntil"`
			} `json:"node"`
		} `json:"auditEdge"`
	} `json:"createAudit"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                        string
		flagFramework                  string
		flagName                       string
		flagState                      string
		flagValidFrom                  string
		flagValidUntil                 string
		flagCompliancePortalVisibility string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new audit",
		Example: `  # Create an audit interactively
  prb audit create

  # Create an audit non-interactively
  prb audit create --name "SOC 2 Type II 2026" --state IN_PROGRESS --valid-from 2026-01-01 --valid-until 2026-12-31`,
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
				if flagName == "" {
					err := huh.NewInput().
						Title("Audit name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagState == "" {
					err := huh.NewSelect[string]().
						Title("Audit state").
						Options(
							huh.NewOption("Not Started", "NOT_STARTED"),
							huh.NewOption("In Progress", "IN_PROGRESS"),
							huh.NewOption("Completed", "COMPLETED"),
							huh.NewOption("Rejected", "REJECTED"),
							huh.NewOption("Outdated", "OUTDATED"),
						).
						Value(&flagState).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
			}

			if flagFramework != "" {
				input["frameworkId"] = flagFramework
			}

			if flagState != "" {
				input["state"] = flagState
			}

			if flagValidFrom != "" {
				input["validFrom"] = flagValidFrom
			}

			if flagValidUntil != "" {
				input["validUntil"] = flagValidUntil
			}

			if flagCompliancePortalVisibility != "" {
				input["compliancePortalVisibility"] = flagCompliancePortalVisibility
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

			a := resp.CreateAudit.AuditEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created audit %s (%s)\n",
				a.ID,
				a.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagFramework, "framework", "", "Framework ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Audit name (required)")
	cmd.Flags().StringVar(&flagState, "state", "", "Audit state: NOT_STARTED, IN_PROGRESS, COMPLETED, REJECTED, OUTDATED")
	cmd.Flags().StringVar(&flagValidFrom, "valid-from", "", "Valid from date (e.g. 2026-01-01)")
	cmd.Flags().StringVar(&flagValidUntil, "valid-until", "", "Valid until date (e.g. 2026-12-31)")
	cmd.Flags().StringVar(&flagCompliancePortalVisibility, "compliance-portal-visibility", "", "Compliance portal visibility: NONE, PRIVATE, PUBLIC")

	return cmd
}
