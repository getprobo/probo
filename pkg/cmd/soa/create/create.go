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

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateStateOfApplicabilityInput!) {
  createStateOfApplicability(input: $input) {
    stateOfApplicabilityEdge {
      node {
        id
        name
        createdAt
      }
    }
  }
}
`

type createResponse struct {
	CreateStateOfApplicability struct {
		StateOfApplicabilityEdge struct {
			Node struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				CreatedAt string `json:"createdAt"`
			} `json:"node"`
		} `json:"stateOfApplicabilityEdge"`
	} `json:"createStateOfApplicability"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg   string
		flagName  string
		flagOwner string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new statement of applicability",
		Example: `  # Create a statement of applicability
  prb soa create --name "ISO 27001 SoA" --owner PROFILE_ID`,
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
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
				"ownerId":        flagOwner,
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

			s := resp.CreateStateOfApplicability.StateOfApplicabilityEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created statement of applicability %s (%s)\n",
				s.ID,
				s.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Name (required)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID (required)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("owner")

	return cmd
}
