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

package root_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
	"go.probo.inc/probo/pkg/cmd/root"
	"go.probo.inc/probo/pkg/gid"
)

func TestAcceptDashPrefixedIDs(t *testing.T) {
	t.Parallel()

	dash := gidWithLead(t, 0xF8, 0x00)
	plain := gidWithLead(t, 0x00, 0x00)
	long := gidWithLead(t, 0xFB, 0xE0)

	require.True(t, strings.HasPrefix(dash, "-") && !strings.HasPrefix(dash, "--"))
	require.False(t, strings.HasPrefix(plain, "-"))
	require.True(t, strings.HasPrefix(long, "--"))

	// The id from the failing e2e run. It starts with the shorthand '-u'.
	const reported = "-uF05HEMAGsAAQAAAaGMWaDZxr42mOTc"

	_, err := gid.ParseGID(reported)
	require.NoError(t, err)

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "view id before flags",
			args: []string{"framework", "view", dash, "--json"},
			want: []string{"framework", "view", "--json", "--", dash},
		},
		{
			name: "reported e2e id",
			args: []string{"framework", "view", reported, "--json"},
			want: []string{"framework", "view", "--json", "--", reported},
		},
		{
			name: "id that looks like a long flag",
			args: []string{"device", "view", long, "--json"},
			want: []string{"device", "view", "--json", "--", long},
		},
		{
			name: "shorthand flag after the id",
			args: []string{"framework", "delete", dash, "-y"},
			want: []string{"framework", "delete", "-y", "--", dash},
		},
		{
			name: "flag value stays attached",
			args: []string{"framework", "view", dash, "-o", "json"},
			want: []string{"framework", "view", "-o", "json", "--", dash},
		},
		{
			name: "update field stays attached",
			args: []string{"framework", "update", dash, "--name", "renamed"},
			want: []string{"framework", "update", "--name", "renamed", "--", dash},
		},
		{
			name: "ordinary id is unchanged",
			args: []string{"framework", "view", plain, "--json"},
			want: []string{"framework", "view", plain, "--json"},
		},
		{
			name: "organization flag keeps a dash id",
			args: []string{"framework", "list", "--organization", dash, "--json"},
			want: []string{"framework", "list", "--organization", dash, "--json"},
		},
		{
			name: "end of flags is preserved",
			args: []string{"framework", "view", "--", dash, "--json"},
			want: []string{"framework", "view", "--", dash, "--json"},
		},
		{
			name: "unknown shorthand is left for pflag",
			args: []string{"framework", "view", "-z"},
			want: []string{"framework", "view", "-z"},
		},
		{
			name: "persistent flag before the command",
			args: []string{"--no-color", "framework", "view", dash, "--json"},
			want: []string{"--no-color", "framework", "view", "--json", "--", dash},
		},
		{
			name: "persistent flag between commands",
			args: []string{"framework", "--no-interactive", "view", dash},
			want: []string{"framework", "--no-interactive", "view", "--", dash},
		},
		{
			name: "help and output shorthand cluster",
			args: []string{"framework", "view", dash, "-ho", "json"},
			want: []string{"framework", "view", "-ho", "json", "--", dash},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := root.AcceptDashPrefixedIDs(newRoot(t), tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDashPrefixedIDReachesRun(t *testing.T) {
	t.Parallel()

	dash := gidWithLead(t, 0xF8, 0x00)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "framework view",
			args: []string{"framework", "view", dash, "--json"},
		},
		{
			name: "framework update",
			args: []string{"framework", "update", dash, "--name", "renamed"},
		},
		{
			name: "framework delete",
			args: []string{"framework", "delete", dash, "-y"},
		},
		{
			name: "device view",
			args: []string{"device", "view", dash, "--json"},
		},
		{
			name: "list organization flag",
			args: []string{"framework", "list", "--organization", dash},
		},
		{
			name: "persistent flag before the command",
			args: []string{"--no-color", "framework", "view", dash, "--json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := newNonInteractiveRoot(t)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(root.AcceptDashPrefixedIDs(cmd, tt.args))

			err := cmd.Execute()
			require.ErrorContains(t, err, "config should not be read")
		})
	}
}

func TestUnknownShorthandStillFails(t *testing.T) {
	t.Parallel()

	cmd := newNonInteractiveRoot(t)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(root.AcceptDashPrefixedIDs(cmd, []string{"framework", "view", "-z"}))

	err := cmd.Execute()
	require.ErrorContains(t, err, "unknown shorthand flag")
}

func TestShorthandClusterKeepsValue(t *testing.T) {
	t.Parallel()

	dash := gidWithLead(t, 0xF8, 0x00)

	rootCmd := &cobra.Command{Use: "prb"}

	var (
		gotID   string
		verbose bool
		output  string
	)

	view := &cobra.Command{
		Use:  "view <id>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gotID = args[0]

			var err error

			verbose, err = cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			output, err = cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			return nil
		},
	}
	rootCmd.AddCommand(view)
	view.Flags().BoolP("verbose", "v", false, "verbose")
	view.Flags().StringP("output", "o", "", "output")

	args := []string{"view", dash, "-vo", "json"}
	got := root.AcceptDashPrefixedIDs(rootCmd, args)
	assert.Equal(t, []string{"view", "-vo", "json", "--", dash}, got)

	rootCmd.SetArgs(got)
	require.NoError(t, rootCmd.Execute())
	assert.Equal(t, dash, gotID)
	assert.True(t, verbose)
	assert.Equal(t, "json", output)

	attached := root.AcceptDashPrefixedIDs(rootCmd, []string{"view", dash, "-vojson"})
	assert.Equal(t, []string{"view", "-vojson", "--", dash}, attached)
}

func gidWithLead(t *testing.T, first, second byte) string {
	t.Helper()

	var id gid.GID

	id[0] = first
	id[1] = second

	encoded := id.String()
	_, err := gid.ParseGID(encoded)
	require.NoError(t, err)

	return encoded
}

func newNonInteractiveRoot(t *testing.T) *cobra.Command {
	t.Helper()

	ios := iostreams.System()
	ios.ForceNonInteractive = true

	return root.NewCmdRoot(&cmdutil.Factory{
		IOStreams: ios,
		Config: func() (*config.Config, error) {
			return nil, errors.New("config should not be read")
		},
	})
}
