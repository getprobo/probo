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
	"go.probo.inc/probo/pkg/cmd/internal-control/fields"
)

const createMutation = `
mutation($input: CreateInternalControlInput!) {
  createInternalControl(input: $input) {
    internalControlEdge {
      node {
        id
        name
        category
        code
        implementationStatus
        state
      }
    }
  }
}
`

type createResponse struct {
	CreateInternalControl struct {
		InternalControlEdge struct {
			Node struct {
				ID                   string  `json:"id"`
				Name                 string  `json:"name"`
				Category             string  `json:"category"`
				Code                 *string `json:"code"`
				ImplementationStatus string  `json:"implementationStatus"`
				State                string  `json:"state"`
			} `json:"node"`
		} `json:"internalControlEdge"`
	} `json:"createInternalControl"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                  string
		flagName                 string
		flagCategory             string
		flagDescription          string
		flagCode                 string
		flagControlType          string
		flagNature               string
		flagOperatingMode        string
		flagOperatingFrequency   string
		flagOperatingEvent       string
		flagEvidenceCadence      string
		flagTestingCadence       string
		flagImplementationStatus string
		flagOwnerID              string
		flagReviewerID           string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an internal control",
		Example: `  # Create an internal control interactively
  prb internal-control create

  # Create an internal control non-interactively
  prb internal-control create --name "Enable encryption at rest" --category "Security"`,
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
						Title("Internal control name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagCategory == "" {
					err := huh.NewInput().
						Title("Internal control category").
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

			if err := fields.SetInternalControlFields(cmd, input, fields.InternalControlFieldFlags{
				Code:                 flagCode,
				ControlType:          flagControlType,
				Nature:               flagNature,
				OperatingMode:        flagOperatingMode,
				OperatingFrequency:   flagOperatingFrequency,
				OperatingEvent:       flagOperatingEvent,
				EvidenceCadence:      flagEvidenceCadence,
				TestingCadence:       flagTestingCadence,
				ImplementationStatus: flagImplementationStatus,
				OwnerID:              flagOwnerID,
				ReviewerID:           flagReviewerID,
			}); err != nil {
				return err
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

			m := resp.CreateInternalControl.InternalControlEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created internal control %s (%s)\n",
				m.ID,
				m.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Internal control name (required)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Internal control category (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Internal control description")
	fields.AddInternalControlFieldFlags(
		cmd,
		&flagCode,
		&flagControlType,
		&flagNature,
		&flagOperatingMode,
		&flagOperatingFrequency,
		&flagOperatingEvent,
		&flagEvidenceCadence,
		&flagTestingCadence,
		&flagImplementationStatus,
		&flagOwnerID,
		&flagReviewerID,
	)

	return cmd
}
