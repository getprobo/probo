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

import { LifebuoyIcon, MoonIcon, SunIcon } from "@phosphor-icons/react";
import { useDisplayMode } from "@probo/ui/src/v2/displayMode/useDisplayMode";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { navPermissions_organization$key } from "#/__generated__/iam/navPermissions_organization.graphql";
import type { NavRail_organization$key } from "#/__generated__/iam/NavRail_organization.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  NAV_GROUPS,
  type NavGroup,
  navGroupByKey,
  type NavGroupKey,
  navHref,
  type NavPermissions,
} from "#/pages/iam/organizations/_lib/navigation";
import { navPermissionsFragment } from "#/pages/iam/organizations/_lib/navPermissions";
import { useActiveNavGroup } from "#/pages/iam/organizations/_lib/useActiveNavGroup";

import { NavRailItem } from "./NavRailItem";
import { OrganizationSwitcher } from "./OrganizationSwitcher";
import { navRail } from "./variants";
import { ViewerMembershipMenu } from "./ViewerMembershipMenu";

const navRailFragment = graphql`
  fragment NavRail_organization on Organization {
    ...OrganizationSwitcher_organization
    ...ViewerMembershipMenu_organization
    ...navPermissions_organization
  }
`;

export interface NavRailProps {
  organizationKey: NavRail_organization$key;
  slackbotAvailable: boolean;
}

function isGovernanceVisible(permissions: NavPermissions): boolean {
  return permissions.canListTasks
    || permissions.canListMeasures
    || permissions.canListFrameworks
    || permissions.canListAudits
    || permissions.canListFindings
    || permissions.canListDocuments
    || permissions.canListStatementsOfApplicability;
}

function isRiskManagementVisible(permissions: NavPermissions): boolean {
  return permissions.canListRisks || permissions.canListRiskAnalyses;
}

function isTprmVisible(permissions: NavPermissions): boolean {
  return permissions.canListThirdParties;
}

function isPrivacyVisible(permissions: NavPermissions): boolean {
  return permissions.canListRightsRequests
    || permissions.canListProcessingActivities
    || permissions.canListDataProtectionImpactAssessments
    || permissions.canListTransferImpactAssessments
    || permissions.canListCookieBanners;
}

function isItamVisible(permissions: NavPermissions): boolean {
  return permissions.canListDevices;
}

function isRegistriesVisible(permissions: NavPermissions): boolean {
  return permissions.canListData
    || permissions.canListAssets
    || permissions.canListBusinessFunctions
    || permissions.canListAiSystems
    || permissions.canListObligations;
}

function isCompliancePortalVisible(permissions: NavPermissions): boolean {
  return permissions.canGetCompliancePortal;
}

function isAccessReviewVisible(permissions: NavPermissions): boolean {
  return permissions.canListAccessReviewCampaigns || permissions.canListAccessReviewSources;
}

function isSettingsVisible(
  permissions: NavPermissions,
  slackbotAvailable: boolean,
): boolean {
  return permissions.canUpdateOrganization
    || permissions.canGetContext
    || permissions.canListWebhookSubscriptions
    || (slackbotAvailable && (permissions.canConnectSlack || permissions.canUninstallSlack))
    || permissions.canListMembers
    || permissions.canListAuditLogEntries;
}

function navGroupIsVisible(
  key: NavGroupKey,
  permissions: NavPermissions,
  slackbotAvailable: boolean,
): boolean {
  switch (key) {
    case "governance":
      return isGovernanceVisible(permissions);
    case "riskManagement":
      return isRiskManagementVisible(permissions);
    case "tprm":
      return isTprmVisible(permissions);
    case "privacy":
      return isPrivacyVisible(permissions);
    case "itam":
      return isItamVisible(permissions);
    case "registries":
      return isRegistriesVisible(permissions);
    case "compliancePortal":
      return isCompliancePortalVisible(permissions);
    case "accessReview":
      return isAccessReviewVisible(permissions);
    case "settings":
      return isSettingsVisible(permissions, slackbotAvailable);
  }
}

export function visibleNavGroups(
  permissions: NavPermissions,
  slackbotAvailable: boolean,
): NavGroup[] {
  return NAV_GROUPS.filter(group =>
    navGroupIsVisible(group.key, permissions, slackbotAvailable),
  );
}

function groupHref(
  organizationId: string,
  key: NavGroupKey,
  path: string,
): string {
  return navHref(organizationId, navGroupByKey(key), path);
}

function governanceHref(organizationId: string, permissions: NavPermissions): string {
  if (permissions.canListTasks) {
    return groupHref(organizationId, "governance", "tasks");
  }
  if (permissions.canListMeasures) {
    return groupHref(organizationId, "governance", "measures");
  }
  if (permissions.canListFrameworks) {
    return groupHref(organizationId, "governance", "frameworks");
  }
  if (permissions.canListAudits) {
    return groupHref(organizationId, "governance", "audits");
  }
  if (permissions.canListFindings) {
    return groupHref(organizationId, "governance", "findings");
  }
  if (permissions.canListDocuments) {
    return groupHref(organizationId, "governance", "documents");
  }
  return groupHref(organizationId, "governance", "statements-of-applicability");
}

function riskManagementHref(organizationId: string, permissions: NavPermissions): string {
  if (permissions.canListRiskAnalyses) {
    return groupHref(organizationId, "riskManagement", "risk-analyses");
  }
  return groupHref(organizationId, "riskManagement", "risks");
}

function privacyHref(organizationId: string, permissions: NavPermissions): string {
  if (permissions.canListRightsRequests) {
    return groupHref(organizationId, "privacy", "rights-requests");
  }
  if (permissions.canListProcessingActivities) {
    return groupHref(organizationId, "privacy", "processing-activities");
  }
  if (permissions.canListDataProtectionImpactAssessments) {
    return groupHref(organizationId, "privacy", "dpias");
  }
  if (permissions.canListTransferImpactAssessments) {
    return groupHref(organizationId, "privacy", "tias");
  }
  return groupHref(organizationId, "privacy", "cookie-banners/new");
}

