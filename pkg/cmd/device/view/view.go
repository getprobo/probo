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
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/device/shared"
)

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Device {
      id
      state
      hostname
      platform
      osVersion
      agentVersion
      serialNumber
      hardwareUuid
      enrolledAt
      lastSeenAt
      revokedAt
      createdAt
      updatedAt
      owner {
        id
      }
      latestPostures {
        id
        checkKey
        status
        value {
          kind
          text
          number
        }
        observedAt
      }
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename     string  `json:"__typename"`
		ID           string  `json:"id"`
		State        string  `json:"state"`
		Hostname     *string `json:"hostname"`
		Platform     *string `json:"platform"`
		OsVersion    *string `json:"osVersion"`
		AgentVersion *string `json:"agentVersion"`
		SerialNumber *string `json:"serialNumber"`
		HardwareUUID *string `json:"hardwareUuid"`
		EnrolledAt   *string `json:"enrolledAt"`
		LastSeenAt   *string `json:"lastSeenAt"`
		RevokedAt    *string `json:"revokedAt"`
		CreatedAt    string  `json:"createdAt"`
		UpdatedAt    string  `json:"updatedAt"`
		Owner        *struct {
			ID string `json:"id"`
		} `json:"owner"`
		LatestPostures []shared.Posture `json:"latestPostures"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View an ITAM device",
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
				return fmt.Errorf("device %s not found", args[0])
			}

			if resp.Node.Typename != "Device" {
				return fmt.Errorf("expected Device node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			d := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			title := d.ID
			if d.Hostname != nil && *d.Hostname != "" {
				title = *d.Hostname
			}

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(title))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), d.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("State:"), d.State)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Hostname:"), ref.UnrefOrZero(d.Hostname))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Platform:"), ref.UnrefOrZero(d.Platform))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("OS Version:"), ref.UnrefOrZero(d.OsVersion))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Agent Version:"), ref.UnrefOrZero(d.AgentVersion))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Serial Number:"), ref.UnrefOrZero(d.SerialNumber))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Hardware UUID:"), ref.UnrefOrZero(d.HardwareUUID))

			if d.Owner != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Owner:"), d.Owner.ID)
			} else {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Owner:"), "")
			}

			if d.EnrolledAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Enrolled:"), cmdutil.FormatTime(*d.EnrolledAt))
			}

			if d.LastSeenAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Last Seen:"), cmdutil.FormatTime(*d.LastSeenAt))
			}

			if d.RevokedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Revoked:"), cmdutil.FormatTime(*d.RevokedAt))
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(d.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(d.UpdatedAt))

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s\n", bold.Render("Latest Postures"))

			if len(d.LatestPostures) == 0 {
				_, _ = fmt.Fprintln(out, "No postures recorded.")
				return nil
			}

			rows := make([][]string, 0, len(d.LatestPostures))
			for _, p := range d.LatestPostures {
				value := p.Value.Kind
				if p.Value.Text != "" {
					value = p.Value.Text
				} else if p.Value.Number != nil {
					value = fmt.Sprintf("%s (%d)", p.Value.Kind, *p.Value.Number)
				}

				rows = append(rows, []string{
					p.CheckKey,
					p.Status,
					value,
					cmdutil.FormatTime(p.ObservedAt),
				})
			}

			t := cmdutil.NewTable("CHECK", "STATUS", "VALUE", "OBSERVED").Rows(rows...)
			_, _ = fmt.Fprintln(out, t)

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
