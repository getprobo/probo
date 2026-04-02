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
mutation($input: UpdateAccessReviewCampaignInput!) {
  updateAccessReviewCampaign(input: $input) {
    accessReviewCampaign {
      id
      name
      status
    }
  }
}
`

type updateResponse struct {
	UpdateAccessReviewCampaign struct {
		AccessReviewCampaign struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"accessReviewCampaign"`
	} `json:"updateAccessReviewCampaign"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName             string
		flagDescription      string
		flagFrameworkControl []string
		flagOutput           *string
	)

	cmd := &cobra.Command{
		Use:   "update <campaign-id>",
		Short: "Update an access review campaign",
		Args:  cobra.ExactArgs(1),
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

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
			)

			input := map[string]any{
				"accessReviewCampaignId": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}

			if cmd.Flags().Changed("description") {
				input["description"] = flagDescription
			}

			if cmd.Flags().Changed("framework-control") {
				input["frameworkControls"] = flagFrameworkControl
			}

			data, err := client.Do(
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

			c := resp.UpdateAccessReviewCampaign.AccessReviewCampaign

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, c)
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Updated access review campaign %s\n", c.ID)
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Name: %s\n", c.Name)
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Status: %s\n", c.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Campaign name")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Campaign description")
	cmd.Flags().StringSliceVar(
		&flagFrameworkControl,
		"framework-control",
		nil,
		"Framework control IDs (can be repeated)",
	)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
