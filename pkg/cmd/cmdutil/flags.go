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

package cmdutil

import (
	"fmt"
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
	for _, v := range allowed {
		if value == v {
			return nil
		}
	}
	return fmt.Errorf(
		"invalid --%s value %q: valid values are %s",
		flag,
		value,
		strings.Join(allowed, ", "),
	)
}
