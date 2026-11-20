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

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query ThirdPartyAgreementView($id: ID!) {
  node(id: $id) {
    __typename
    ... on ThirdPartyAgreement {
      id
      type
      validity {
        start
        end
      }
      file {
        id
        fileName
      }
      thirdParty {
        id
        name
      }
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename   string  `json:"__typename"`
		ID         string  `json:"id"`
		Type       string  `json:"type"`
		Validity   *period `json:"validity"`
		File       file    `json:"file"`
		ThirdParty struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"thirdParty"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"node"`
}

type period struct {
	Start *string `json:"start"`
	End   *string `json:"end"`
}

type file struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a third-party agreement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			data, err := client.Do(viewQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil || resp.Node.Typename != "ThirdPartyAgreement" {
				return fmt.Errorf("agreement %s not found", args[0])
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			agreement := resp.Node
			out := f.IOStreams.Out
			_, _ = fmt.Fprintf(out, "ID:           %s\n", agreement.ID)
			_, _ = fmt.Fprintf(out, "Type:         %s\n", agreement.Type)
			_, _ = fmt.Fprintf(out, "Third party:  %s (%s)\n", agreement.ThirdParty.Name, agreement.ThirdParty.ID)
			_, _ = fmt.Fprintf(out, "File:         %s\n", agreement.File.FileName)
			_, _ = fmt.Fprintf(out, "Valid from:   %s\n", formatOptionalTime(agreement.Validity, true))
			_, _ = fmt.Fprintf(out, "Valid until:  %s\n", formatOptionalTime(agreement.Validity, false))
			_, _ = fmt.Fprintf(out, "Created:      %s\n", cmdutil.FormatTime(agreement.CreatedAt))
			_, _ = fmt.Fprintf(out, "Updated:      %s\n", cmdutil.FormatTime(agreement.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func formatOptionalTime(validity *period, start bool) string {
	if validity == nil {
		return ""
	}

	value := validity.End
	if start {
		value = validity.Start
	}

	if value == nil || *value == "" {
		return ""
	}

	return cmdutil.FormatTime(*value)
}
