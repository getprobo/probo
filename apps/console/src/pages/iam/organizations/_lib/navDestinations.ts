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

import {
  type NavGroupKey,
  type NavPermissions,
} from "#/pages/iam/organizations/_lib/navigation";

interface NavDestination {
  id: string;
  group: NavGroupKey;
  labelKey: string;
  path: string;
  isVisible: (permissions: NavPermissions, slackbotAvailable: boolean) => boolean;
}

const NAV_DESTINATIONS = [
  {
    id: "tasks",
    group: "governance",
    labelKey: "nav.tasks",
    path: "tasks",
    isVisible: permissions => permissions.canListTasks,
  },
  {
    id: "measures",
    group: "governance",
    labelKey: "nav.measures",
    path: "measures",
    isVisible: permissions => permissions.canListMeasures,
  },
  {
    id: "frameworks",
    group: "governance",
    labelKey: "nav.frameworks",
    path: "frameworks",
    isVisible: permissions => permissions.canListFrameworks,
  },
  {
    id: "audits",
    group: "governance",
    labelKey: "nav.audits",
    path: "audits",
    isVisible: permissions => permissions.canListAudits,
  },
  {
    id: "findings",
    group: "governance",
    labelKey: "nav.findings",
    path: "findings",
    isVisible: permissions => permissions.canListFindings,
  },
  {
    id: "documents",
    group: "governance",
    labelKey: "nav.documents",
    path: "documents",
    isVisible: permissions => permissions.canListDocuments,
  },
  {
    id: "statements-of-applicability",
    group: "governance",
    labelKey: "nav.statementsOfApplicability",
    path: "statements-of-applicability",
    isVisible: permissions => permissions.canListStatementsOfApplicability,
  },
  {
    id: "risk-analyses",
    group: "riskManagement",
    labelKey: "nav.riskAnalyses",
    path: "risk-analyses",
    isVisible: permissions => permissions.canListRiskAnalyses,
  },
  {
    id: "risks",
    group: "riskManagement",
    labelKey: "nav.risks",
    path: "risks",
    isVisible: permissions => permissions.canListRisks,
  },
  {
    id: "third-parties",
    group: "tprm",
    labelKey: "nav.allThirdParties",
    path: "third-parties",
    isVisible: permissions => permissions.canListThirdParties,
  },
  {
    id: "rights-requests",
    group: "privacy",
    labelKey: "nav.rightsRequests",
    path: "rights-requests",
    isVisible: permissions => permissions.canListRightsRequests,
  },
  {
    id: "processing-activities",
    group: "privacy",
    labelKey: "nav.processingActivities",
    path: "processing-activities",
    isVisible: permissions => permissions.canListProcessingActivities,
  },
  {
    id: "dpias",
    group: "privacy",
    labelKey: "nav.dataProtectionImpactAssessments",
    path: "dpias",
    isVisible: permissions => permissions.canListDataProtectionImpactAssessments,
  },
  {
    id: "tias",
    group: "privacy",
    labelKey: "nav.transferImpactAssessments",
    path: "tias",
    isVisible: permissions => permissions.canListTransferImpactAssessments,
  },
  {
    id: "devices",
    group: "itam",
    labelKey: "nav.devices",
    path: "devices",
    isVisible: permissions => permissions.canListDevices,
  },
  {
    id: "data",
    group: "registries",
    labelKey: "nav.data",
    path: "data",
    isVisible: permissions => permissions.canListData,
  },
  {
    id: "assets",
    group: "registries",
    labelKey: "nav.assets",
    path: "assets",
    isVisible: permissions => permissions.canListAssets,
  },
  {
    id: "business-functions",
    group: "registries",
    labelKey: "nav.businessFunctions",
    path: "business-functions",
    isVisible: permissions => permissions.canListBusinessFunctions,
  },
  {
    id: "ai-systems",
    group: "registries",
    labelKey: "nav.aiSystems",
    path: "ai-systems",
    isVisible: permissions => permissions.canListAiSystems,
  },
  {
    id: "obligations",
    group: "registries",
    labelKey: "nav.obligations",
    path: "obligations",
    isVisible: permissions => permissions.canListObligations,
  },
  {
    id: "compliance-portals",
    group: "compliancePortal",
    labelKey: "nav.compliancePortals",
    path: "compliance-portals",
    isVisible: permissions => permissions.canGetCompliancePortal,
  },
  {
    id: "campaigns",
    group: "accessReview",
    labelKey: "nav.campaigns",
    path: "campaigns",
    isVisible: permissions => permissions.canListAccessReviewCampaigns,
  },
  {
    id: "connections",
    group: "accessReview",
    labelKey: "nav.connections",
    path: "connections",
    isVisible: permissions => permissions.canListAccessReviewSources,
  },
  {
    id: "workspace",
    group: "settings",
    labelKey: "nav.workspaceSettings",
    path: "general",
    isVisible: permissions => permissions.canUpdateOrganization,
  },
  {
    id: "context",
    group: "settings",
    labelKey: "nav.context",
    path: "context",
    isVisible: permissions => permissions.canGetContext,
  },
  {
    id: "webhooks",
    group: "settings",
    labelKey: "nav.webhooks",
    path: "webhooks",
    isVisible: permissions => permissions.canListWebhookSubscriptions,
  },
  {
    id: "slackbot",
    group: "settings",
    labelKey: "nav.slackBot",
    path: "slackbot",
    isVisible: (permissions, slackbotAvailable) =>
      slackbotAvailable && (permissions.canConnectSlack || permissions.canUninstallSlack),
  },
  {
    id: "users",
    group: "settings",
    labelKey: "nav.users",
    path: "people",
    isVisible: permissions => permissions.canListMembers,
  },
  {
    id: "saml-sso",
    group: "settings",
    labelKey: "nav.samlSso",
    path: "auth/saml-sso",
    isVisible: permissions => permissions.canUpdateOrganization,
  },
  {
    id: "scim",
    group: "settings",
    labelKey: "nav.scim",
    path: "auth/scim",
    isVisible: permissions => permissions.canUpdateOrganization,
  },
  {
    id: "audit-log",
    group: "settings",
    labelKey: "nav.auditLog",
    path: "audit-log",
    isVisible: permissions => permissions.canListAuditLogEntries,
  },
] as const satisfies readonly NavDestination[];

export type NavDestinationConfig = (typeof NAV_DESTINATIONS)[number];

export function visibleNavDestinations(
  permissions: NavPermissions,
  slackbotAvailable: boolean,
): NavDestinationConfig[] {
  return NAV_DESTINATIONS.filter(destination =>
    destination.isVisible(permissions, slackbotAvailable),
  );
}
