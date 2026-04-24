// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
import {
  IconBank,
  IconBook,
  IconBox,
  IconCircleProgress,
  IconClock,
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
        canListSnapshots: permission(action: "core:snapshot:list")
        canGetTrustCenter: permission(action: "core:trust-center:get")
        # canListCookieBanners: permission(action: "core:cookie-banner:list")
        canUpdateOrganization: permission(action: "iam:organization:update")
        canListStatementsOfApplicability: permission(
            action: "core:statement-of-applicability:list"
        )
        canListAccessReviewCampaigns: permission(
            action: "core:access-review-campaign:list"
        )
    }
`;

export function Sidebar(props: { fKey: SidebarFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const organization = useFragment<SidebarFragment$key>(fragment, fKey);

  const prefix = `/organizations/${organizationId}`;

  return (
    <ul className="space-y-[2px]">
      {organization.canGetContext && (
        <SidebarItem
          label={__("Context")}
          icon={IconPageTextSolid}
          to={`${prefix}/context`}
        />
      )}
      {organization.canListTasks && (
        <SidebarItem
          label={__("Tasks")}
          icon={IconInboxEmpty}
          to={`${prefix}/tasks`}
        />
      )}
      {organization.canListMeasures && (
        <SidebarItem
          label={__("Measures")}
          icon={IconTodo}
          to={`${prefix}/measures`}
        />
      )}
      {organization.canListRisks && (
        <SidebarItem
          label={__("Risks")}
          icon={IconFire3}
          to={`${prefix}/risks`}
        />
      )}
      {organization.canListFrameworks && (
        <SidebarItem
          label={__("Frameworks")}
          icon={IconBank}
          to={`${prefix}/frameworks`}
        />
      )}
      {organization.canListMembers && (
        <SidebarItem
          label={__("People")}
          icon={IconGroup1}
          to={`${prefix}/people`}
        />
      )}
      {organization.canListThirdParties && (
        <SidebarItem
          label={__("Third Parties")}
          icon={IconStore}
          to={`${prefix}/third-parties`}
        />
      )}
      {organization.canListDocuments && (
        <SidebarItem
          label={__("Documents")}
          icon={IconPageTextLine}
          to={`${prefix}/documents`}
        />
      )}
      {organization.canListAssets && (
        <SidebarItem
          label={__("Assets")}
          icon={IconBox}
          to={`${prefix}/assets`}
        />
      )}
      {organization.canListData && (
        <SidebarItem
          label={__("Data")}
          icon={IconListStack}
          to={`${prefix}/data`}
        />
      )}
      {organization.canListAudits && (
        <SidebarItem
          label={__("Audits")}
          icon={IconMedal}
          to={`${prefix}/audits`}
        />
      )}
      {organization.canListFindings && (
        <SidebarItem
          label={__("Findings")}
          icon={IconMagnifyingGlass}
          to={`${prefix}/findings`}
        />
      )}
      {organization.canListObligations && (
        <SidebarItem
          label={__("Obligations")}
          icon={IconBook}
          to={`${prefix}/obligations`}
        />
      )}
      {organization.canListProcessingActivities && (
        <SidebarItem
          label={__("Processing Activities")}
          icon={IconCircleProgress}
          to={`${prefix}/processing-activities`}
        />
      )}
      {organization.canListStatementsOfApplicability && (
        <SidebarItem
          label={__("Statements of Applicability")}
          icon={IconPageCheck}
          to={`${prefix}/statements-of-applicability`}
        />
      )}
      {organization.canListRightsRequests && (
        <SidebarItem
          label={__("Rights Requests")}
          icon={IconLock}
          to={`${prefix}/rights-requests`}
        />
      )}
      {organization.canListSnapshots && (
        <SidebarItem
          label={__("Snapshots")}
          icon={IconClock}
          to={`${prefix}/snapshots`}
        />
      )}
      {organization.canListAccessReviewCampaigns && (
        <SidebarItem
          label={__("Access Reviews")}
          icon={IconKey}
          to={`${prefix}/access-reviews`}
        />
      )}
      {organization.canGetTrustCenter && (
        <SidebarItem
          label={__("Compliance Page")}
          icon={IconShield}
          to={`${prefix}/compliance-page`}
        />
      )}
      {/* {organization.canListCookieBanners && (
        <SidebarItem
          label={__("Cookie Banners")}
          icon={CookieIcon}
          to={`${prefix}/cookie-banners`}
        />
      )} */}
      {organization.canUpdateOrganization && (
        <SidebarItem
          label={__("Settings")}
          icon={IconSettingsGear2}
          to={`${prefix}/settings`}
        />
      )}
    </ul>
  );
}
