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

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

var (
	schemaEndpoints = map[string]string{
		"console": "/api/console/v1/graphql",
		"connect": "/api/connect/v1/graphql",
	}
)

func NewCmdAPI(f *cmdutil.Factory) *cobra.Command {
	var (
		flagFields []string
		flagSchema string
	)

	cmd := &cobra.Command{
		Use:   "api <query>",
		Short: "Make an authenticated GraphQL request",
		Long:  "Send a GraphQL query or mutation to the Probo API and print the response.",
		Example: `  # Run a query against the console schema (default)
  prb api 'query { viewer { id email } }'

  # Run a mutation with variables
  prb api 'mutation($input: CreateRiskInput!) { createRisk(input: $input) { riskEdge { node { id } } } }' \
    -f input='{"organizationId":"...","name":"Test","category":"Operational","treatment":"ACCEPTED","inherentLikelihood":3,"inherentImpact":3}'

  # Query the connect schema
  prb api --schema connect 'query { viewer { id email } }'

  # Read query from stdin
  echo '{ viewer { id } }' | prb api -`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, ok := schemaEndpoints[flagSchema]
			if !ok {
				return fmt.Errorf("unknown schema %q: expected console or connect", flagSchema)
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
				endpoint,
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

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

	cmd.Flags().StringVar(
		&flagSchema,
		"schema",
		"console",
		"GraphQL schema to query (console or connect)",
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
