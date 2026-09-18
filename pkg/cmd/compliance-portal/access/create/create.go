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
mutation($input: CreateCompliancePortalAccessInput!) {
  createCompliancePortalAccess(input: $input) {
    compliancePortalAccessEdge {
      node {
        id
        state
        authenticatedAt
        identity {
          fullName
          email
        }
      }
    }
  }
}
`

type createResponse struct {
	CreateCompliancePortalAccess struct {
		CompliancePortalAccessEdge struct {
			Node struct {
				ID              string  `json:"id"`
				State           string  `json:"state"`
				AuthenticatedAt *string `json:"authenticatedAt"`
				Identity        struct {
					FullName string `json:"fullName"`
					Email    string `json:"email"`
				} `json:"identity"`
			} `json:"node"`
		} `json:"compliancePortalAccessEdge"`
	} `json:"createCompliancePortalAccess"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagPortal    string
		flagEmail     string
		flagProfile   string
		flagDocuments []string
		flagReports   []string
		flagFiles     []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add a visitor to a compliance portal",
		Args:  cobra.NoArgs,
		Example: `  # Add by email
  prb compliance-portal access create --portal PORTAL_ID --email visitor@example.com

  # Add an existing organization member
  prb compliance-portal visitor create --portal PORTAL_ID --profile PROFILE_ID`,
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

			if f.IOStreams.IsInteractive() && flagEmail == "" && flagProfile == "" {
				err := huh.NewInput().
					Title("Visitor email").
					Value(&flagEmail).
					Run()
				if err != nil {
					return err
				}
			}

			if flagEmail == "" && flagProfile == "" {
				return fmt.Errorf("email or profile is required; pass --email or --profile")
			}

			input := map[string]any{
				"compliancePortalId": flagPortal,
			}

			if flagEmail != "" {
				input["email"] = flagEmail
			}

			if flagProfile != "" {
				input["profileId"] = flagProfile
			}

			if len(flagDocuments) > 0 {
				input["documents"] = flagDocuments
			}

			if len(flagReports) > 0 {
				input["reports"] = flagReports
			}

			if len(flagFiles) > 0 {
				input["compliancePortalFiles"] = flagFiles
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

			n := resp.CreateCompliancePortalAccess.CompliancePortalAccessEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created visitor access %s (%s)\n",
				n.ID,
				n.Identity.Email,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagPortal, "portal", "", "Compliance portal ID")
	_ = cmd.MarkFlagRequired("portal")
	cmd.Flags().StringVar(&flagEmail, "email", "", "Visitor email address")
	cmd.Flags().StringVar(&flagProfile, "profile", "", "Organization profile ID")
	cmd.Flags().StringSliceVar(&flagDocuments, "document", nil, "Document IDs to grant (repeatable)")
	cmd.Flags().StringSliceVar(&flagReports, "report", nil, "Report file IDs to grant (repeatable)")
	cmd.Flags().StringSliceVar(&flagFiles, "file", nil, "Compliance portal file IDs to grant (repeatable)")
	cmd.MarkFlagsMutuallyExclusive("email", "profile")

	return cmd
}
