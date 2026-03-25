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

package entry

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/accessreview/entry/decide"
	"go.probo.inc/probo/pkg/cmd/accessreview/entry/decideall"
	entryflag "go.probo.inc/probo/pkg/cmd/accessreview/entry/flag"
	"go.probo.inc/probo/pkg/cmd/accessreview/entry/list"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdEntry(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entry <command>",
		Short: "Manage access review entries",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(entryflag.NewCmdFlag(f))
	cmd.AddCommand(decide.NewCmdDecide(f))
	cmd.AddCommand(decideall.NewCmdDecideAll(f))

	return cmd
}
