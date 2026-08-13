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

import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { NavRail_organization$key } from "#/__generated__/iam/NavRail_organization.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navGroupLandingItem, navItemHref, visibleNavGroups } from "#/pages/iam/organizations/_lib/navigation";
import { useActiveNavGroup } from "#/pages/iam/organizations/_lib/useActiveNavGroup";

import { NavRailItem } from "./NavRailItem";
import { OrganizationSwitcher } from "./OrganizationSwitcher";
import { navRail } from "./variants";
import { ViewerMembershipMenu } from "./ViewerMembershipMenu";

// Every permission below is read, but through the `permission` key on each
// entry of the nav table rather than by name here, which the rule cannot
// follow. See visibleNavGroups in _lib/navigation.ts.
/* eslint-disable relay/unused-fields */
const navRailFragment = graphql`
  fragment NavRail_organization on Organization {
    ...OrganizationSwitcher_organization
    ...ViewerMembershipMenu_organization

    canGetContext: permission(action: "core:organization-context:get")
    canListTasks: permission(action: "core:task:list")
    canListMeasures: permission(action: "core:measure:list")
    canListRisks: permission(action: "core:risk:list")
    canListFrameworks: permission(action: "core:framework:list")
    canListMembers: permission(action: "iam:membership:list")
    canListThirdParties: permission(action: "core:thirdParty:list")
    canListDocuments: permission(action: "core:document:list")
    canListAssets: permission(action: "core:asset:list")
    canListDevices: permission(action: "itam:device:list")
    canListData: permission(action: "core:datum:list")
    canListAudits: permission(action: "core:audit:list")
    canListFindings: permission(action: "core:finding:list")
    canListBusinessFunctions: permission(action: "core:business-function:list")
    canListAiSystems: permission(action: "core:ai-system:list")
    canListObligations: permission(action: "core:obligation:list")
    canListProcessingActivities: permission(action: "core:processing-activity:list")
    canListStatementsOfApplicability: permission(action: "core:statement-of-applicability:list")
    canListRightsRequests: permission(action: "core:rights-request:list")
    canListAccessReviewCampaigns: permission(action: "access-review:campaign:list")
    canGetCompliancePortal: permission(action: "compliance-portal:portal:get")
    canListCookieBanners: permission(action: "core:cookie-banner:list")
    canUpdateOrganization: permission(action: "iam:organization:update")
  }
`;

export interface NavRailProps {
  organizationKey: NavRail_organization$key;
}

/**
 * The whole chrome for an organization: which organization you are in, the
 * products, and who you are signed in as, all naming themselves when the rail
 * is hovered or focused.
 *
 * Each product icon links to the first link of its product (switchers have no
 * index route), so the rail alone is enough to move around; the panel beside
 * it then offers the rest.
 */
export function NavRail({ organizationKey }: NavRailProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const organization = useFragment(navRailFragment, organizationKey);

  const groups = useMemo(() => visibleNavGroups(organization), [organization]);
  const activeGroup = useActiveNavGroup(groups);

  const slots = navRail();

  return (
    <div className={slots.root()}>
      <nav className={slots.rail()} aria-label={t("nav.products")}>
        <OrganizationSwitcher variant="rail" organizationKey={organization} />

        <div className={slots.items()}>
          {groups.map(group => (
            <NavRailItem
              key={group.key}
              icon={group.icon}
              label={t(`nav.groups.${group.key}`)}
              to={navItemHref(organizationId, group, navGroupLandingItem(group))}
              active={group.key === activeGroup?.key}
            />
          ))}
        </div>

        <ViewerMembershipMenu variant="rail" organizationKey={organization} />
      </nav>
    </div>
  );
}
