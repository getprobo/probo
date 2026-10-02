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
mutation($input: CreateCompliancePortalCommitmentGroupInput!) {
  createCompliancePortalCommitmentGroup(input: $input) {
    compliancePortalCommitmentGroupEdge {
      node {
        id
        title
        description
        rank
      }
    }
  }
}
`

type createResponse struct {
	CreateCompliancePortalCommitmentGroup struct {
		CompliancePortalCommitmentGroupEdge struct {
			Node struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Rank        int    `json:"rank"`
			} `json:"node"`
		} `json:"compliancePortalCommitmentGroupEdge"`
	} `json:"createCompliancePortalCommitmentGroup"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagPortal      string
		flagTitle       string
		flagDescription string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a compliance portal commitment group",
		Example: `  # Create a commitment group interactively
  prb compliance-portal commitment-group create --portal <compliance-portal-id>

  # Create a commitment group non-interactively
  prb compliance-portal cg create --portal <compliance-portal-id> --title "Security" --description "Our security commitments"`,
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

			if f.IOStreams.IsInteractive() {
				if flagTitle == "" {
					err := huh.NewInput().
						Title("Group title").
						Value(&flagTitle).
						Run()
					if err != nil {
						return err
					}
				}

				if flagDescription == "" {
					err := huh.NewText().
						Title("Description").
						Value(&flagDescription).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagTitle == "" {
				return fmt.Errorf("title is required; pass --title or run interactively")
			}

			if flagDescription == "" {
				return fmt.Errorf("description is required; pass --description or run interactively")
			}

			data, err := client.Do(
				createMutation,
				map[string]any{
					"input": map[string]any{
						"compliancePortalId": flagPortal,
						"title":              flagTitle,
						"description":        flagDescription,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			g := resp.CreateCompliancePortalCommitmentGroup.CompliancePortalCommitmentGroupEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created commitment group %s (%s)\n",
				g.ID,
				g.Title,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagPortal, "portal", "", "Compliance portal ID (required)")
	_ = cmd.MarkFlagRequired("portal")
	cmd.Flags().StringVar(&flagTitle, "title", "", "Group title (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Group description (required)")

	return cmd
}
