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

package inventory

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/access-review/inventory/listchangeevents"
	"go.probo.inc/probo/pkg/cmd/access-review/inventory/listgrantchangeevents"
	"go.probo.inc/probo/pkg/cmd/access-review/inventory/listgrants"
	"go.probo.inc/probo/pkg/cmd/access-review/inventory/listprincipals"
	"go.probo.inc/probo/pkg/cmd/access-review/inventory/listresources"
	"go.probo.inc/probo/pkg/cmd/access-review/inventory/viewgrant"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdInventory(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inventory <command>",
		Short: "Browse live access inventory",
	}

	cmd.AddCommand(listgrants.NewCmdListGrants(f))
	cmd.AddCommand(viewgrant.NewCmdViewGrant(f))
	cmd.AddCommand(listprincipals.NewCmdListPrincipals(f))
	cmd.AddCommand(listresources.NewCmdListResources(f))
	cmd.AddCommand(listchangeevents.NewCmdListChangeEvents(f))
	cmd.AddCommand(listgrantchangeevents.NewCmdListGrantChangeEvents(f))

	return cmd
}
