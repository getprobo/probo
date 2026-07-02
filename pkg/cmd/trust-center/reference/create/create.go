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

const trustCenterQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Organization {
      trustCenter {
        id
      }
    }
  }
}
`

const createMutation = `
mutation($input: CreateTrustCenterReferenceInput!) {
  createTrustCenterReference(input: $input) {
    trustCenterReferenceEdge {
      node {
        id
        name
        description
        websiteUrl
        rank
      }
    }
  }
}
`

type trustCenterQueryResponse struct {
	Node *struct {
		Typename    string `json:"__typename"`
		TrustCenter *struct {
			ID string `json:"id"`
		} `json:"trustCenter"`
	} `json:"node"`
}

type createResponse struct {
	CreateTrustCenterReference struct {
		TrustCenterReferenceEdge struct {
			Node struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				WebsiteUrl  *string `json:"websiteUrl"`
				Rank        int     `json:"rank"`
			} `json:"node"`
		} `json:"trustCenterReferenceEdge"`
	} `json:"createTrustCenterReference"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagName        string
		flagDescription string
		flagWebsite     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a trust center reference",
		Example: `  # Create a reference interactively
  prb trust-center reference create

  # Create a reference non-interactively
  prb trust-center ref create --name "Acme Corp" --description "Enterprise customer" --website "https://acme.com"`,
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

			// Fetch trust center ID from organization.
			data, err := client.Do(
				trustCenterQuery,
				map[string]any{"id": flagOrg},
			)
			if err != nil {
				return err
			}

			var tcResp trustCenterQueryResponse
			if err := json.Unmarshal(data, &tcResp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if tcResp.Node == nil {
				return fmt.Errorf("organization %s not found", flagOrg)
			}

			if tcResp.Node.Typename != "Organization" {
				return fmt.Errorf("expected Organization node, got %s", tcResp.Node.Typename)
			}

			if tcResp.Node.TrustCenter == nil {
				return fmt.Errorf("trust center not found for organization %s", flagOrg)
			}

			if f.IOStreams.IsInteractive() {
				if flagName == "" {
					err := huh.NewInput().
						Title("Reference name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagDescription == "" {
					err := huh.NewInput().
						Title("Description (optional)").
						Value(&flagDescription).
						Run()
					if err != nil {
						return err
					}
				}

				if flagWebsite == "" {
					err := huh.NewInput().
						Title("Website URL (optional)").
						Value(&flagWebsite).
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
				"trustCenterId": tcResp.Node.TrustCenter.ID,
				"name":          flagName,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}

			if flagWebsite != "" {
				input["websiteUrl"] = flagWebsite
			}

			data, err = client.Do(
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

			r := resp.CreateTrustCenterReference.TrustCenterReferenceEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created reference %s (%s)\n",
				r.ID,
				r.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Reference name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Reference description")
	cmd.Flags().StringVar(&flagWebsite, "website", "", "Website URL")

	return cmd
}
