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
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
	"go.probo.inc/probo/pkg/cmd/root"
)

func TestSkillsAreNotACommand(t *testing.T) {
	t.Parallel()

	rootCmd := newRoot(t)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.SetArgs([]string{"skills", "--help"})

	err := rootCmd.Execute()
	require.ErrorContains(t, err, `unknown command "skills"`)
}

func TestListFlagsRejectBadValues(t *testing.T) {
	t.Parallel()

	var (
		limitCommands []string
		directionCmds []string
	)

	walkCommands(newRoot(t), func(cmd *cobra.Command) {
		if cmd.Flags().Lookup("limit") != nil {
			limitCommands = append(limitCommands, cmd.CommandPath())
		}

		if cmd.Flags().Lookup("order-direction") != nil {
			directionCmds = append(directionCmds, cmd.CommandPath())
		}
	})

	require.GreaterOrEqual(t, len(limitCommands), 60)
	require.GreaterOrEqual(t, len(directionCmds), 45)

	for _, path := range limitCommands {
		t.Run(path+" limit", func(t *testing.T) {
			t.Parallel()

			err := execute(t, path, []string{"--limit", "0"})
			require.ErrorContains(t, err, "invalid --limit value 0")
		})
	}

	for _, path := range directionCmds {
		t.Run(path+" direction", func(t *testing.T) {
			t.Parallel()

			err := execute(t, path, []string{"--order-direction", "SIDEWAYS"})
			require.ErrorContains(t, err, `invalid --order-direction value "SIDEWAYS"`)
		})
	}
}

func execute(t *testing.T, path string, extra []string) error {
	t.Helper()

	rootCmd := newRoot(t)

	name := strings.Fields(path)
	target, _, err := rootCmd.Find(name[1:])
	require.NoError(t, err)

	args := append([]string{}, name[1:]...)
	args = append(args, placeholders(target)...)
	args = append(args, requiredFlagArgs(target)...)
	args = append(args, extra...)

	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	rootCmd.SetArgs(args)

	return rootCmd.Execute()
}

func newRoot(t *testing.T) *cobra.Command {
	t.Helper()

	return root.NewCmdRoot(&cmdutil.Factory{
		IOStreams: iostreams.System(),
		Config: func() (*config.Config, error) {
			return nil, errors.New("config should not be read")
		},
	})
}

func walkCommands(cmd *cobra.Command, visit func(*cobra.Command)) {
	visit(cmd)

	for _, child := range cmd.Commands() {
		walkCommands(child, visit)
	}
}

func placeholders(cmd *cobra.Command) []string {
	count := strings.Count(cmd.Use, "<")

	args := make([]string, count)
	for i := range args {
		args[i] = "id_1"
	}

	return args
}

func requiredFlagArgs(cmd *cobra.Command) []string {
	var args []string

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if len(flag.Annotations[cobra.BashCompOneRequiredFlag]) == 0 {
			return
		}

		switch flag.Value.Type() {
		case "bool":
			args = append(args, "--"+flag.Name)
		case "int":
			args = append(args, "--"+flag.Name, "1")
		default:
			args = append(args, "--"+flag.Name, "value")
		}
	})

	return args
}
