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
  BooksIcon,
  GearIcon,
  type Icon,
  KeyIcon,
  LaptopIcon,
  LockIcon,
  ScalesIcon,
  ShieldIcon,
  StorefrontIcon,
  WarningIcon,
} from "@phosphor-icons/react";

// Product groups: icon, key, and the URL segment each feature route is
// nested under in routes.tsx. Rail visibility and default hrefs live in
// NavRail. Panel contents live in the matching *NavPanel.

export type NavPermission
  = | "canGetContext"
    | "canListTasks"
    | "canListMeasures"
    | "canListRisks"
    | "canListRiskAnalyses"
    | "canListFrameworks"
    | "canListMembers"
    | "canListThirdParties"
    | "canListDocuments"
    | "canListAssets"
    | "canListDevices"
    | "canListData"
    | "canListAudits"
    | "canListFindings"
    | "canListBusinessFunctions"
    | "canListAiSystems"
    | "canListObligations"
    | "canListProcessingActivities"
    | "canListDataProtectionImpactAssessments"
    | "canListTransferImpactAssessments"
    | "canListStatementsOfApplicability"
    | "canListRightsRequests"
    | "canListAccessReviewCampaigns"
    | "canListAccessReviewSources"
    | "canGetCompliancePortal"
    | "canListCookieBanners"
    | "canListAuditLogEntries"
    | "canListWebhookSubscriptions"
    | "canConnectSlack"
    | "canUninstallSlack"
    | "canUpdateOrganization";

interface NavGroupConfig {
  key: string;
  segment: string | null;
  icon: Icon;
}

export const NAV_GROUPS = [
  {
    key: "governance",
    segment: "governance",
    icon: ScalesIcon,
  },
  {
    key: "riskManagement",
    segment: "risk-management",
    icon: WarningIcon,
  },
  {
    key: "tprm",
    segment: "tprm",
    icon: StorefrontIcon,
  },
  {
    key: "privacy",
    segment: "privacy",
    icon: LockIcon,
  },
  {
    key: "itam",
    segment: "itam",
    icon: LaptopIcon,
  },
  {
    key: "registries",
    segment: "registries",
    icon: BooksIcon,
  },
  {
    key: "compliancePortal",
    segment: null,
    icon: ShieldIcon,
  },
  {
    key: "accessReview",
    segment: "access-reviews",
    icon: KeyIcon,
  },
  {
    key: "settings",
    segment: "settings",
    icon: GearIcon,
  },
] as const satisfies readonly NavGroupConfig[];

export type NavGroup = (typeof NAV_GROUPS)[number];
export type NavGroupKey = NavGroup["key"];

export type NavPermissions = { readonly [K in NavPermission]: boolean };

export function navGroupByKey(key: NavGroupKey): NavGroup {
  const group = NAV_GROUPS.find(candidate => candidate.key === key);
  if (group == null) {
    throw new Error(`unknown nav group: ${key}`);
  }
  return group;
}

export function navLandingPath(group: NavGroup, path: string): string {
  return group.segment == null ? path : `${group.segment}/${path}`;
}

export function navHref(organizationId: string, group: NavGroup, path: string): string {
  return `/organizations/${organizationId}/${navLandingPath(group, path)}`;
}
