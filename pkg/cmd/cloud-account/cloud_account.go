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

package cloudaccount

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cloud-account/create"
	cadelete "go.probo.inc/probo/pkg/cmd/cloud-account/delete"
	"go.probo.inc/probo/pkg/cmd/cloud-account/get"
	installassets "go.probo.inc/probo/pkg/cmd/cloud-account/install-assets"
	"go.probo.inc/probo/pkg/cmd/cloud-account/list"
	"go.probo.inc/probo/pkg/cmd/cloud-account/rotate"
	"go.probo.inc/probo/pkg/cmd/cloud-account/verify"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdCloudAccount(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloud-account <command>",
		Short:   "Manage cloud accounts (AWS, GCP, Azure)",
		Aliases: []string{"ca"},
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(installassets.NewCmdInstallAssets(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(verify.NewCmdVerify(f))
	cmd.AddCommand(rotate.NewCmdRotate(f))
	cmd.AddCommand(cadelete.NewCmdDelete(f))

	return cmd
}
