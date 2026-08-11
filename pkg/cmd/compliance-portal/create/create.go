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
mutation($input: CreateCompliancePortalInput!) {
  createCompliancePortal(input: $input) {
    compliancePortalEdge {
      node {
        id
        entityName
      }
    }
  }
}
`

type createResponse struct {
	CreateCompliancePortal struct {
		CompliancePortalEdge struct {
			Node struct {
				ID         string `json:"id"`
				EntityName string `json:"entityName"`
			} `json:"node"`
		} `json:"compliancePortalEdge"`
	} `json:"createCompliancePortal"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg        string
		flagEntityName string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a compliance portal",
		Example: `  # Create a compliance portal interactively
  prb compliance-portal create

  # Create a compliance portal non-interactively
  prb compliance-portal create --entity-name "Acme Corp"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("cannot load configuration: %w", err)
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return fmt.Errorf("cannot load default host: %w", err)
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

			if f.IOStreams.IsInteractive() && flagEntityName == "" {
				err := huh.NewInput().
					Title("Entity name").
					Value(&flagEntityName).
					Run()
				if err != nil {
					return fmt.Errorf("cannot read entity name: %w", err)
				}
			}

			if flagEntityName == "" {
				return fmt.Errorf("entity name is required; pass --entity-name or run interactively")
			}

			data, err := client.Do(
				createMutation,
				map[string]any{
					"input": map[string]any{
						"organizationId": flagOrg,
						"entityName":     flagEntityName,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("cannot create compliance portal: %w", err)
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			portal := resp.CreateCompliancePortal.CompliancePortalEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created compliance portal %s (%s)\n",
				portal.ID,
				portal.EntityName,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagEntityName, "entity-name", "", "Entity name shown on the public compliance page")

	return cmd
}
