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
mutation($input: CreateAccessReviewCampaignInput!) {
  createAccessReviewCampaign(input: $input) {
    accessReviewCampaignEdge {
      node {
        id
        name
        status
      }
    }
  }
}
`

type createResponse struct {
	CreateAccessReviewCampaign struct {
		AccessReviewCampaignEdge struct {
			Node struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"node"`
		} `json:"accessReviewCampaignEdge"`
	} `json:"createAccessReviewCampaign"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg         string
		flagName        string
		flagDescription string
		flagSourceIDs   []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an access review campaign",
		Args:  cobra.NoArgs,
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
				return fmt.Errorf("cannot determine organization, use --org or 'prb auth login'")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}

			if len(flagSourceIDs) > 0 {
				input["accessSourceIds"] = flagSourceIDs
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

			c := resp.CreateAccessReviewCampaign.AccessReviewCampaignEdge.Node
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Created access review campaign %s\n", c.ID)
			_, _ = fmt.Fprintf(out, "Name: %s\n", c.Name)
			_, _ = fmt.Fprintf(out, "Status: %s\n", c.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Campaign name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Campaign description")
	cmd.Flags().StringSliceVar(
		&flagSourceIDs,
		"source-id",
		nil,
		"Access source IDs to include (can be repeated)",
	)

	_ = cmd.MarkFlagRequired("name")

	return cmd
}
