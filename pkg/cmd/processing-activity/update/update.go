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
mutation($input: UpdateProcessingActivityInput!) {
  updateProcessingActivity(input: $input) {
    processingActivity {
      id
      name
      role
      lawfulBasis
    }
  }
}
`

type updateResponse struct {
	UpdateProcessingActivity struct {
		ProcessingActivity struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Role        string `json:"role"`
			LawfulBasis string `json:"lawfulBasis"`
		} `json:"processingActivity"`
	} `json:"updateProcessingActivity"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName        string
		flagPurpose     string
		flagRole        string
		flagLawfulBasis string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a processing activity",
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
			if cmd.Flags().Changed("purpose") {
				input["purpose"] = flagPurpose
			}
			if cmd.Flags().Changed("role") {
				input["role"] = flagRole
			}
			if cmd.Flags().Changed("lawful-basis") {
				input["lawfulBasis"] = flagLawfulBasis
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

			a := resp.UpdateProcessingActivity.ProcessingActivity
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated processing activity %s (%s)\n",
				a.ID,
				a.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Processing activity name")
	cmd.Flags().StringVar(&flagPurpose, "purpose", "", "Purpose of processing")
	cmd.Flags().StringVar(&flagRole, "role", "", "Role: CONTROLLER, PROCESSOR")
	cmd.Flags().StringVar(&flagLawfulBasis, "lawful-basis", "", "Lawful basis: LEGITIMATE_INTEREST, CONSENT, CONTRACTUAL_NECESSITY, LEGAL_OBLIGATION, VITAL_INTERESTS, PUBLIC_TASK")

	return cmd
}
