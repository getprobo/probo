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

import { CaretDownIcon } from "@phosphor-icons/react";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useQueryLoader } from "react-relay";
import { useLocation, useParams } from "react-router";

import type { CompliancePortalSwitcherMenuQuery } from "#/__generated__/core/CompliancePortalSwitcherMenuQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { CompliancePortalNavItems } from "./CompliancePortalNavItems";
import {
  CompliancePortalSwitcherMenu,
  compliancePortalSwitcherMenuQuery,
} from "./CompliancePortalSwitcherMenu";
import {
  CompliancePortalSwitcherValue,
  CompliancePortalSwitcherValueSkeleton,
} from "./CompliancePortalSwitcherValue";
import { compliancePortalSwitcher } from "./variants";

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
    <>
      <CoreRelayProvider>
        <CompliancePortalSwitcherInner />
      </CoreRelayProvider>
      <CompliancePortalNavItems />
    </>
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
  // Gold only on /new: once a portal is selected, the in-page tabs carry the
  // accent and the trigger stays an outline picker.
  const active = isNew;
  const selectLabel = t("nav.compliancePortalSwitcher.select");
  const newLabel = t("nav.compliancePortalSwitcher.new");

  const handleOpenChange = useCallback((open: boolean) => {
    if (open) {
      loadQuery({ organizationId });
    }
  }, [loadQuery, organizationId]);

  const slots = compliancePortalSwitcher({ outlined: !active, active });

  return (
    <Dropdown onOpenChange={handleOpenChange}>
      <DropdownTrigger
        render={(
          <button type="button" className={slots.trigger()}>
            <span className={slots.value()}>
              {compliancePortalId != null
                ? (
                    <Suspense fallback={<CompliancePortalSwitcherValueSkeleton />}>
                      <CompliancePortalSwitcherValue fallback={selectLabel} />
                    </Suspense>
                  )
                : (
                    <Text size={2} weight="medium" color="neutral" highContrast className={slots.itemName()}>
                      {isNew ? newLabel : selectLabel}
                    </Text>
                  )}
            </span>
            <CaretDownIcon className={slots.valueCaret()} />
          </button>
        )}
      />
      <DropdownPopup side="bottom" align="start" className={slots.popup()}>
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
      </DropdownPopup>
    </Dropdown>
  );
}
