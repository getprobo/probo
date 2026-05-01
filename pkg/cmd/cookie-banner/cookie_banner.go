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

package cookiebanner

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/activate"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/create"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/deactivate"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/delete"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/list"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/publish"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/translate"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/update"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/latestversion"
	"go.probo.inc/probo/pkg/cmd/cookie-banner/view"
)

func NewCmdCookieBanner(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cookie-banner <command>",
		Short: "Manage cookie banners",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(delete.NewCmdDelete(f))
	cmd.AddCommand(activate.NewCmdActivate(f))
	cmd.AddCommand(deactivate.NewCmdDeactivate(f))
	cmd.AddCommand(publish.NewCmdPublish(f))
	cmd.AddCommand(translate.NewCmdTranslate(f))
	cmd.AddCommand(latestversion.NewCmdLatestVersion(f))

	return cmd
}
