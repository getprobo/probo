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

package task

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/task/activity"
	asanaprojectsetup "go.probo.inc/probo/pkg/cmd/task/asana-project-setup"
	"go.probo.inc/probo/pkg/cmd/task/comment"
	"go.probo.inc/probo/pkg/cmd/task/create"
	"go.probo.inc/probo/pkg/cmd/task/delete"
	linkasana "go.probo.inc/probo/pkg/cmd/task/link-asana"
	linklinear "go.probo.inc/probo/pkg/cmd/task/link-linear"
	"go.probo.inc/probo/pkg/cmd/task/list"
	listasanaprojects "go.probo.inc/probo/pkg/cmd/task/list-asana-projects"
	listasanatasks "go.probo.inc/probo/pkg/cmd/task/list-asana-tasks"
	listasanateams "go.probo.inc/probo/pkg/cmd/task/list-asana-teams"
	listlinearissues "go.probo.inc/probo/pkg/cmd/task/list-linear-issues"
	listlinearteams "go.probo.inc/probo/pkg/cmd/task/list-linear-teams"
	publishasana "go.probo.inc/probo/pkg/cmd/task/publish-asana"
	publishlinear "go.probo.inc/probo/pkg/cmd/task/publish-linear"
	unlinkexternal "go.probo.inc/probo/pkg/cmd/task/unlink-external"
	"go.probo.inc/probo/pkg/cmd/task/update"
	updateasanaprojectmapping "go.probo.inc/probo/pkg/cmd/task/update-asana-project-mapping"
	updateasanasettings "go.probo.inc/probo/pkg/cmd/task/update-asana-settings"
	updatelinearsettings "go.probo.inc/probo/pkg/cmd/task/update-linear-settings"
	"go.probo.inc/probo/pkg/cmd/task/view"
)

func NewCmdTask(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task <command>",
		Short: "Manage tasks",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(delete.NewCmdDelete(f))
	cmd.AddCommand(comment.NewCmdComment(f))
	cmd.AddCommand(activity.NewCmdActivity(f))
	cmd.AddCommand(listlinearteams.NewCmdListLinearTeams(f))
	cmd.AddCommand(updatelinearsettings.NewCmdUpdateLinearSettings(f))
	cmd.AddCommand(listlinearissues.NewCmdListLinearIssues(f))
	cmd.AddCommand(publishlinear.NewCmdPublishLinear(f))
	cmd.AddCommand(linklinear.NewCmdLinkLinear(f))
	cmd.AddCommand(publishasana.NewCmdPublishAsana(f))
	cmd.AddCommand(linkasana.NewCmdLinkAsana(f))
	cmd.AddCommand(listasanateams.NewCmdListAsanaTeams(f))
	cmd.AddCommand(listasanaprojects.NewCmdListAsanaProjects(f))
	cmd.AddCommand(listasanatasks.NewCmdListAsanaTasks(f))
	cmd.AddCommand(updateasanasettings.NewCmdUpdateAsanaSettings(f))
	cmd.AddCommand(asanaprojectsetup.NewCmdAsanaProjectSetup(f))
	cmd.AddCommand(updateasanaprojectmapping.NewCmdUpdateAsanaProjectMapping(f))
	cmd.AddCommand(unlinkexternal.NewCmdUnlinkExternal(f))

	return cmd
}
