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

package root

import (
	"github.com/spf13/cobra"
	accessreview "go.probo.inc/probo/pkg/cmd/access-review"
	cmdapi "go.probo.inc/probo/pkg/cmd/api"
	"go.probo.inc/probo/pkg/cmd/auditlog"
	"go.probo.inc/probo/pkg/cmd/auth"
	"go.probo.inc/probo/pkg/cmd/browse"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/completion"
	cmdconfig "go.probo.inc/probo/pkg/cmd/config"
	cmdcontext "go.probo.inc/probo/pkg/cmd/context"
	"go.probo.inc/probo/pkg/cmd/control"
	"go.probo.inc/probo/pkg/cmd/document"
	"go.probo.inc/probo/pkg/cmd/evidence"
	"go.probo.inc/probo/pkg/cmd/finding"
	"go.probo.inc/probo/pkg/cmd/framework"
	"go.probo.inc/probo/pkg/cmd/org"
	"go.probo.inc/probo/pkg/cmd/risk"
	"go.probo.inc/probo/pkg/cmd/soa"
	"go.probo.inc/probo/pkg/cmd/user"
	"go.probo.inc/probo/pkg/cmd/version"
	"go.probo.inc/probo/pkg/cmd/webhook"
)

func NewCmdRoot(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "prb <command> [flags]",
		Short:         "Probo CLI",
		Long:          "prb is a command-line tool for interacting with the Probo platform.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if noInteractive, _ := cmd.Flags().GetBool("no-interactive"); noInteractive {
				f.IOStreams.ForceNonInteractive = true
			}
			if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
				f.IOStreams.ForceNoColor = true
			}
			f.IOStreams.ApplyColorProfile()
		},
	}

	cmd.PersistentFlags().Bool(
		"no-interactive",
		false,
		"Disable interactive prompts (also set via PROBO_NO_INTERACTIVE=1, CI=true, or TERM=dumb)",
	)

	cmd.PersistentFlags().Bool(
		"no-color",
		false,
		"Disable ANSI color output (also set via NO_COLOR or TERM=dumb)",
	)

	cmd.AddCommand(accessreview.NewCmdAccessReview(f))
	cmd.AddCommand(cmdapi.NewCmdAPI(f))
	cmd.AddCommand(auditlog.NewCmdAuditLog(f))
	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(browse.NewCmdBrowse(f))
	cmd.AddCommand(completion.NewCmdCompletion(f))
	cmd.AddCommand(cmdconfig.NewCmdConfig(f))
	cmd.AddCommand(cmdcontext.NewCmdContext(f))
	cmd.AddCommand(control.NewCmdControl(f))
	cmd.AddCommand(document.NewCmdDocument(f))
	cmd.AddCommand(evidence.NewCmdEvidence(f))
	cmd.AddCommand(finding.NewCmdFinding(f))
	cmd.AddCommand(framework.NewCmdFramework(f))
	cmd.AddCommand(org.NewCmdOrg(f))
	cmd.AddCommand(risk.NewCmdRisk(f))
	cmd.AddCommand(soa.NewCmdSoa(f))
	cmd.AddCommand(user.NewCmdUser(f))
	cmd.AddCommand(version.NewCmdVersion(f))
	cmd.AddCommand(webhook.NewCmdWebhook(f))

	return cmd
}
