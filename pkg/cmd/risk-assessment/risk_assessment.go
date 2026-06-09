// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package riskassessment

import (
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/boundary"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/create"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/delete"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/list"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/node"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/process"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/scenario"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/scope"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/threat"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/update"
	"go.probo.inc/probo/pkg/cmd/risk-assessment/view"
)

func NewCmdRiskAssessment(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "risk-assessment <command>",
		Short: "Manage risk assessments",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(delete.NewCmdDelete(f))
	cmd.AddCommand(scope.NewCmdScope(f))
	cmd.AddCommand(node.NewCmdNode(f))
	cmd.AddCommand(boundary.NewCmdBoundary(f))
	cmd.AddCommand(process.NewCmdProcess(f))
	cmd.AddCommand(threat.NewCmdThreat(f))
	cmd.AddCommand(scenario.NewCmdScenario(f))

	return cmd
}
