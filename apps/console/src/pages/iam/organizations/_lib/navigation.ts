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

// Single source of truth for the icon rail and the product segment each
// feature route is nested under in routes.tsx. Panel contents live in the
// matching *NavPanel. Moving a feature between products is an edit here plus
// the matching move in the route tree.

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
    | "canUpdateOrganization";

export interface NavLanding {
  path: string;
  permission: NavPermission;
}

interface NavGroupConfig {
  key: string;
  segment: string | null;
  icon: Icon;
  landings: readonly NavLanding[];
}

export const NAV_GROUPS = [
  {
    key: "governance",
    segment: "governance",
    icon: ScalesIcon,
    landings: [
      { path: "tasks", permission: "canListTasks" },
      { path: "measures", permission: "canListMeasures" },
      { path: "frameworks", permission: "canListFrameworks" },
      { path: "audits", permission: "canListAudits" },
      { path: "findings", permission: "canListFindings" },
      { path: "documents", permission: "canListDocuments" },
      { path: "statements-of-applicability", permission: "canListStatementsOfApplicability" },
    ],
  },
  {
    key: "riskManagement",
    segment: "risk-management",
    icon: WarningIcon,
    landings: [
      { path: "risks", permission: "canListRisks" },
      { path: "risk-analyses", permission: "canListRiskAnalyses" },
    ],
  },
  {
    key: "tprm",
    segment: "tprm",
    icon: StorefrontIcon,
    landings: [
      { path: "third-parties", permission: "canListThirdParties" },
    ],
  },
  {
    key: "privacy",
    segment: "privacy",
    icon: LockIcon,
    landings: [
      { path: "rights-requests", permission: "canListRightsRequests" },
      { path: "processing-activities", permission: "canListProcessingActivities" },
      { path: "dpias", permission: "canListDataProtectionImpactAssessments" },
      { path: "tias", permission: "canListTransferImpactAssessments" },
      { path: "cookie-banners", permission: "canListCookieBanners" },
    ],
  },
  {
    key: "itam",
    segment: "itam",
    icon: LaptopIcon,
    landings: [
      { path: "devices", permission: "canListDevices" },
    ],
  },
  {
    key: "registries",
    segment: "registries",
    icon: BooksIcon,
    landings: [
      { path: "data", permission: "canListData" },
      { path: "assets", permission: "canListAssets" },
      { path: "business-functions", permission: "canListBusinessFunctions" },
      { path: "ai-systems", permission: "canListAiSystems" },
      { path: "obligations", permission: "canListObligations" },
    ],
  },
  {
    key: "compliancePortal",
    segment: null,
    icon: ShieldIcon,
    landings: [
      { path: "compliance-portals", permission: "canGetCompliancePortal" },
    ],
  },
  {
    key: "accessReview",
    segment: "access-reviews",
    icon: KeyIcon,
    landings: [
      { path: "campaigns", permission: "canListAccessReviewCampaigns" },
      { path: "connections", permission: "canListAccessReviewSources" },
    ],
  },
  {
    key: "settings",
    segment: "settings",
    icon: GearIcon,
    landings: [
      { path: "general", permission: "canUpdateOrganization" },
      { path: "context", permission: "canGetContext" },
      { path: "webhooks", permission: "canListWebhookSubscriptions" },
      { path: "people", permission: "canListMembers" },
      { path: "auth", permission: "canUpdateOrganization" },
      { path: "audit-log", permission: "canListAuditLogEntries" },
    ],
  },
] as const satisfies readonly NavGroupConfig[];

export type NavGroup = (typeof NAV_GROUPS)[number];
export type NavGroupKey = NavGroup["key"];

export type NavPermissions = { readonly [K in NavPermission]: boolean };

export function visibleNavGroups(permissions: NavPermissions): NavGroup[] {
  return NAV_GROUPS.filter(group =>
    group.landings.some(landing => permissions[landing.permission]),
  );
}

export function navLandingPath(group: NavGroup, path: string): string {
  return group.segment == null ? path : `${group.segment}/${path}`;
}

export function navHref(organizationId: string, group: NavGroup, path: string): string {
  return `/organizations/${organizationId}/${navLandingPath(group, path)}`;
}

export function navGroupHref(
  organizationId: string,
  group: NavGroup,
  permissions: NavPermissions,
): string {
  for (const landing of group.landings) {
    if (permissions[landing.permission]) {
      return navHref(organizationId, group, landing.path);
    }
  }
  return navHref(organizationId, group, group.landings[0].path);
}
