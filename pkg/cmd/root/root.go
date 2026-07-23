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

package root

import (
	"github.com/spf13/cobra"
	accessreview "go.probo.inc/probo/pkg/cmd/access-review"
	cmdapi "go.probo.inc/probo/pkg/cmd/api"
	"go.probo.inc/probo/pkg/cmd/asset"
	"go.probo.inc/probo/pkg/cmd/audit"
	"go.probo.inc/probo/pkg/cmd/auditlog"
	"go.probo.inc/probo/pkg/cmd/auth"
	"go.probo.inc/probo/pkg/cmd/browse"
	businessfunction "go.probo.inc/probo/pkg/cmd/businessfunction"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/completion"
	complianceportal "go.probo.inc/probo/pkg/cmd/compliance-portal"
	cmdconfig "go.probo.inc/probo/pkg/cmd/config"
	consentrecord "go.probo.inc/probo/pkg/cmd/consent-record"
	cmdcontext "go.probo.inc/probo/pkg/cmd/context"
	"go.probo.inc/probo/pkg/cmd/control"
	cookiebanner "go.probo.inc/probo/pkg/cmd/cookie-banner"
	cookiecategory "go.probo.inc/probo/pkg/cmd/cookie-category"
	"go.probo.inc/probo/pkg/cmd/datum"
	"go.probo.inc/probo/pkg/cmd/document"
	"go.probo.inc/probo/pkg/cmd/dpia"
	"go.probo.inc/probo/pkg/cmd/evidence"
	"go.probo.inc/probo/pkg/cmd/finding"
	"go.probo.inc/probo/pkg/cmd/framework"
	"go.probo.inc/probo/pkg/cmd/measure"
	"go.probo.inc/probo/pkg/cmd/obligation"
	"go.probo.inc/probo/pkg/cmd/org"
	processingactivity "go.probo.inc/probo/pkg/cmd/processing-activity"
	resourcealias "go.probo.inc/probo/pkg/cmd/resource-alias"
	rightsrequest "go.probo.inc/probo/pkg/cmd/rights-request"
	"go.probo.inc/probo/pkg/cmd/risk"
	riskassessment "go.probo.inc/probo/pkg/cmd/risk-assessment"
	"go.probo.inc/probo/pkg/cmd/scim"
	"go.probo.inc/probo/pkg/cmd/soa"
	"go.probo.inc/probo/pkg/cmd/task"
	"go.probo.inc/probo/pkg/cmd/thirdpartymgmt"
	"go.probo.inc/probo/pkg/cmd/tia"
	trackerpattern "go.probo.inc/probo/pkg/cmd/tracker-pattern"
	trackerresource "go.probo.inc/probo/pkg/cmd/tracker-resource"
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
	cmd.AddCommand(asset.NewCmdAsset(f))
	cmd.AddCommand(audit.NewCmdAudit(f))
	cmd.AddCommand(auditlog.NewCmdAuditLog(f))
	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(browse.NewCmdBrowse(f))
	cmd.AddCommand(businessfunction.NewCmdBusinessFunction(f))
	cmd.AddCommand(completion.NewCmdCompletion(f))
	cmd.AddCommand(cmdconfig.NewCmdConfig(f))
	cmd.AddCommand(consentrecord.NewCmdConsentRecord(f))
	cmd.AddCommand(cmdcontext.NewCmdContext(f))
	cmd.AddCommand(control.NewCmdControl(f))
	cmd.AddCommand(cookiebanner.NewCmdCookieBanner(f))
	cmd.AddCommand(cookiecategory.NewCmdCookieCategory(f))
	cmd.AddCommand(trackerpattern.NewCmdTrackerPattern(f))
	cmd.AddCommand(trackerresource.NewCmdTrackerResource(f))
	cmd.AddCommand(datum.NewCmdDatum(f))
	cmd.AddCommand(document.NewCmdDocument(f))
	cmd.AddCommand(dpia.NewCmdDPIA(f))
	cmd.AddCommand(evidence.NewCmdEvidence(f))
	cmd.AddCommand(finding.NewCmdFinding(f))
	cmd.AddCommand(framework.NewCmdFramework(f))
	cmd.AddCommand(measure.NewCmdMeasure(f))
	cmd.AddCommand(obligation.NewCmdObligation(f))
	cmd.AddCommand(org.NewCmdOrg(f))
	cmd.AddCommand(processingactivity.NewCmdProcessingActivity(f))
	cmd.AddCommand(rightsrequest.NewCmdRightsRequest(f))
	cmd.AddCommand(risk.NewCmdRisk(f))
	cmd.AddCommand(riskassessment.NewCmdRiskAssessment(f))
	cmd.AddCommand(resourcealias.NewCmdResourceAlias(f))
	cmd.AddCommand(scim.NewCmdScim(f))
	cmd.AddCommand(soa.NewCmdSoa(f))
	cmd.AddCommand(task.NewCmdTask(f))
	cmd.AddCommand(tia.NewCmdTIA(f))
	cmd.AddCommand(complianceportal.NewCmdCompliancePortal(f))
	cmd.AddCommand(user.NewCmdUser(f))
	cmd.AddCommand(thirdpartymgmt.NewCmdThirdParty(f))
	cmd.AddCommand(version.NewCmdVersion(f))
	cmd.AddCommand(webhook.NewCmdWebhook(f))

	return cmd
}