function registriesHref(organizationId: string, permissions: NavPermissions): string {
  if (permissions.canListData) {
    return groupHref(organizationId, "registries", "data");
  }
  if (permissions.canListAssets) {
    return groupHref(organizationId, "registries", "assets");
  }
  if (permissions.canListBusinessFunctions) {
    return groupHref(organizationId, "registries", "business-functions");
  }
  if (permissions.canListAiSystems) {
    return groupHref(organizationId, "registries", "ai-systems");
  }
  return groupHref(organizationId, "registries", "obligations");
}

function accessReviewHref(organizationId: string, permissions: NavPermissions): string {
  if (permissions.canListAccessReviewCampaigns) {
    return groupHref(organizationId, "accessReview", "campaigns");
  }
  return groupHref(organizationId, "accessReview", "connections");
}

function settingsHref(
  organizationId: string,
  permissions: NavPermissions,
  slackbotAvailable: boolean,
): string {
  if (permissions.canUpdateOrganization) {
    return groupHref(organizationId, "settings", "general");
  }
  if (permissions.canGetContext) {
    return groupHref(organizationId, "settings", "context");
  }
  if (permissions.canListWebhookSubscriptions) {
    return groupHref(organizationId, "settings", "webhooks");
  }
  if (slackbotAvailable && (permissions.canConnectSlack || permissions.canUninstallSlack)) {
    return groupHref(organizationId, "settings", "slackbot");
  }
  if (permissions.canListMembers) {
    return groupHref(organizationId, "settings", "people");
  }
  return groupHref(organizationId, "settings", "audit-log");
}

export function NavRail({ organizationKey, slackbotAvailable }: NavRailProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const organization = useFragment(navRailFragment, organizationKey);
  const { displayMode, toggleDisplayMode } = useDisplayMode();

  const permissions = useFragment<navPermissions_organization$key>(
    navPermissionsFragment,
    organization,
  );
  const groups = useMemo(
    () => visibleNavGroups(permissions, slackbotAvailable),
    [permissions, slackbotAvailable],
  );
  const activeGroup = useActiveNavGroup(groups);
  const activeKey = activeGroup?.key;

  const slots = navRail();

  return (
    <div className={slots.root()}>
      <nav className={slots.rail()} aria-label={t("nav.products")}>
        <OrganizationSwitcher organizationKey={organization} />

        <div className={slots.items()}>
          {isGovernanceVisible(permissions) && (
            <NavRailItem
              icon={navGroupByKey("governance").icon}
              label={t("nav.groups.governance")}
              to={governanceHref(organizationId, permissions)}
              active={activeKey === "governance"}
            />
          )}
          {isRiskManagementVisible(permissions) && (
            <NavRailItem
              icon={navGroupByKey("riskManagement").icon}
              label={t("nav.groups.riskManagement")}
              to={riskManagementHref(organizationId, permissions)}
              active={activeKey === "riskManagement"}
            />
          )}
          {isTprmVisible(permissions) && (
            <NavRailItem
              icon={navGroupByKey("tprm").icon}
              label={t("nav.groups.tprm")}
              to={groupHref(organizationId, "tprm", "third-parties")}
              active={activeKey === "tprm"}
            />
          )}
          {isPrivacyVisible(permissions) && (
            <NavRailItem
              icon={navGroupByKey("privacy").icon}
              label={t("nav.groups.privacy")}
              to={privacyHref(organizationId, permissions)}
              active={activeKey === "privacy"}
            />
          )}
          {isItamVisible(permissions) && (
            <NavRailItem
              icon={navGroupByKey("itam").icon}
              label={t("nav.groups.itam")}
              to={groupHref(organizationId, "itam", "devices")}
              active={activeKey === "itam"}
            />
          )}
          {isRegistriesVisible(permissions) && (
            <NavRailItem
              icon={navGroupByKey("registries").icon}
              label={t("nav.groups.registries")}
              to={registriesHref(organizationId, permissions)}
              active={activeKey === "registries"}
            />
          )}
          {isCompliancePortalVisible(permissions) && (
            <NavRailItem
              icon={navGroupByKey("compliancePortal").icon}
              label={t("nav.groups.compliancePortal")}
              to={groupHref(organizationId, "compliancePortal", "compliance-portals")}
              active={activeKey === "compliancePortal"}
            />
          )}
          {isAccessReviewVisible(permissions) && (
            <NavRailItem
              icon={navGroupByKey("accessReview").icon}
              label={t("nav.groups.accessReview")}
              to={accessReviewHref(organizationId, permissions)}
              active={activeKey === "accessReview"}
            />
          )}
          {isSettingsVisible(permissions, slackbotAvailable) && (
            <NavRailItem
              icon={navGroupByKey("settings").icon}
              label={t("nav.groups.settings")}
              to={settingsHref(organizationId, permissions, slackbotAvailable)}
              active={activeKey === "settings"}
            />
          )}
        </div>

        <NavRailItem
          icon={displayMode === "dark" ? SunIcon : MoonIcon}
          label={
            displayMode === "dark"
              ? t("nav.switchToLightMode")
              : t("nav.switchToDarkMode")
          }
          onClick={toggleDisplayMode}
          weight="regular"
        />
        <NavRailItem
          icon={LifebuoyIcon}
          label={t("nav.help")}
          href="mailto:support@probo.com"
          weight="regular"
        />
        <ViewerMembershipMenu organizationKey={organization} />
      </nav>
    </div>
  );
}
