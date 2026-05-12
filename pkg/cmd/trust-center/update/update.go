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

package update

import (
	"encoding/json"
	"fmt"

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

const updateMutation = `
mutation($input: UpdateTrustCenterInput!) {
  updateTrustCenter(input: $input) {
    trustCenter {
      id
      active
      searchEngineIndexing
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

type updateResponse struct {
	UpdateTrustCenter struct {
		TrustCenter struct {
			ID                   string `json:"id"`
			Active               bool   `json:"active"`
			SearchEngineIndexing string `json:"searchEngineIndexing"`
		} `json:"trustCenter"`
	} `json:"updateTrustCenter"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                  string
		flagActive               bool
		flagSearchEngineIndexing string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update trust center settings",
		Example: `  # Enable the trust center
  prb trust-center update --active

  # Disable search engine indexing
  prb trust-center update --search-engine-indexing NOT_INDEXABLE`,
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

			input := map[string]any{
				"trustCenterId": tcResp.Node.TrustCenter.ID,
			}

			if cmd.Flags().Changed("active") {
				input["active"] = flagActive
			}
			if cmd.Flags().Changed("search-engine-indexing") {
				if err := cmdutil.ValidateEnum("search-engine-indexing", flagSearchEngineIndexing, []string{"INDEXABLE", "NOT_INDEXABLE"}); err != nil {
					return err
				}
				input["searchEngineIndexing"] = flagSearchEngineIndexing
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			data, err = client.Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			tc := resp.UpdateTrustCenter.TrustCenter
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated trust center %s\n",
				tc.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().BoolVar(&flagActive, "active", false, "Enable or disable the trust center")
	cmd.Flags().StringVar(&flagSearchEngineIndexing, "search-engine-indexing", "", "Search engine indexing: INDEXABLE, NOT_INDEXABLE")

	return cmd
}
