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
mutation TagCreate($input: CreateTagInput!) {
  createTag(input: $input) {
    tagEdge {
      node {
        id
        name
        color
      }
    }
  }
}
`

type createResponse struct {
	CreateTag struct {
		TagEdge struct {
			Node struct {
				ID    string  `json:"id"`
				Name  string  `json:"name"`
				Color *string `json:"color"`
			} `json:"node"`
		} `json:"tagEdge"`
	} `json:"createTag"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg   string
		flagName  string
		flagColor string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tag",
		Example: `  # Create a tag interactively
  prb tag create

  # Create a tag non-interactively
  prb tag create --org prborg_... --name Production --color "#FF0000"`,
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

				if flagName == "" {
					err := huh.NewInput().
						Title("Name").
						Value(&flagName).
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
				return fmt.Errorf("organization ID is required; pass --org or set a default with 'prb auth login'")
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
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

			t := resp.CreateTag.TagEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created tag %s (%s)\n",
				t.ID,
				t.Name,
			)

			return nil
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Tag name; unique within the organization")
	cmd.Flags().StringVar(&flagColor, "color", "", "Optional hex color (#RGB or #RRGGBB)")

	return cmd
}
