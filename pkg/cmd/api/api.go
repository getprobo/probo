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

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdAPI(f *cmdutil.Factory) *cobra.Command {
	var flagFields []string

	cmd := &cobra.Command{
		Use:   "api <query>",
		Short: "Make an authenticated GraphQL request",
		Long:  "Send a GraphQL query or mutation to the Probo API and print the response.",
		Example: `  # Run a query
  proboctl api 'query { viewer { id email } }'

  # Run a mutation with variables
  proboctl api 'mutation($input: CreateRiskInput!) { createRisk(input: $input) { riskEdge { node { id } } } }' \
    -f input='{"organizationId":"...","name":"Test","category":"Operational","treatment":"ACCEPTED","inherentLikelihood":3,"inherentImpact":3}'

  # Read query from stdin
  echo '{ viewer { id } }' | proboctl api -`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var query string
			if len(args) == 1 && args[0] != "-" {
				query = args[0]
			} else {
				if len(args) == 0 && f.IOStreams.IsStdinTTY() {
					return fmt.Errorf("query argument is required when not reading from stdin")
				}
				data, err := io.ReadAll(f.IOStreams.In)
				if err != nil {
					return fmt.Errorf("cannot read query from stdin: %w", err)
				}
				query = string(data)
			}

			if query == "" {
				return fmt.Errorf("query is required")
			}

			variables, err := parseFields(flagFields)
			if err != nil {
				return err
			}

			raw, err := client.DoRaw(query, variables)
			if err != nil {
				return err
			}

			var indented bytes.Buffer
			if err := json.Indent(&indented, raw, "", "  "); err != nil {
				_, _ = f.IOStreams.Out.Write(raw)
			} else {
				indented.WriteByte('\n')
				_, _ = indented.WriteTo(f.IOStreams.Out)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(
		&flagFields,
		"field",
		"f",
		nil,
		"GraphQL variable in key=value format",
	)

	return cmd
}

func parseFields(fields []string) (map[string]any, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	vars := make(map[string]any, len(fields))
	for _, f := range fields {
		key, value, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid field format %q: expected key=value", f)
		}

		var parsed any
		if err := json.Unmarshal([]byte(value), &parsed); err != nil {
			parsed = value
		}
		vars[key] = parsed
	}

	return vars, nil
}
