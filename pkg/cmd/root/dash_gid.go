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

package root

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.probo.inc/probo/pkg/gid"
)

// AcceptDashPrefixedIDs keeps a positional GID that starts with '-' from
// being parsed as a flag. Base64url encoding uses '-' in its alphabet, so
// some resource ids start with '-'. Flags before the subcommand stay in
// place, and values of real flags are left alone.
func AcceptDashPrefixedIDs(root *cobra.Command, args []string) []string {
	cmd, idx := commandBoundary(root, args)
	leaf := args[idx:]

	rewritten := shieldDashGIDs(cmd, leaf)
	if slices.Equal(rewritten, leaf) {
		return args
	}

	out := make([]string, 0, idx+len(rewritten))
	out = append(out, args[:idx]...)
	out = append(out, rewritten...)

	return out
}

// commandBoundary returns the command selected by args and the index where
// that command's own arguments begin. Flags placed before the subcommand
// stay in the prefix so they are not required to form a suffix.
func commandBoundary(root *cobra.Command, args []string) (*cobra.Command, int) {
	cmd := root

	for i := 0; i < len(args); i++ {
		prepareFlags(cmd)

		arg := args[i]
		if arg == "--" {
			return cmd, i
		}

		if _, err := gid.ParseGID(arg); err == nil {
			return cmd, i
		}

		if isFlagToken(arg) {
			if flagConsumesNext(cmd.Flags(), arg) && i+1 < len(args) {
				i++
			}

			continue
		}

		child := findChild(cmd, arg)
		if child == nil {
			return cmd, i
		}

		cmd = child
	}

	return cmd, len(args)
}

func shieldDashGIDs(cmd *cobra.Command, args []string) []string {
	prepareFlags(cmd)

	flags := cmd.Flags()

	var (
		flagArgs    []string
		positionals []string
		dashGID     bool
	)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}

		if _, err := gid.ParseGID(arg); err == nil {
			positionals = append(positionals, arg)
			if strings.HasPrefix(arg, "-") {
				dashGID = true
			}

			continue
		}

		if isFlagToken(arg) {
			flagArgs = append(flagArgs, arg)
			if flagConsumesNext(flags, arg) && i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}

			continue
		}

		positionals = append(positionals, arg)
	}

	if !dashGID {
		return args
	}

	out := make([]string, 0, len(flagArgs)+len(positionals)+1)
	out = append(out, flagArgs...)
	out = append(out, "--")
	out = append(out, positionals...)

	return out
}

func prepareFlags(cmd *cobra.Command) {
	// Merge parent flags and register --help before shorthand clusters are read.
	cmd.InitDefaultHelpFlag()
}

func findChild(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name || child.HasAlias(name) {
			return child
		}
	}

	return nil
}

func isFlagToken(arg string) bool {
	return len(arg) >= 2 && arg[0] == '-'
}

func flagConsumesNext(flags *pflag.FlagSet, arg string) bool {
	if strings.Contains(arg, "=") || len(arg) < 2 || arg[0] != '-' {
		return false
	}

	if strings.HasPrefix(arg, "--") {
		flag := flags.Lookup(arg[2:])
		if flag == nil {
			return false
		}

		return flag.NoOptDefVal == ""
	}

	// A value-taking shorthand consumes the next argument only when it
	// is the last character. An earlier one takes the rest of the cluster.
	shorthands := arg[1:]
	for i := 0; i < len(shorthands); i++ {
		flag := flags.ShorthandLookup(shorthands[i : i+1])
		if flag == nil {
			return false
		}

		if flag.NoOptDefVal != "" {
			continue
		}

		return i == len(shorthands)-1
	}

	return false
}
