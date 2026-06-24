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
		flagName        string
		flagDescription string
		flagOutput      *string
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
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
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
