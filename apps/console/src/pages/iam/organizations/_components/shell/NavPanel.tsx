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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { NavPanel_organization$key } from "#/__generated__/iam/NavPanel_organization.graphql";
import { NAV_GROUPS } from "#/pages/iam/organizations/_lib/navigation";
import { useActiveNavGroup } from "#/pages/iam/organizations/_lib/useActiveNavGroup";

import { navPanels } from "./navPanels";
import { navPanel } from "./variants";

// The body is chosen from navPanels rather than rendered by name here, which
// the colocation rule cannot follow.
/* eslint-disable relay/must-colocate-fragment-spreads */
const navPanelFragment = graphql`
  fragment NavPanel_organization on Organization {
    ...AccessReviewNavPanel_organization
    ...CompliancePortalNavPanel_organization
    ...GovernanceNavPanel_organization
    ...ItamNavPanel_organization
    ...PrivacyNavPanel_organization
    ...RegistriesNavPanel_organization
    ...RiskManagementNavPanel_organization
    ...SettingsNavPanel_organization
    ...TprmNavPanel_organization
  }
`;

export interface NavPanelProps {
  organizationKey: NavPanel_organization$key;
}

export function NavPanel({ organizationKey }: NavPanelProps) {
  const { t } = useTranslation();
  const organization = useFragment(navPanelFragment, organizationKey);
  const activeGroup = useActiveNavGroup(NAV_GROUPS);
  const slots = navPanel();

  if (activeGroup == null) {
    return <aside className={slots.panel()} />;
  }

  const Body = navPanels[activeGroup.key];

  return (
    <aside className={slots.panel()}>
      <Text size={2} weight="medium" color="faint" className={slots.title()}>
        {t(`nav.groups.${activeGroup.key}`)}
      </Text>
      <div className={slots.list()}>
        <Body organizationKey={organization} group={activeGroup} />
      </div>
    </aside>
  );
}
