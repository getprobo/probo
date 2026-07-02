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
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/webhook/shared"
)

const updateMutation = `
mutation($input: UpdateWebhookSubscriptionInput!) {
  updateWebhookSubscription(input: $input) {
    webhookSubscription {
      id
      endpointUrl
      selectedEvents
    }
  }
}
`

type updateResponse struct {
	UpdateWebhookSubscription struct {
		WebhookSubscription struct {
			ID             string   `json:"id"`
			EndpointURL    string   `json:"endpointUrl"`
			SelectedEvents []string `json:"selectedEvents"`
		} `json:"webhookSubscription"`
	} `json:"updateWebhookSubscription"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagURL    string
		flagEvents []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a webhook subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("url") {
				input["endpointUrl"] = flagURL
			}

			if cmd.Flags().Changed("event") {
				for _, e := range flagEvents {
					valid := slices.Contains(shared.ValidEvents, e)
					if !valid {
						return fmt.Errorf("invalid --event value %q: valid values are %s", e, strings.Join(shared.ValidEvents, ", "))
					}
				}

				input["selectedEvents"] = flagEvents
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one of --url or --event must be specified")
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

			w := resp.UpdateWebhookSubscription.WebhookSubscription
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated webhook subscription %s\n",
				w.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagURL, "url", "", "Webhook endpoint URL")
	cmd.Flags().StringSliceVar(&flagEvents, "event", nil, "Event types to subscribe to (replaces existing)")

	return cmd
}
