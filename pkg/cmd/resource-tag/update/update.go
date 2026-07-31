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
mutation($input: UpdateResourceTagInput!) {
  updateResourceTag(input: $input) {
    resourceTag {
      id
      key
      value
      color
    }
  }
}
`

type updateResponse struct {
	UpdateResourceTag struct {
		ResourceTag struct {
			ID    string  `json:"id"`
			Key   string  `json:"key"`
			Value string  `json:"value"`
			Color *string `json:"color"`
		} `json:"resourceTag"`
	} `json:"updateResourceTag"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagID    string
		flagKey   string
		flagValue string
		flagColor string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a resource tag",
		Example: `  # Update a tag interactively
  prb resource-tag update --id prbrtg_...

  # Update key and value
  prb resource-tag update --id prbrtg_... --key environment --value staging`,
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
				if flagID == "" {
					err := huh.NewInput().
						Title("Resource tag ID").
						Value(&flagID).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagID == "" {
				return fmt.Errorf("resource tag ID is required; pass --id or run interactively")
			}

			input := map[string]any{
				"id": flagID,
			}

			if cmd.Flags().Changed("key") {
				input["key"] = flagKey
			}

			if cmd.Flags().Changed("value") {
				input["value"] = flagValue
			}

			if cmd.Flags().Changed("color") {
				input["color"] = flagColor
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one of --key, --value, or --color must be specified")
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

			t := resp.UpdateResourceTag.ResourceTag
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated resource tag %s (%s=%s)\n",
				t.ID,
				t.Key,
				t.Value,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagID, "id", "", "Resource tag ID")
	cmd.Flags().StringVar(&flagKey, "key", "", "Unique slug key for the tag")
	cmd.Flags().StringVar(&flagValue, "value", "", "Display value for the tag")
	cmd.Flags().StringVar(&flagColor, "color", "", "Optional hex color (#RGB or #RRGGBB)")

	return cmd
}
