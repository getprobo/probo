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

package deactivate

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const deactivateMutation = `
mutation($input: DeactivateUserInput!) {
  deactivateUser(input: $input) {
    profile {
      id
      fullName
      state
    }
  }
}
`

type deactivateResponse struct {
	DeactivateUser *struct {
		Profile *struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
			State    string `json:"state"`
		} `json:"profile"`
	} `json:"deactivateUser"`
}

func NewCmdDeactivate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "deactivate <id>",
		Short: "Mark a user as inactive without deleting them",
		Args:  cobra.ExactArgs(1),
		Example: `  # Deactivate a user in the default organization
  prb user deactivate prfl_01H...

  # Deactivate a user in a specific organization
  prb user deactivate prfl_01H... --org org_01H...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/connect/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			data, err := client.Do(
				deactivateMutation,
				map[string]any{
					"input": map[string]any{
						"organizationId": flagOrg,
						"profileId":      args[0],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp deactivateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.DeactivateUser == nil || resp.DeactivateUser.Profile == nil {
				return fmt.Errorf("user %s not found", args[0])
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.DeactivateUser.Profile)
			}

			profile := resp.DeactivateUser.Profile
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Deactivated user %s (%s) — state: %s\n",
				profile.FullName,
				profile.ID,
				profile.State,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
