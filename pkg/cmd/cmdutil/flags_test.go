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

package cmdutil_test

import (
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func TestAddOutputFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr string
	}{
		{
			name: "defaults to table",
			args: []string{},
			want: cmdutil.OutputTable,
		},
		{
			name: "json shorthand",
			args: []string{"--json"},
			want: cmdutil.OutputJSON,
		},
		{
			name: "output json",
			args: []string{"--output", "json"},
			want: cmdutil.OutputJSON,
		},
		{
			name: "short output table",
			args: []string{"-o", "table"},
			want: cmdutil.OutputTable,
		},
		{
			name:    "json conflicts with table",
			args:    []string{"--json", "--output", "table"},
			wantErr: "--json cannot be combined with --output table",
		},
		{
			name:    "rejects unknown format",
			args:    []string{"--output", "yaml"},
			wantErr: `invalid --output value "yaml"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var output *string

			cmd := &cobra.Command{
				Use: "list",
				RunE: func(cmd *cobra.Command, args []string) error {
					return cmdutil.ValidateOutputFlag(output)
				},
			}
			output = cmdutil.AddOutputFlag(cmd)
			cmd.SetArgs(tt.args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			err := cmd.Execute()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, output)
			assert.Equal(t, tt.want, *output)
		})
	}
}

func TestAddOrderDirectionFlag(t *testing.T) {
	t.Parallel()

	t.Run("rejects an unknown direction before RunE", func(t *testing.T) {
		t.Parallel()

		var direction string

		ran := false
		cmd := &cobra.Command{
			Use: "list",
			RunE: func(cmd *cobra.Command, args []string) error {
				ran = true

				return nil
			},
		}
		cmdutil.AddOrderDirectionFlag(cmd, &direction, "DESC")
		cmd.SetArgs([]string{"--order-direction", "SIDEWAYS"})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		err := cmd.Execute()
		require.ErrorContains(t, err, `invalid --order-direction value "SIDEWAYS"`)
		assert.False(t, ran)
	})

	t.Run("accepts the default", func(t *testing.T) {
		t.Parallel()

		var direction string

		cmd := &cobra.Command{
			Use: "list",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		cmdutil.AddOrderDirectionFlag(cmd, &direction, "ASC")
		cmd.SetArgs([]string{})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		require.NoError(t, cmd.Execute())
		assert.Equal(t, "ASC", direction)
	})

	t.Run("keeps the output flag check", func(t *testing.T) {
		t.Parallel()

		var direction string

		cmd := &cobra.Command{
			Use: "list",
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
		}
		cmdutil.AddOrderDirectionFlag(cmd, &direction, "DESC")
		cmdutil.AddOutputFlag(cmd)
		cmd.SetArgs([]string{"--json", "--output", "table"})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		err := cmd.Execute()
		require.ErrorContains(t, err, "--json cannot be combined with --output table")
	})
}

func TestAddOrganizationFlag(t *testing.T) {
	t.Parallel()

	var org string

	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmdutil.AddOrganizationFlag(cmd, &org, "Organization ID")
	cmd.SetArgs([]string{"--organization", "org_alias"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "org_alias", org)

	cmd.SetArgs([]string{"--org", "org_short"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "org_short", org)
}

func TestResolveOrganization(t *testing.T) {
	t.Parallel()

	t.Run("flag wins", func(t *testing.T) {
		t.Parallel()

		got, err := cmdutil.ResolveOrganization("org_flag", "org_saved")
		require.NoError(t, err)
		assert.Equal(t, "org_flag", got)
	})

	t.Run("falls back to the saved organization", func(t *testing.T) {
		t.Parallel()

		got, err := cmdutil.ResolveOrganization("", "org_saved")
		require.NoError(t, err)
		assert.Equal(t, "org_saved", got)
	})

	t.Run("requires an organization", func(t *testing.T) {
		t.Parallel()

		_, err := cmdutil.ResolveOrganization("", "")
		require.Error(t, err)
		assert.EqualError(t, err, cmdutil.OrganizationRequiredMessage)
	})
}

func TestResolveOrganizationOrgOnly(t *testing.T) {
	t.Parallel()

	t.Run("flag wins", func(t *testing.T) {
		t.Parallel()

		got, err := cmdutil.ResolveOrganizationOrgOnly("org_flag", "org_saved")
		require.NoError(t, err)
		assert.Equal(t, "org_flag", got)
	})

	t.Run("falls back to the saved organization", func(t *testing.T) {
		t.Parallel()

		got, err := cmdutil.ResolveOrganizationOrgOnly("", "org_saved")
		require.NoError(t, err)
		assert.Equal(t, "org_saved", got)
	})

	t.Run("requires an organization without mentioning --organization", func(t *testing.T) {
		t.Parallel()

		_, err := cmdutil.ResolveOrganizationOrgOnly("", "")
		require.Error(t, err)
		assert.EqualError(t, err, cmdutil.OrganizationRequiredMessageOrgOnly)
		assert.NotContains(t, err.Error(), "--organization")
	})
}
