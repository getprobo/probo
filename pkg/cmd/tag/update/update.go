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

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation TagUpdate($input: UpdateTagInput!) {
  updateTag(input: $input) {
    tag {
      id
      name
      color
    }
  }
}
`

type updateResponse struct {
	UpdateTag struct {
		Tag struct {
			ID    string  `json:"id"`
			Name  string  `json:"name"`
			Color *string `json:"color"`
		} `json:"tag"`
	} `json:"updateTag"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagID    string
		flagName  string
		flagColor string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a tag",
		Example: `  # Update a tag interactively
  prb tag update --id prbrtg_...

  # Update the name
  prb tag update --id prbrtg_... --name Staging

  # Remove the color
  prb tag update --id prbrtg_... --color ""`,
		Args: cobra.NoArgs,
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

			setName := cmd.Flags().Changed("name")
			setColor := cmd.Flags().Changed("color")

			if f.IOStreams.IsInteractive() {
				if flagID == "" {
					err := huh.NewInput().
						Title("Tag ID").
						Value(&flagID).
						Run()
					if err != nil {
						return err
					}
				}

				if !setName && !setColor {
					err := huh.NewForm(
						huh.NewGroup(
							huh.NewInput().
								Title("Name (leave empty to keep)").
								Value(&flagName),
							huh.NewInput().
								Title("Color (leave empty to keep)").
								Value(&flagColor),
						),
					).Run()
					if err != nil {
						return err
					}

					setName = flagName != ""
					setColor = flagColor != ""
				}
			}

			if flagID == "" {
				return fmt.Errorf("tag ID is required; pass --id or run interactively")
			}

			input := map[string]any{
				"id": flagID,
			}

			if setName {
				input["name"] = flagName
			}

			if setColor {
				if flagColor == "" {
					input["color"] = nil
				} else {
					input["color"] = flagColor
				}
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one of --name or --color must be specified")
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

			t := resp.UpdateTag.Tag
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated tag %s (%s)\n",
				t.ID,
				t.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagID, "id", "", "Tag ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Tag name; unique within the organization")
	cmd.Flags().StringVar(&flagColor, "color", "", "Hex color (#RGB or #RRGGBB); pass an empty value to clear it")

	return cmd
}
