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

package soa

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/soa/create"
	"go.probo.inc/probo/pkg/cmd/soa/delete"
	"go.probo.inc/probo/pkg/cmd/soa/list"
	"go.probo.inc/probo/pkg/cmd/soa/publish"
	"go.probo.inc/probo/pkg/cmd/soa/statement"
	"go.probo.inc/probo/pkg/cmd/soa/update"
	"go.probo.inc/probo/pkg/cmd/soa/view"
)

func NewCmdSoa(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "soa <command>",
		Short: "Manage statements of applicability",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(delete.NewCmdDelete(f))
	cmd.AddCommand(publish.NewCmdPublish(f))
	cmd.AddCommand(statement.NewCmdStatement(f))

	return cmd
}
