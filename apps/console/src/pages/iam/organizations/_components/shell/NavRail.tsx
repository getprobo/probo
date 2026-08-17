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
import { navGroupHref, visibleNavGroups } from "#/pages/iam/organizations/_lib/navigation";
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
}

export function NavRail({ organizationKey }: NavRailProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const organization = useFragment(navRailFragment, organizationKey);
  const { displayMode, toggleDisplayMode } = useDisplayMode();

  const permissions = useFragment<navPermissions_organization$key>(
    navPermissionsFragment,
    organization,
  );
  const groups = useMemo(() => visibleNavGroups(permissions), [permissions]);
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
              to={navGroupHref(organizationId, group, permissions)}
              active={group.key === activeGroup?.key}
            />
          ))}
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
        <ViewerMembershipMenu variant="rail" organizationKey={organization} />
      </nav>
    </div>
  );
}
