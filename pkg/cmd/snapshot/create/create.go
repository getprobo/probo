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
mutation($input: CreateSnapshotInput!) {
  createSnapshot(input: $input) {
    snapshotEdge {
      node {
        id
        name
        type
      }
    }
  }
}
`

type createResponse struct {
	CreateSnapshot struct {
		SnapshotEdge struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"node"`
		} `json:"snapshotEdge"`
	} `json:"createSnapshot"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagName        string
		flagType        string
		flagDescription string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new snapshot",
		Example: `  # Create a snapshot interactively
  prb snapshot create

  # Create a snapshot non-interactively
  prb snapshot create --name "Q1 2026 Risks" --type RISKS`,
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
						Title("Snapshot name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagType == "" {
					err := huh.NewSelect[string]().
						Title("Snapshot type").
						Options(
							huh.NewOption("Risks", "RISKS"),
							huh.NewOption("Third Parties", "THIRD_PARTIES"),
							huh.NewOption("Assets", "ASSETS"),
							huh.NewOption("Findings", "FINDINGS"),
							huh.NewOption("Obligations", "OBLIGATIONS"),
							huh.NewOption("Processing Activities", "PROCESSING_ACTIVITIES"),
							huh.NewOption("Statements of Applicability", "STATEMENTS_OF_APPLICABILITY"),
						).
						Value(&flagType).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}
			if flagType == "" {
				return fmt.Errorf("type is required; pass --type or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
				"type":           flagType,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
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

			s := resp.CreateSnapshot.SnapshotEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created snapshot %s (%s)\n",
				s.ID,
				s.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Snapshot name (required)")
	cmd.Flags().StringVar(&flagType, "type", "", "Snapshot type: RISKS, VENDORS, ASSETS, FINDINGS, OBLIGATIONS, PROCESSING_ACTIVITIES, STATEMENTS_OF_APPLICABILITY (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Snapshot description")

	return cmd
}
