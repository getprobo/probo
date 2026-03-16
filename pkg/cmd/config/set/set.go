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

package set

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdConfigSet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Update configuration with a value for the given key",
		Long:  "Update configuration with a value for the given key.\n\nAvailable keys: editor, browser, pager, prompt, http_timeout",
		Example: `  # Set the default editor
  prb config set editor vim

  # Set the HTTP timeout to 60 seconds
  prb config set http_timeout 60s

  # Disable interactive prompts
  prb config set prompt disabled`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			if err := cfg.Set(args[0], args[1]); err != nil {
				return err
			}

			if err := cfg.Save(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.ErrOut,
				"Set %s to %s\n",
				args[0],
				args[1],
			)

			return nil
		},
	}
}
