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
)

const (
	OutputJSON  = "json"
	OutputTable = "table"
)

// AddOutputFlag registers --output / -o on cmd and returns a pointer to the
// value. The default is "table". Callers should call ValidateOutputFlag early
// in RunE, then branch on *p.
func AddOutputFlag(cmd *cobra.Command) *string {
	var output string
	cmd.Flags().StringVarP(
		&output,
		"output",
		"o",
		"",
		"Output format: json, table (default)",
	)

	return &output
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
