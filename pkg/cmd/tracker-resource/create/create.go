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

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateTrackerResourceInput!) {
  createTrackerResource(input: $input) {
    trackerResourceEdge {
      node {
        id
        displayName
      }
    }
    cookieBanner {
      id
    }
  }
}
`

type createResponse struct {
	CreateTrackerResource struct {
		TrackerResourceEdge struct {
			Node struct {
				ID          string `json:"id"`
				DisplayName string `json:"displayName"`
			} `json:"node"`
		} `json:"trackerResourceEdge"`
	} `json:"createTrackerResource"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCategoryID   string
		flagResourceType string
		flagOrigin       string
		flagPath         string
		flagDisplayName  string
		flagDescription  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tracker resource",
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
				if flagResourceType == "" {
					if err := huh.NewSelect[string]().
						Title("Resource type").
						Options(
							huh.NewOption("Script", "SCRIPT"),
							huh.NewOption("Iframe", "IFRAME"),
							huh.NewOption("Image", "IMAGE"),
							huh.NewOption("Stylesheet", "STYLESHEET"),
							huh.NewOption("Font", "FONT"),
							huh.NewOption("Beacon", "BEACON"),
							huh.NewOption("Fetch / XHR", "FETCH"),
							huh.NewOption("Media", "MEDIA"),
						).
						Value(&flagResourceType).Run(); err != nil {
						return err
					}
				}
				if flagOrigin == "" {
					if err := huh.NewInput().Title("Origin").Value(&flagOrigin).Run(); err != nil {
						return err
					}
				}
				if flagPath == "" {
					if err := huh.NewInput().Title("Path").Value(&flagPath).Run(); err != nil {
						return err
					}
				}
				if flagDisplayName == "" {
					if err := huh.NewInput().Title("Display name").Value(&flagDisplayName).Run(); err != nil {
						return err
					}
				}
			}

			if flagResourceType == "" {
				return fmt.Errorf("resource-type is required; pass --resource-type or run interactively")
			}
			if flagOrigin == "" {
				return fmt.Errorf("origin is required; pass --origin or run interactively")
			}
			if flagPath == "" {
				return fmt.Errorf("path is required; pass --path or run interactively")
			}
			if flagDisplayName == "" {
				return fmt.Errorf("display-name is required; pass --display-name or run interactively")
			}

			input := map[string]any{
				"cookieCategoryId": flagCategoryID,
				"type":             flagResourceType,
				"origin":           flagOrigin,
				"path":             flagPath,
				"displayName":      flagDisplayName,
			}
			if flagDescription != "" {
				input["description"] = flagDescription
			}

			data, err := client.Do(createMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			r := resp.CreateTrackerResource.TrackerResourceEdge.Node
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Created tracker resource %s (%s)\n", r.ID, r.DisplayName)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagCategoryID, "category-id", "", "Cookie category ID (required)")
	_ = cmd.MarkFlagRequired("category-id")
	cmd.Flags().StringVar(&flagResourceType, "resource-type", "", "Resource type: SCRIPT, IFRAME, IMAGE, STYLESHEET, FONT, BEACON, FETCH or MEDIA (required)")
	cmd.Flags().StringVar(&flagOrigin, "origin", "", "Origin URL (required)")
	cmd.Flags().StringVar(&flagPath, "path", "", "Resource path (required)")
	cmd.Flags().StringVar(&flagDisplayName, "display-name", "", "Display name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Description")

	return cmd
}
