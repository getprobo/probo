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
		Example: `  # Create a webhook for vendor events
  prb webhook create --url https://example.com/webhook --event VENDOR_CREATED --event VENDOR_UPDATED

  # Create a webhook for all supported events
  prb webhook create --url https://example.com/webhook --event VENDOR_CREATED --event VENDOR_UPDATED --event VENDOR_DELETED --event USER_CREATED --event USER_UPDATED --event USER_DELETED --event OBLIGATION_CREATED --event OBLIGATION_UPDATED --event OBLIGATION_DELETED`,
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
