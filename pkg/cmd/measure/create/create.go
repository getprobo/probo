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
mutation($input: CreateMeasureInput!) {
  createMeasure(input: $input) {
    measureEdge {
      node {
        id
        name
        category
        state
      }
    }
  }
}
`

type createResponse struct {
	CreateMeasure struct {
		MeasureEdge struct {
			Node struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Category string `json:"category"`
				State    string `json:"state"`
			} `json:"node"`
		} `json:"measureEdge"`
	} `json:"createMeasure"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg           string
		flagName          string
		flagCategory      string
		flagDescription   string
		flagThirdPartyIDs []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new measure",
		Example: `  # Create a measure interactively
  prb measure create

  # Create a measure non-interactively
  prb measure create --name "Enable encryption at rest" --category "Security"`,
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
						Title("Measure name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagCategory == "" {
					err := huh.NewInput().
						Title("Measure category").
						Value(&flagCategory).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			if flagCategory == "" {
				return fmt.Errorf("category is required; pass --category or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
				"category":       flagCategory,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}

			if len(flagThirdPartyIDs) > 0 {
				input["thirdPartyIds"] = flagThirdPartyIDs
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

			m := resp.CreateMeasure.MeasureEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created measure %s (%s)\n",
				m.ID,
				m.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Measure name (required)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Measure category (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Measure description")
	cmd.Flags().StringSliceVar(&flagThirdPartyIDs, "third-party-ids", nil, "ThirdParty IDs to link (comma-separated)")

	return cmd
}
