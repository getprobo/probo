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

package view

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Measure {
      id
      name
      description
      category
      state
      code
      controlType
      nature
      operatingFrequency {
        mode
        interval
        event
      }
      evidenceCadence
      testingCadence
      nextEvidenceDue
      nextTestDue
      implementationStatus
      owner {
        id
        fullName
      }
      reviewer {
        id
        fullName
      }
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename             string              `json:"__typename"`
		ID                   string              `json:"id"`
		Name                 string              `json:"name"`
		Description          *string             `json:"description"`
		Category             string              `json:"category"`
		State                string              `json:"state"`
		Code                 *string             `json:"code"`
		ControlType          *string             `json:"controlType"`
		Nature               *string             `json:"nature"`
		OperatingFrequency   *operatingFrequency `json:"operatingFrequency"`
		EvidenceCadence      *string             `json:"evidenceCadence"`
		TestingCadence       *string             `json:"testingCadence"`
		NextEvidenceDue      *string             `json:"nextEvidenceDue"`
		NextTestDue          *string             `json:"nextTestDue"`
		ImplementationStatus string              `json:"implementationStatus"`
		Owner                *profileRef         `json:"owner"`
		Reviewer             *profileRef         `json:"reviewer"`
		CreatedAt            string              `json:"createdAt"`
		UpdatedAt            string              `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View an internal control",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
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
				viewQuery,
				map[string]any{"id": args[0]},
			)
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("measure %s not found", args[0])
			}

			if resp.Node.Typename != "Measure" {
				return fmt.Errorf("expected Measure node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			m := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(m.Name))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), m.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Code:"), deref(m.Code))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Category:"), m.Category)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Status:"), m.ImplementationStatus)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("State:"), m.State)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Type:"), deref(m.ControlType))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Nature:"), deref(m.Nature))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Operating:"), formatOperatingFrequency(m.OperatingFrequency))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Evidence cadence:"), deref(m.EvidenceCadence))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Testing cadence:"), deref(m.TestingCadence))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Evidence due:"), formatOptionalTime(m.NextEvidenceDue))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Test due:"), formatOptionalTime(m.NextTestDue))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Owner:"), profileName(m.Owner))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Reviewer:"), profileName(m.Reviewer))

			if m.Description != nil && *m.Description != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), *m.Description)
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(m.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(m.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

type profileRef struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
}

func profileName(profile *profileRef) string {
	if profile == nil {
		return ""
	}

	return profile.FullName
}

func formatOptionalTime(value *string) string {
	if value == nil || *value == "" {
		return ""
	}

	return cmdutil.FormatTime(*value)
}

type operatingFrequency struct {
	Mode     string  `json:"mode"`
	Interval *string `json:"interval"`
	Event    *string `json:"event"`
}

func formatOperatingFrequency(freq *operatingFrequency) string {
	if freq == nil {
		return ""
	}

	switch freq.Mode {
	case "CONTINUOUS":
		return "continuous"
	case "EVENT":
		if freq.Event != nil && *freq.Event != "" {
			return *freq.Event
		}

		return "event"
	case "PERIODIC":
		if freq.Interval != nil {
			return *freq.Interval
		}

		return freq.Mode
	default:
		return freq.Mode
	}
}

func deref(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
