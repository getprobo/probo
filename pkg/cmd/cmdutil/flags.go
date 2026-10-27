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

package cmdutil

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	OutputJSON  = "json"
	OutputTable = "table"
)

func TaskStates() []string {
	states := coredata.TaskStates()

	out := make([]string, len(states))
	for i, state := range states {
		out[i] = state.String()
	}

	return out
}

func TaskStateFlagUsage() string {
	return "Task state: " + strings.Join(TaskStates(), ", ")
}

func TaskPriorities() []string {
	priorities := coredata.TaskPriorities()

	out := make([]string, len(priorities))
	for i, priority := range priorities {
		out[i] = priority.String()
	}

	return out
}

func TaskPriorityFlagUsage() string {
	return "Task priority: " + strings.Join(TaskPriorities(), ", ")
}

// AddOutputFlag registers --output / -o and --json on cmd and returns a
// pointer to the value. The default is "table". --json is shorthand for
// --output json. Callers should call ValidateOutputFlag early in RunE, then
// branch on *p.
//
// This chains onto any cmd.PreRun or cmd.PreRunE already set (see
// chainPreRunE) to resolve --json before RunE runs. Call this, and the
// other cmdutil Add*Flag helpers that document the same behavior, before
// setting a PreRun or PreRunE of your own directly on cmd (for example
// cmd.PreRunE = func(...) {...} or via the cobra.Command struct literal) so
// that a later direct assignment wraps this chain instead of replacing it.
func AddOutputFlag(cmd *cobra.Command) *string {
	var (
		output string
		asJSON bool
	)

	cmd.Flags().StringVarP(
		&output,
		"output",
		"o",
		"",
		"Output format: json, table (default)",
	)
	cmd.Flags().BoolVar(
		&asJSON,
		"json",
		false,
		"Output JSON (same as --output json)",
	)

	chainPreRunE(cmd, func(cmd *cobra.Command, args []string) error {
		if asJSON {
			if cmd.Flags().Changed("output") && output != "" && output != OutputJSON {
				return fmt.Errorf("--json cannot be combined with --output %s", output)
			}

			output = OutputJSON
		}

		return nil
	})

	return &output
}

// chainPreRunE installs fn as cmd.PreRunE, running it before whatever
// PreRunE (or, lacking that, PreRun) was already set on cmd. Cobra runs
// PreRun only when PreRunE is nil, so replacing PreRunE outright would
// silently drop a caller's PreRun; this chains both.
func chainPreRunE(cmd *cobra.Command, fn func(cmd *cobra.Command, args []string) error) {
	previousE := cmd.PreRunE
	previous := cmd.PreRun
	cmd.PreRun = nil
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := fn(cmd, args); err != nil {
			return err
		}

		if previousE != nil {
			return previousE(cmd, args)
		}

		if previous != nil {
			previous(cmd, args)
		}

		return nil
	}
}

// AddOrganizationFlag registers --org. --organization is the same flag.
// Do not call this on a command that already uses --organization for a
// different meaning, such as connector connect.
func AddOrganizationFlag(cmd *cobra.Command, dest *string, usage string) {
	cmd.Flags().StringVar(dest, "org", "", usage)
	cmd.Flags().SetNormalizeFunc(normalizeOrganizationFlag)
}

// AddAfterFlag registers --after. The value is the cursor printed as
// "Next cursor" by the previous page of the same list.
func AddAfterFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVar(
		dest,
		"after",
		"",
		"Cursor from the previous page (the Next cursor line)",
	)
}

// AddOrderDirectionFlag registers --order-direction. value is the default,
// ASC or DESC. The flag is checked before RunE, including when --order-by
// is omitted, so a bad direction is not silently dropped. See AddOutputFlag
// for the PreRun/PreRunE chaining contract this relies on.
func AddOrderDirectionFlag(cmd *cobra.Command, dest *string, value string) {
	cmd.Flags().StringVar(
		dest,
		"order-direction",
		value,
		"Sort direction (ASC, DESC)",
	)

	chainPreRunE(cmd, func(cmd *cobra.Command, args []string) error {
		return ValidateOrderDirection(*dest)
	})
}

func normalizeOrganizationFlag(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	name = strings.ReplaceAll(name, "_", "-")
	if name == "organization" {
		return "org"
	}

	return pflag.NormalizedName(name)
}

// ValidateOutputFlag checks that value is a supported output format. An empty
// string is treated as table (the default).
func ValidateOutputFlag(value *string) error {
	switch *value {
	case "":
		*value = OutputTable
		return nil
	case OutputJSON, OutputTable:
		return nil
	default:
		return fmt.Errorf(
			"invalid --output value %q: valid values are json, table",
			*value,
		)
	}
}

// ValidateOrderDirection checks that value is ASC or DESC.
func ValidateOrderDirection(value string) error {
	return ValidateEnum("order-direction", value, []string{"ASC", "DESC"})
}

// ValidateEnum checks that value is one of the allowed values. It returns a
// user-friendly error mentioning the flag name and the valid choices.
func ValidateEnum(flag string, value string, allowed []string) error {
	if slices.Contains(allowed, value) {
		return nil
	}

	return fmt.Errorf(
		"invalid --%s value %q: valid values are %s",
		flag,
		value,
		strings.Join(allowed, ", "),
	)
}

// ValidateLimit checks that a --limit value is positive. A non-positive limit
// would otherwise cause pagination to return no results without an error.
func ValidateLimit(value int) error {
	if value <= 0 {
		return fmt.Errorf("invalid --limit value %d: must be greater than 0", value)
	}

	return nil
}
