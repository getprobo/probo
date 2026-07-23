// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { CookieIcon } from "@phosphor-icons/react";
import {
  IconBank,
  IconBook,
  IconBox,
  IconCircleProgress,
  IconFire3,
  IconGroup1,
  IconInboxEmpty,
  IconKey,
  IconListStack,
  IconLock,
  IconMagnifyingGlass,
  IconMedal,
  IconPageCheck,
  IconPageTextLine,
  IconPageTextSolid,
  IconSettingsGear2,
  IconShield,
  IconStore,
  IconTodo,
  SidebarItem,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SidebarFragment$key } from "#/__generated__/iam/SidebarFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
    fragment SidebarFragment on Organization {
        canGetContext: permission(action: "core:organization-context:get")
        canListTasks: permission(action: "core:task:list")
        canListMeasures: permission(action: "core:measure:list")
        canListRisks: permission(action: "core:risk:list")

        canListFrameworks: permission(action: "core:framework:list")
        canListMembers: permission(action: "iam:membership:list")
        canListThirdParties: permission(action: "core:thirdParty:list")
        canListDocuments: permission(action: "core:document:list")
        canListAssets: permission(action: "core:asset:list")
        canListData: permission(action: "core:datum:list")
        canListAudits: permission(action: "core:audit:list")
        canListFindings: permission(action: "core:finding:list")
        canListObligations: permission(action: "core:obligation:list")
        canListProcessingActivities: permission(
            action: "core:processing-activity:list"
        )
        canListRightsRequests: permission(action: "core:rights-request:list")
        canGetCompliancePage: permission(action: "compliance-portal:portal:get")
        canListCookieBanners: permission(action: "core:cookie-banner:list")
        canUpdateOrganization: permission(action: "iam:organization:update")
        canListStatementsOfApplicability: permission(
            action: "core:statement-of-applicability:list"
        )
        canListAccessReviewCampaigns: permission(
            action: "access-review:campaign:list"
        )
    }
`;

export function Sidebar(props: { fKey: SidebarFragment$key }) {
  const { fKey } = props;

  const { t } = useTranslation();
  const organizationId = useOrganizationId();

  const organization = useFragment<SidebarFragment$key>(fragment, fKey);

  const prefix = `/organizations/${organizationId}`;

  return (
    <ul className="space-y-[2px]">
      {organization.canGetContext && (
        <SidebarItem
          label={t("sidebar.context")}
          icon={IconPageTextSolid}
          to={`${prefix}/context`}
        />
      )}
      {organization.canListTasks && (
        <SidebarItem
          label={t("sidebar.tasks")}
          icon={IconInboxEmpty}
          to={`${prefix}/tasks`}
        />
      )}
      {organization.canListMeasures && (
        <SidebarItem
          label={t("sidebar.measures")}
          icon={IconTodo}
          to={`${prefix}/measures`}
        />
      )}
      {organization.canListRisks && (
        <SidebarItem
          label={t("sidebar.risks")}
          icon={IconFire3}
          to={`${prefix}/risks`}
        />
      )}

      {organization.canListFrameworks && (
        <SidebarItem
          label={t("sidebar.frameworks")}
          icon={IconBank}
          to={`${prefix}/frameworks`}
        />
      )}
      {organization.canListMembers && (
        <SidebarItem
          label={t("sidebar.people")}
          icon={IconGroup1}
          to={`${prefix}/people`}
        />
      )}
      {organization.canListThirdParties && (
        <SidebarItem
          label={t("sidebar.thirdParties")}
          icon={IconStore}
          to={`${prefix}/third-parties`}
        />
      )}
      {organization.canListDocuments && (
        <SidebarItem
          label={t("sidebar.documents")}
          icon={IconPageTextLine}
          to={`${prefix}/documents`}
        />
      )}
      {organization.canListAssets && (
        <SidebarItem
          label={t("sidebar.assets")}
          icon={IconBox}
          to={`${prefix}/assets`}
        />
      )}
      {organization.canListData && (
        <SidebarItem
          label={t("sidebar.data")}
          icon={IconListStack}
          to={`${prefix}/data`}
        />
      )}
      {organization.canListAudits && (
        <SidebarItem
          label={t("sidebar.audits")}
          icon={IconMedal}
          to={`${prefix}/audits`}
        />
      )}
      {organization.canListFindings && (
        <SidebarItem
          label={t("sidebar.findings")}
          icon={IconMagnifyingGlass}
          to={`${prefix}/findings`}
        />
      )}
      {organization.canListObligations && (
        <SidebarItem
          label={t("sidebar.obligations")}
          icon={IconBook}
          to={`${prefix}/obligations`}
        />
      )}
      {organization.canListProcessingActivities && (
        <SidebarItem
          label={t("sidebar.processingActivities")}
          icon={IconCircleProgress}
          to={`${prefix}/processing-activities`}
        />
      )}
      {organization.canListStatementsOfApplicability && (
        <SidebarItem
          label={t("sidebar.statementsOfApplicability")}
          icon={IconPageCheck}
          to={`${prefix}/statements-of-applicability`}
        />
      )}
      {organization.canListRightsRequests && (
        <SidebarItem
          label={t("sidebar.rightsRequests")}
          icon={IconLock}
          to={`${prefix}/rights-requests`}
        />
      )}
      {organization.canListAccessReviewCampaigns && (
        <SidebarItem
          label={t("sidebar.accessReviews")}
          icon={IconKey}
          to={`${prefix}/access-reviews`}
        />
      )}
      {organization.canGetCompliancePage && (
        <SidebarItem
          label={t("sidebar.compliancePage")}
          icon={IconShield}
          to={`${prefix}/compliance-page`}
        />
      )}
      {organization.canListCookieBanners && (
        <SidebarItem
          label={t("sidebar.cookieBanners")}
          icon={CookieIcon}
          to={`${prefix}/cookie-banners`}
        />
      )}
      {organization.canUpdateOrganization && (
        <SidebarItem
          label={t("sidebar.settings")}
          icon={IconSettingsGear2}
          to={`${prefix}/settings`}
        />
      )}
    </ul>
  );
}
