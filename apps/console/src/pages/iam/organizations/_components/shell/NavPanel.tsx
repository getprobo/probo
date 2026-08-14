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
import { Fragment, Suspense, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { NavPanel_organization$key } from "#/__generated__/iam/NavPanel_organization.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navItemHref, visibleNavGroups } from "#/pages/iam/organizations/_lib/navigation";
import { useActiveNavGroup } from "#/pages/iam/organizations/_lib/useActiveNavGroup";

import { NavPanelGroup } from "./NavPanelGroup";
import { NavPanelItem } from "./NavPanelItem";
import { navPanelSwitcher } from "./navPanelSwitchers";
import { navPanel } from "./variants";

// Every permission below is read, but through the `permission` key on each
// entry of the nav table rather than by name here, which the rule cannot
// follow. See visibleNavGroups in _lib/navigation.ts.
/* eslint-disable relay/unused-fields */
const navPanelFragment = graphql`
  fragment NavPanel_organization on Organization {
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
    canListAccessReviewSources: permission(action: "access-review:source:list")
    canGetCompliancePortal: permission(action: "compliance-portal:portal:get")
    canListCookieBanners: permission(action: "core:cookie-banner:list")
    canListAuditLogEntries: permission(action: "iam:audit-log-entry:list")
    canListWebhookSubscriptions: permission(action: "core:webhook-subscription:list")
    canUpdateOrganization: permission(action: "iam:organization:update")
  }
`;

export interface NavPanelProps {
  organizationKey: NavPanel_organization$key;
}

/**
 * The entries of whichever product the rail has selected.
 *
 * The column stays in place even for a product with a single entry: a panel
 * that appeared and vanished between products would shift the page under the
 * pointer on every move along the rail.
 */
export function NavPanel({ organizationKey }: NavPanelProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const permissions = useFragment(navPanelFragment, organizationKey);

  const groups = useMemo(() => visibleNavGroups(permissions), [permissions]);
  const activeGroup = useActiveNavGroup(groups);

  const slots = navPanel();

  return (
    <aside className={slots.panel()}>
      {activeGroup != null && (
        <>
          <Text size={2} weight="medium" color="faint" className={slots.title()}>
            {t(`nav.groups.${activeGroup.key}`)}
          </Text>
          <div className={slots.list()}>
            {activeGroup.items.map((item) => {
              if (item.kind === "section") {
                return (
                  <NavPanelGroup key={item.key} label={t(item.labelKey)}>
                    {item.items.map(child => (
                      <NavPanelItem
                        key={child.path}
                        label={t(child.labelKey)}
                        to={navItemHref(organizationId, activeGroup, child)}
                      />
                    ))}
                  </NavPanelGroup>
                );
              }

              if (item.kind === "switcher") {
                const Switcher = navPanelSwitcher(item.path);
                const switcher = (
                  <Suspense fallback={<span className={slots.groupFallback()} aria-hidden />}>
                    <Switcher />
                  </Suspense>
                );
                // A sole switcher already sits under the product title; a
                // second heading would only repeat it.
                if (activeGroup.items.length === 1) {
                  return <Fragment key={item.path}>{switcher}</Fragment>;
                }
                return (
                  <NavPanelGroup key={item.path} label={t(item.labelKey)}>
                    {switcher}
                  </NavPanelGroup>
                );
              }

              return (
                <NavPanelItem
                  key={item.path}
                  label={t(item.labelKey)}
                  to={navItemHref(organizationId, activeGroup, item)}
                />
              );
            })}
          </div>
        </>
      )}
    </aside>
  );
}
