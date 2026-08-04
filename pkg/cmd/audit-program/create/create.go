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

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateAuditProgramInput!) {
  createAuditProgram(input: $input) {
    auditProgramEdge {
      node {
        id
        name
        validFrom
        validUntil
      }
    }
  }
}
`

type createResponse struct {
	CreateAuditProgram struct {
		AuditProgramEdge struct {
			Node struct {
				ID         string  `json:"id"`
				Name       string  `json:"name"`
				ValidFrom  *string `json:"validFrom"`
				ValidUntil *string `json:"validUntil"`
			} `json:"node"`
		} `json:"auditProgramEdge"`
	} `json:"createAuditProgram"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg        string
		flagFramework  string
		flagName       string
		flagValidFrom  string
		flagValidUntil string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new audit program",
		Example: `  # Create an audit program
  prb audit-program create --framework <id> --name "ISO 27001 2024-2027"`,
		Args: cobra.NoArgs,
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

			input := map[string]any{
				"organizationId": flagOrg,
				"frameworkId":    flagFramework,
				"name":           flagName,
			}

			if flagValidFrom != "" {
				input["validFrom"] = flagValidFrom
			}

			if flagValidUntil != "" {
				input["validUntil"] = flagValidUntil
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

			ap := resp.CreateAuditProgram.AuditProgramEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created audit program %s (%s)\n",
				ap.ID,
				ap.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagFramework, "framework", "", "Framework ID (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "Audit program name (required)")
	cmd.Flags().StringVar(&flagValidFrom, "valid-from", "", "Valid from date (e.g. 2024-01-01)")
	cmd.Flags().StringVar(&flagValidUntil, "valid-until", "", "Valid until date (e.g. 2027-12-31)")

	_ = cmd.MarkFlagRequired("framework")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
