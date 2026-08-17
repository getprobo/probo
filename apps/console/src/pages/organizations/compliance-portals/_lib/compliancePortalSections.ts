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

import { graphql } from "relay-runtime";

import type { compliancePortalSections_compliancePortal$data } from "#/__generated__/core/compliancePortalSections_compliancePortal.graphql";

export const compliancePortalSectionsFragment = graphql`
  fragment compliancePortalSections_compliancePortal on CompliancePortal {
    canUpdatePortal: permission(action: "compliance-portal:portal:update")
    canGetCustomDomain: permission(action: "compliance-portal:custom-domain:get")
    canListAccesses: permission(action: "compliance-portal:portal-access:list")
    canListCustomLinks: permission(action: "compliance-portal:compliance-custom-link:list")
    canListCommitmentGroups: permission(action: "compliance-portal:commitment-group:list")
    canListCommitments: permission(action: "compliance-portal:commitment:list")
    canListReferences: permission(action: "compliance-portal:portal-reference:list")
    canListFrameworks: permission(action: "compliance-portal:compliance-framework:list")
    canListAudits: permission(action: "core:audit:list")
    canListDocuments: permission(action: "core:document:list")
    canListFiles: permission(action: "compliance-portal:portal-file:list")
    canListThirdParties: permission(action: "core:thirdParty:list")
    canListMailingListSubscribers: permission(
      action: "compliance-portal:mailing-list-subscriber:list"
    )
    canListMailingListUpdates: permission(
      action: "compliance-portal:mailing-list-update:list"
    )
    organization @required(action: THROW) {
      canConnectSlack: permission(action: "core:connector:initiate")
      canListSlackConnections: permission(action: "core:slack-connection:list")
    }
  }
`;

export type CompliancePortalSectionId
  = | "hosting"
    | "permissions"
    | "integrations"
    | "landing"
    | "documents"
    | "subprocessors"
    | "updates"
    | "right-requests";

export type CompliancePortalSectionGroup = "settings" | "pages";

export interface CompliancePortalSectionPermissions {
  canUpdatePortal: boolean;
  canGetCustomDomain: boolean;
  canListAccesses: boolean;
  canConnectSlack: boolean;
  canListSlackConnections: boolean;
  canListCustomLinks: boolean;
  canListCommitmentGroups: boolean;
  canListCommitments: boolean;
  canListReferences: boolean;
  canListFrameworks: boolean;
  canListAudits: boolean;
  canListDocuments: boolean;
  canListFiles: boolean;
  canListThirdParties: boolean;
  canListMailingListSubscribers: boolean;
  canListMailingListUpdates: boolean;
}

export interface CompliancePortalSection {
  id: CompliancePortalSectionId;
  path: string;
  labelKey: string;
  group: CompliancePortalSectionGroup;
  isVisible: (permissions: CompliancePortalSectionPermissions) => boolean;
}

export const COMPLIANCE_PORTAL_SECTIONS: CompliancePortalSection[] = [
  {
    id: "hosting",
    path: "hosting",
    labelKey: "nav.compliancePortalsHosting",
    group: "settings",
    isVisible: permissions =>
      permissions.canUpdatePortal && permissions.canGetCustomDomain,
  },
  {
    id: "permissions",
    path: "permissions",
    labelKey: "nav.compliancePortalsPermissions",
    group: "settings",
    isVisible: permissions => permissions.canListAccesses,
  },
  {
    id: "integrations",
    path: "integrations",
    labelKey: "nav.compliancePortalsIntegrations",
    group: "settings",
    isVisible: permissions =>
      permissions.canConnectSlack && permissions.canListSlackConnections,
  },
  {
    id: "landing",
    path: "landing",
    labelKey: "nav.compliancePortalsLanding",
    group: "pages",
    isVisible: permissions =>
      permissions.canListCustomLinks
      || (
        permissions.canListFrameworks
        && permissions.canListCommitmentGroups
        && permissions.canListCommitments
        && permissions.canListReferences
      ),
  },
  {
    id: "documents",
    path: "documents",
    labelKey: "nav.compliancePortalsDocuments",
    group: "pages",
    isVisible: permissions =>
      permissions.canListAudits
      || permissions.canListDocuments
      || permissions.canListFiles,
  },
  {
    id: "subprocessors",
    path: "subprocessors",
    labelKey: "nav.compliancePortalsSubprocessors",
    group: "pages",
    isVisible: permissions => permissions.canListThirdParties,
  },
  {
    id: "updates",
    path: "updates",
    labelKey: "nav.compliancePortalsUpdates",
    group: "pages",
    isVisible: permissions =>
      permissions.canListMailingListSubscribers
      && permissions.canListMailingListUpdates,
  },
  {
    id: "right-requests",
    path: "right-requests",
    labelKey: "nav.compliancePortalsRightRequests",
    group: "pages",
    isVisible: permissions => permissions.canUpdatePortal,
  },
];

export function sectionPermissionsFrom(
  data: compliancePortalSections_compliancePortal$data,
): CompliancePortalSectionPermissions {
  return {
    canUpdatePortal: data.canUpdatePortal,
    canGetCustomDomain: data.canGetCustomDomain,
    canListAccesses: data.canListAccesses,
    canConnectSlack: data.organization.canConnectSlack,
    canListSlackConnections: data.organization.canListSlackConnections,
    canListCustomLinks: data.canListCustomLinks,
    canListCommitmentGroups: data.canListCommitmentGroups,
    canListCommitments: data.canListCommitments,
    canListReferences: data.canListReferences,
    canListFrameworks: data.canListFrameworks,
    canListAudits: data.canListAudits,
    canListDocuments: data.canListDocuments,
    canListFiles: data.canListFiles,
    canListThirdParties: data.canListThirdParties,
    canListMailingListSubscribers: data.canListMailingListSubscribers,
    canListMailingListUpdates: data.canListMailingListUpdates,
  };
}

export function visibleCompliancePortalSections(
  permissions: CompliancePortalSectionPermissions,
): CompliancePortalSection[] {
  return COMPLIANCE_PORTAL_SECTIONS.filter(section => section.isVisible(permissions));
}

export function firstVisibleSection(
  permissions: CompliancePortalSectionPermissions,
): CompliancePortalSection | undefined {
  return visibleCompliancePortalSections(permissions)[0];
}

export const COMPLIANCE_PORTAL_SECTION_GROUPS: {
  id: CompliancePortalSectionGroup;
  labelKey: string;
}[] = [
  { id: "settings", labelKey: "nav.compliancePortalsSettings" },
  { id: "pages", labelKey: "nav.compliancePortalsPages" },
];
