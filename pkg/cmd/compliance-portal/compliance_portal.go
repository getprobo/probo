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

package complianceportal

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/compliance-portal/commitment"
	"go.probo.inc/probo/pkg/cmd/compliance-portal/commitmentgroup"
	"go.probo.inc/probo/pkg/cmd/compliance-portal/file"
	"go.probo.inc/probo/pkg/cmd/compliance-portal/reference"
	"go.probo.inc/probo/pkg/cmd/compliance-portal/update"
	"go.probo.inc/probo/pkg/cmd/compliance-portal/view"
)

func NewCmdCompliancePortal(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "compliance-portal <command>",
		Short:   "Manage compliance portal",
		Aliases: []string{"tc"},
	}

	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(reference.NewCmdReference(f))
	cmd.AddCommand(commitmentgroup.NewCmdCommitmentGroup(f))
	cmd.AddCommand(commitment.NewCmdCommitment(f))
	cmd.AddCommand(file.NewCmdFile(f))

	return cmd
}
