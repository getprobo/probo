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

/**
 * The console navigation, as one table.
 *
 * This is the single source of truth for three things that must not drift
 * apart: the icon rail, the panel beside it, and the product segment each
 * feature route is nested under in routes.tsx. Moving a feature between
 * products is an edit here plus the matching move in the route tree.
 */

/**
 * Permission aliases requested by the nav fragments. Each maps to a
 * `permission(action:)` field on Organization; an item is hidden when its
 * alias resolves false, and a group disappears once all of its items do.
 */
export type NavPermission
  = | "canGetContext"
    | "canListTasks"
    | "canListMeasures"
    | "canListRisks"
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
    | "canListStatementsOfApplicability"
    | "canListRightsRequests"
    | "canListAccessReviewCampaigns"
    | "canGetCompliancePortal"
    | "canListCookieBanners"
    | "canUpdateOrganization";

export interface NavItem {
  /** Path relative to the group segment, e.g. "frameworks". */
  path: string;
  labelKey: string;
  permission: NavPermission;
}

export interface NavGroup {
  /** Stable id; also the i18n key suffix and the React key. */
  key: string;
  /**
   * URL segment the group's routes are nested under, or null when the product
   * is a single route that already names itself (compliance portals, access
   * reviews) and would only gain a redundant level.
   */
  segment: string | null;
  icon: Icon;
  items: NavItem[];
}

export const NAV_GROUPS: NavGroup[] = [
  {
    key: "governance",
    segment: "governance",
    icon: ScalesIcon,
    items: [
      { path: "frameworks", labelKey: "nav.frameworks", permission: "canListFrameworks" },
      { path: "audits", labelKey: "nav.audits", permission: "canListAudits" },
      { path: "findings", labelKey: "nav.findings", permission: "canListFindings" },
      { path: "measures", labelKey: "nav.measures", permission: "canListMeasures" },
      { path: "documents", labelKey: "nav.documents", permission: "canListDocuments" },
      { path: "tasks", labelKey: "nav.tasks", permission: "canListTasks" },
      {
        path: "statements-of-applicability",
        labelKey: "nav.statementsOfApplicability",
        permission: "canListStatementsOfApplicability",
      },
    ],
  },
  {
    key: "privacy",
    segment: "privacy",
    icon: LockIcon,
    items: [
      { path: "rights-requests", labelKey: "nav.rightsRequests", permission: "canListRightsRequests" },
      {
        path: "processing-activities",
        labelKey: "nav.processingActivities",
        permission: "canListProcessingActivities",
      },
      { path: "cookie-banners", labelKey: "nav.cookieBanners", permission: "canListCookieBanners" },
    ],
  },
  {
    key: "tprm",
    segment: "tprm",
    icon: StorefrontIcon,
    items: [
      { path: "third-parties", labelKey: "nav.thirdParties", permission: "canListThirdParties" },
    ],
  },
  {
    key: "itam",
    segment: "itam",
    icon: LaptopIcon,
    items: [
      { path: "devices", labelKey: "nav.devices", permission: "canListDevices" },
    ],
  },
  {
    key: "riskManagement",
    segment: "risk-management",
    icon: WarningIcon,
    items: [
      { path: "data", labelKey: "nav.data", permission: "canListData" },
      { path: "assets", labelKey: "nav.assets", permission: "canListAssets" },
      { path: "risks", labelKey: "nav.risks", permission: "canListRisks" },
    ],
  },
  {
    key: "registries",
    segment: "registries",
    icon: BooksIcon,
    items: [
      {
        path: "business-functions",
        labelKey: "nav.businessFunctions",
        permission: "canListBusinessFunctions",
      },
      { path: "ai-systems", labelKey: "nav.aiSystems", permission: "canListAiSystems" },
      { path: "obligations", labelKey: "nav.obligations", permission: "canListObligations" },
    ],
  },
  {
    key: "compliancePortal",
    segment: null,
    icon: ShieldIcon,
    items: [
      {
        path: "compliance-portals",
        labelKey: "nav.compliancePortals",
        permission: "canGetCompliancePortal",
      },
    ],
  },
  {
    key: "accessReview",
    segment: null,
    icon: KeyIcon,
    items: [
      {
        path: "access-reviews",
        labelKey: "nav.accessReviews",
        permission: "canListAccessReviewCampaigns",
      },
    ],
  },
  {
    key: "settings",
    segment: "settings",
    icon: GearIcon,
    items: [
      { path: "context", labelKey: "nav.context", permission: "canGetContext" },
      { path: "people", labelKey: "nav.people", permission: "canListMembers" },
      // The organization pages keep their own tab bar, so the panel links to
      // the first tab rather than repeating every tab as a panel entry.
      { path: "general", labelKey: "nav.organization", permission: "canUpdateOrganization" },
    ],
  },
];

/**
 * The permission aliases as the nav fragments resolve them. Both NavRail and
 * NavPanel request the same set, so either fragment's data satisfies this.
 */
export type NavPermissions = { readonly [K in NavPermission]: boolean };

/**
 * The table reduced to what this viewer may see. A group whose every item is
 * denied disappears from the rail rather than leading to an empty panel.
 */
export function visibleNavGroups(permissions: NavPermissions): NavGroup[] {
  const groups: NavGroup[] = [];
  for (const group of NAV_GROUPS) {
    const items = group.items.filter(item => permissions[item.permission]);
    if (items.length > 0) {
      groups.push({ ...group, items });
    }
  }
  return groups;
}

/** Path of an item relative to the organization root. */
export function navItemPath(group: NavGroup, item: NavItem): string {
  return group.segment == null ? item.path : `${group.segment}/${item.path}`;
}

/** Absolute href of an item. */
export function navItemHref(organizationId: string, group: NavGroup, item: NavItem): string {
  return `/organizations/${organizationId}/${navItemPath(group, item)}`;
}
