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
import { Suspense, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useQueryLoader } from "react-relay";
import { useLocation, useParams } from "react-router";

import type { CompliancePortalSwitcherMenuQuery } from "#/__generated__/core/CompliancePortalSwitcherMenuQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  navPanelSwitcher,
  NavPanelSwitcher,
  NavPanelSwitcherValue,
  NavPanelSwitcherValueSkeleton,
} from "#/pages/organizations/_components/NavPanelSwitcher";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePortalNavItems } from "./CompliancePortalNavItems";
import {
  CompliancePortalSwitcherMenu,
  compliancePortalSwitcherMenuQuery,
} from "./CompliancePortalSwitcherMenu";
import { CompliancePortalSwitcherValue } from "./CompliancePortalSwitcherValue";

/**
 * Product-panel control that picks a compliance portal instead of linking to
 * a list.
 *
 * Lives next to the compliance-portal routes (not in the IAM shell) so Relay
 * compiles the query against the core schema. The list is fetched on open:
 * the panel is visible for every page in this product, and most of those
 * visits never open this menu. The selected portal's name is a separate
 * query so the trigger can show it without loading the list.
 * CoreRelayProvider is local because the surrounding chrome runs on the IAM
 * environment. This product has no sibling panel entries, so NavPanel
 * skips the group heading and the product title is enough. Settings and
 * Pages hang off the same column once a portal is in the URL.
 */
export function CompliancePortalSwitcher() {
  return (
    <CoreRelayProvider>
      <CompliancePortalSwitcherInner />
      <CompliancePortalNavItems />
    </CoreRelayProvider>
  );
}

function CompliancePortalSwitcherInner() {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { pathname } = useLocation();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<CompliancePortalSwitcherMenuQuery>(
    compliancePortalSwitcherMenuQuery,
  );

  const prefix = `/organizations/${organizationId}/compliance-portals/`;
  const isNew = pathname === `${prefix}new`;
  const newLabel = t("nav.compliancePortalSwitcher.new");
  const slots = navPanelSwitcher();

  const handleOpenChange = useCallback((open: boolean) => {
    if (open) {
      loadQuery({ organizationId });
    }
  }, [loadQuery, organizationId]);

  return (
    <NavPanelSwitcher
      active={isNew}
      onOpenChange={handleOpenChange}
      value={isNew
        ? <NavPanelSwitcherValue>{newLabel}</NavPanelSwitcherValue>
        : (
            <Suspense fallback={<NavPanelSwitcherValueSkeleton />}>
              {compliancePortalId != null
                ? <CompliancePortalSwitcherValue />
                : <NavPanelSwitcherValueSkeleton />}
            </Suspense>
          )}
    >
      {queryRef != null && (
        <Suspense
          fallback={(
            <Text size={2} color="faint" className={slots.empty()}>
              {t("nav.compliancePortalSwitcher.loading")}
            </Text>
          )}
        >
          <CompliancePortalSwitcherMenu queryRef={queryRef} />
        </Suspense>
      )}
    </NavPanelSwitcher>
  );
}
