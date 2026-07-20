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

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

var validIcons = []string{
	"LOCK_KEY",
	"EYE_SLASH",
	"FINGERPRINT",
	"SHIELD_WARNING",
	"SHIELD_CHECK",
	"SIREN",
	"KEY",
	"LOCK",
	"CLOUD",
	"DATABASE",
	"GLOBE",
	"EYE",
	"USERS",
	"CERTIFICATE",
	"GAVEL",
	"HEARTBEAT",
	"BELL",
	"BUG",
	"CODE",
	"SERVER",
}

const createMutation = `
mutation($input: CreateCompliancePortalCommitmentInput!) {
  createCompliancePortalCommitment(input: $input) {
    compliancePortalCommitmentEdge {
      node {
        id
        icon
        eyebrow
        title
        description
        rank
      }
    }
  }
}
`

type createResponse struct {
	CreateCompliancePortalCommitment struct {
		CompliancePortalCommitmentEdge struct {
			Node struct {
				ID          string `json:"id"`
				Icon        string `json:"icon"`
				Eyebrow     string `json:"eyebrow"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Rank        int    `json:"rank"`
			} `json:"node"`
		} `json:"compliancePortalCommitmentEdge"`
	} `json:"createCompliancePortalCommitment"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagGroup       string
		flagIcon        string
		flagEyebrow     string
		flagTitle       string
		flagDescription string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a compliance portal commitment",
		Example: `  # Create a commitment interactively
  prb trust-center commitment create --group <group-id>

  # Create a commitment non-interactively
  prb trust-center cmt create --group <group-id> --icon SHIELD_CHECK --eyebrow "Security" --title "Encryption" --description "Data encrypted at rest"`,
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

			if f.IOStreams.IsInteractive() {
				if flagGroup == "" {
					err := huh.NewInput().
						Title("Commitment group ID").
						Value(&flagGroup).
						Run()
					if err != nil {
						return err
					}
				}

				if flagIcon == "" {
					iconOptions := make([]huh.Option[string], 0, len(validIcons))
					for _, icon := range validIcons {
						iconOptions = append(iconOptions, huh.NewOption(icon, icon))
					}

					err := huh.NewSelect[string]().
						Title("Icon").
						Options(iconOptions...).
						Value(&flagIcon).
						Run()
					if err != nil {
						return err
					}
				}

				if flagEyebrow == "" {
					err := huh.NewInput().
						Title("Eyebrow").
						Value(&flagEyebrow).
						Run()
					if err != nil {
						return err
					}
				}

				if flagTitle == "" {
					err := huh.NewInput().
						Title("Title").
						Value(&flagTitle).
						Run()
					if err != nil {
						return err
					}
				}

				if flagDescription == "" {
					err := huh.NewText().
						Title("Description").
						Value(&flagDescription).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagGroup == "" {
				return fmt.Errorf("group is required; pass --group or run interactively")
			}

			if flagIcon == "" {
				return fmt.Errorf("icon is required; pass --icon or run interactively")
			}

			if err := cmdutil.ValidateEnum("icon", flagIcon, validIcons); err != nil {
				return err
			}

			if flagEyebrow == "" {
				return fmt.Errorf("eyebrow is required; pass --eyebrow or run interactively")
			}

			if flagTitle == "" {
				return fmt.Errorf("title is required; pass --title or run interactively")
			}

			if flagDescription == "" {
				return fmt.Errorf("description is required; pass --description or run interactively")
			}

			data, err := client.Do(
				createMutation,
				map[string]any{
					"input": map[string]any{
						"groupId":     flagGroup,
						"icon":        flagIcon,
						"eyebrow":     flagEyebrow,
						"title":       flagTitle,
						"description": flagDescription,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			c := resp.CreateCompliancePortalCommitment.CompliancePortalCommitmentEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created commitment %s (%s)\n",
				c.ID,
				c.Title,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagGroup, "group", "", "Commitment group ID (required)")
	cmd.Flags().StringVar(&flagIcon, "icon", "", "Commitment icon (required)")
	cmd.Flags().StringVar(&flagEyebrow, "eyebrow", "", "Commitment eyebrow (required)")
	cmd.Flags().StringVar(&flagTitle, "title", "", "Commitment title (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Commitment description (required)")

	return cmd
}
