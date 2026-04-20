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
	publishmajor "go.probo.inc/probo/pkg/cmd/document/publish-major"
	publishminor "go.probo.inc/probo/pkg/cmd/document/publish-minor"
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
	cmd.AddCommand(publishmajor.NewCmdPublishMajor(f))
	cmd.AddCommand(publishminor.NewCmdPublishMinor(f))
	cmd.AddCommand(listapprovalquorums.NewCmdListApprovalQuorums(f))
	cmd.AddCommand(viewapprovalquorum.NewCmdViewApprovalQuorum(f))
	cmd.AddCommand(listapprovaldecisions.NewCmdListApprovalDecisions(f))
	cmd.AddCommand(viewapprovaldecision.NewCmdViewApprovalDecision(f))

	return cmd
}
