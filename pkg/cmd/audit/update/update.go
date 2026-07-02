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
mutation($input: UpdateAuditInput!) {
  updateAudit(input: $input) {
    audit {
      id
      name
      state
      validFrom
      validUntil
    }
  }
}
`

type updateResponse struct {
	UpdateAudit struct {
		Audit struct {
			ID         string  `json:"id"`
			Name       string  `json:"name"`
			State      string  `json:"state"`
			ValidFrom  *string `json:"validFrom"`
			ValidUntil *string `json:"validUntil"`
		} `json:"audit"`
	} `json:"updateAudit"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName                  string
		flagState                 string
		flagValidFrom             string
		flagValidUntil            string
		flagTrustCenterVisibility string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an audit",
		Args:  cobra.ExactArgs(1),
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

			input := map[string]any{
				"id": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}

			if cmd.Flags().Changed("state") {
				input["state"] = flagState
			}

			if cmd.Flags().Changed("valid-from") {
				input["validFrom"] = flagValidFrom
			}

			if cmd.Flags().Changed("valid-until") {
				input["validUntil"] = flagValidUntil
			}

			if cmd.Flags().Changed("trust-center-visibility") {
				input["trustCenterVisibility"] = flagTrustCenterVisibility
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
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

			a := resp.UpdateAudit.Audit
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated audit %s (%s)\n",
				a.ID,
				a.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Audit name")
	cmd.Flags().StringVar(&flagState, "state", "", "Audit state: NOT_STARTED, IN_PROGRESS, COMPLETED, REJECTED, OUTDATED")
	cmd.Flags().StringVar(&flagValidFrom, "valid-from", "", "Valid from date (e.g. 2026-01-01)")
	cmd.Flags().StringVar(&flagValidUntil, "valid-until", "", "Valid until date (e.g. 2026-12-31)")
	cmd.Flags().StringVar(&flagTrustCenterVisibility, "trust-center-visibility", "", "Trust center visibility: NONE, PRIVATE, PUBLIC")

	return cmd
}
