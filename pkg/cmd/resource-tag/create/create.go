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

const createMutation = `
mutation($input: CreateResourceTagInput!) {
  createResourceTag(input: $input) {
    resourceTagEdge {
      node {
        id
        key
        value
        color
      }
    }
  }
}
`

type createResponse struct {
	CreateResourceTag struct {
		ResourceTagEdge struct {
			Node struct {
				ID    string  `json:"id"`
				Key   string  `json:"key"`
				Value string  `json:"value"`
				Color *string `json:"color"`
			} `json:"node"`
		} `json:"resourceTagEdge"`
	} `json:"createResourceTag"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg   string
		flagKey   string
		flagValue string
		flagColor string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource tag",
		Example: `  # Create a tag interactively
  prb resource-tag create

  # Create a tag non-interactively
  prb resource-tag create --organization-id prborg_... --key environment --value production --color "#FF0000"`,
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

			if f.IOStreams.IsInteractive() {
				if flagOrg == "" {
					err := huh.NewInput().
						Title("Organization ID").
						Value(&flagOrg).
						Run()
					if err != nil {
						return err
					}
				}

				if flagKey == "" {
					err := huh.NewInput().
						Title("Key").
						Value(&flagKey).
						Run()
					if err != nil {
						return err
					}
				}

				if flagValue == "" {
					err := huh.NewInput().
						Title("Value").
						Value(&flagValue).
						Run()
					if err != nil {
						return err
					}
				}

				if flagColor == "" {
					err := huh.NewInput().
						Title("Color (optional, hex)").
						Value(&flagColor).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagOrg == "" {
				return fmt.Errorf("organization ID is required; pass --organization-id or set a default with 'prb auth login'")
			}

			if flagKey == "" {
				return fmt.Errorf("key is required; pass --key or run interactively")
			}

			if flagValue == "" {
				return fmt.Errorf("value is required; pass --value or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"key":            flagKey,
				"value":          flagValue,
			}

			if flagColor != "" {
				input["color"] = flagColor
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

			t := resp.CreateResourceTag.ResourceTagEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created resource tag %s (%s=%s)\n",
				t.ID,
				t.Key,
				t.Value,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "organization-id", "", "Organization ID")
	cmd.Flags().StringVar(&flagKey, "key", "", "Unique slug key for the tag")
	cmd.Flags().StringVar(&flagValue, "value", "", "Display value for the tag")
	cmd.Flags().StringVar(&flagColor, "color", "", "Optional hex color (#RGB or #RRGGBB)")

	return cmd
}
