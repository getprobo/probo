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

package document

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/document/archive"
	"go.probo.inc/probo/pkg/cmd/document/create"
	"go.probo.inc/probo/pkg/cmd/document/delete"
	deletedraft "go.probo.inc/probo/pkg/cmd/document/delete-draft"
	"go.probo.inc/probo/pkg/cmd/document/list"
	listapprovaldecisions "go.probo.inc/probo/pkg/cmd/document/list-approval-decisions"
	listapprovalquorums "go.probo.inc/probo/pkg/cmd/document/list-approval-quorums"
	listversions "go.probo.inc/probo/pkg/cmd/document/list-versions"
	"go.probo.inc/probo/pkg/cmd/document/publish"
	"go.probo.inc/probo/pkg/cmd/document/unarchive"
	"go.probo.inc/probo/pkg/cmd/document/update"
	"go.probo.inc/probo/pkg/cmd/document/view"
	viewapprovaldecision "go.probo.inc/probo/pkg/cmd/document/view-approval-decision"
	viewapprovalquorum "go.probo.inc/probo/pkg/cmd/document/view-approval-quorum"
	viewversion "go.probo.inc/probo/pkg/cmd/document/view-version"
)

func NewCmdDocument(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "document <command>",
		Short: "Manage documents",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(delete.NewCmdDelete(f))
	cmd.AddCommand(archive.NewCmdArchive(f))
	cmd.AddCommand(unarchive.NewCmdUnarchive(f))
	cmd.AddCommand(listversions.NewCmdListVersions(f))
	cmd.AddCommand(viewversion.NewCmdViewVersion(f))
	cmd.AddCommand(deletedraft.NewCmdDeleteDraft(f))
	cmd.AddCommand(publish.NewCmdPublish(f))
	cmd.AddCommand(listapprovalquorums.NewCmdListApprovalQuorums(f))
	cmd.AddCommand(viewapprovalquorum.NewCmdViewApprovalQuorum(f))
	cmd.AddCommand(listapprovaldecisions.NewCmdListApprovalDecisions(f))
	cmd.AddCommand(viewapprovaldecision.NewCmdViewApprovalDecision(f))

	return cmd
}
