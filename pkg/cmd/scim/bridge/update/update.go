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

const updateMutation = `
mutation($input: UpdateSCIMBridgeInput!) {
  updateSCIMBridge(input: $input) {
    scimBridge {
      id
      state
      excludedUserNames
    }
  }
}
`

type updateResponse struct {
	UpdateSCIMBridge struct {
		ScimBridge struct {
			ID                string   `json:"id"`
			State             string   `json:"state"`
			ExcludedUserNames []string `json:"excludedUserNames"`
		} `json:"scimBridge"`
	} `json:"updateSCIMBridge"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg               string
		flagExcludedUserNames []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a SCIM bridge",
		Example: `  # Set excluded user names
  prb scim bridge update <id> --excluded-user-names admin@example.com --excluded-user-names bot@example.com

  # Clear excluded user names
  prb scim bridge update <id> --excluded-user-names ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("excluded-user-names") {
				return fmt.Errorf("at least one field must be specified for update; use --excluded-user-names")
			}

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
				"/api/connect/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			excluded := make([]string, 0, len(flagExcludedUserNames))
			for _, name := range flagExcludedUserNames {
				if name != "" {
					excluded = append(excluded, name)
				}
			}

			data, err := client.Do(
				updateMutation,
				map[string]any{
					"input": map[string]any{
						"organizationId":    flagOrg,
						"scimBridgeId":      args[0],
						"excludedUserNames": excluded,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated SCIM bridge %s\n",
				resp.UpdateSCIMBridge.ScimBridge.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringSliceVar(&flagExcludedUserNames, "excluded-user-names", nil, "User names to exclude from SCIM sync (repeatable)")

	return cmd
}
