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
mutation($input: CreateTrackerPatternInput!) {
  createTrackerPattern(input: $input) {
    trackerPatternEdge {
      node {
        id
        pattern
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
	CreateTrackerPattern struct {
		TrackerPatternEdge struct {
			Node struct {
				ID          string `json:"id"`
				Pattern     string `json:"pattern"`
				DisplayName string `json:"displayName"`
			} `json:"node"`
		} `json:"trackerPatternEdge"`
	} `json:"createTrackerPattern"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCategoryID  string
		flagPattern     string
		flagMatchType   string
		flagDisplayName string
		flagDescription string
		flagMaxAge      int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tracker pattern",
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
				if flagPattern == "" {
					if err := huh.NewInput().Title("Tracker pattern").Value(&flagPattern).Run(); err != nil {
						return err
					}
				}
				if flagMatchType == "" {
					if err := huh.NewSelect[string]().
						Title("Match type").
						Options(
							huh.NewOption("Exact", "EXACT"),
							huh.NewOption("Prefix", "PREFIX"),
						).
						Value(&flagMatchType).Run(); err != nil {
						return err
					}
				}
				if flagDisplayName == "" {
					if err := huh.NewInput().Title("Display name").Value(&flagDisplayName).Run(); err != nil {
						return err
					}
				}
			}

			if flagPattern == "" {
				return fmt.Errorf("pattern is required; pass --pattern or run interactively")
			}
			if flagMatchType == "" {
				return fmt.Errorf("match-type is required; pass --match-type or run interactively")
			}
			if flagDisplayName == "" {
				return fmt.Errorf("display-name is required; pass --display-name or run interactively")
			}

			input := map[string]any{
				"cookieCategoryId": flagCategoryID,
				"pattern":          flagPattern,
				"matchType":        flagMatchType,
				"displayName":      flagDisplayName,
			}
			if flagDescription != "" {
				input["description"] = flagDescription
			}
			if cmd.Flags().Changed("max-age-seconds") {
				input["maxAgeSeconds"] = flagMaxAge
			}

			data, err := client.Do(createMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			p := resp.CreateTrackerPattern.TrackerPatternEdge.Node
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Created tracker pattern %s (%s)\n", p.ID, p.DisplayName)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagCategoryID, "category-id", "", "Cookie category ID (required)")
	_ = cmd.MarkFlagRequired("category-id")
	cmd.Flags().StringVar(&flagPattern, "pattern", "", "Tracker pattern (required)")
	cmd.Flags().StringVar(&flagMatchType, "match-type", "", "Match type: EXACT or PREFIX (required)")
	cmd.Flags().StringVar(&flagDisplayName, "display-name", "", "Display name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Description")
	cmd.Flags().IntVar(&flagMaxAge, "max-age-seconds", 0, "Maximum age in seconds")

	return cmd
}
