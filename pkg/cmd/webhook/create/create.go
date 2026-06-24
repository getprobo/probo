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
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/webhook/shared"
)

const createMutation = `
mutation($input: CreateWebhookSubscriptionInput!) {
  createWebhookSubscription(input: $input) {
    webhookSubscriptionEdge {
      node {
        id
        endpointUrl
        selectedEvents
      }
    }
  }
}
`

type createResponse struct {
	CreateWebhookSubscription struct {
		WebhookSubscriptionEdge struct {
			Node struct {
				ID             string   `json:"id"`
				EndpointURL    string   `json:"endpointUrl"`
				SelectedEvents []string `json:"selectedEvents"`
			} `json:"node"`
		} `json:"webhookSubscriptionEdge"`
	} `json:"createWebhookSubscription"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagURL    string
		flagEvents []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook subscription",
		Example: `  # Create a webhook for thirdParty events
  prb webhook create --url https://example.com/webhook --event THIRD_PARTY_CREATED --event THIRD_PARTY_UPDATED

  # Create a webhook for all supported events
  prb webhook create --url https://example.com/webhook --event THIRD_PARTY_CREATED --event THIRD_PARTY_UPDATED --event THIRD_PARTY_DELETED --event USER_CREATED --event USER_UPDATED --event USER_DELETED --event OBLIGATION_CREATED --event OBLIGATION_UPDATED --event OBLIGATION_DELETED`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, e := range flagEvents {
				valid := slices.Contains(shared.ValidEvents, e)
				if !valid {
					return fmt.Errorf("invalid --event value %q: valid values are %s", e, strings.Join(shared.ValidEvents, ", "))
				}
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"endpointUrl":    flagURL,
				"selectedEvents": flagEvents,
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

			w := resp.CreateWebhookSubscription.WebhookSubscriptionEdge.Node
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "Created webhook subscription %s\n", w.ID)
			_, _ = fmt.Fprintf(out, "Endpoint: %s\n", w.EndpointURL)
			_, _ = fmt.Fprintf(out, "Events: %s\n", strings.Join(w.SelectedEvents, ", "))

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagURL, "url", "", "Webhook endpoint URL (required)")
	cmd.Flags().StringSliceVar(&flagEvents, "event", nil, "Event types to subscribe to (required, can be repeated)")

	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("event")

	return cmd
}
