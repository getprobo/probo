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
import { type PreloadedQuery, useQueryLoader } from "react-relay";
import { useLocation, useParams } from "react-router";

import type { CompliancePortalSwitcherMenuQuery } from "#/__generated__/core/CompliancePortalSwitcherMenuQuery.graphql";
import type { CompliancePortalSwitcherRowQuery } from "#/__generated__/core/CompliancePortalSwitcherRowQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  navPanelSwitcher,
  NavPanelSwitcher,
  NavPanelSwitcherValue,
  NavPanelSwitcherValueSkeleton,
} from "#/pages/organizations/_components/NavPanelSwitcher";

import {
  CompliancePortalSwitcherMenu,
  compliancePortalSwitcherMenuQuery,
} from "./CompliancePortalSwitcherMenu";
import { CompliancePortalSwitcherRow } from "./CompliancePortalSwitcherRow";

export interface CompliancePortalSwitcherProps {
  queryRef: PreloadedQuery<CompliancePortalSwitcherRowQuery> | null;
}

export function CompliancePortalSwitcher({ queryRef }: CompliancePortalSwitcherProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { pathname } = useLocation();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const [menuQueryRef, loadMenuQuery] = useQueryLoader<CompliancePortalSwitcherMenuQuery>(
    compliancePortalSwitcherMenuQuery,
  );

  const prefix = `/organizations/${organizationId}/compliance-portals/`;
  const isNew = pathname === `${prefix}new`;
  const newLabel = t("nav.compliancePortalSwitcher.new");
  const slots = navPanelSwitcher();

  const handleOpenChange = useCallback((open: boolean) => {
    if (open) {
      loadMenuQuery({ organizationId });
    }
  }, [loadMenuQuery, organizationId]);

  const menu = menuQueryRef != null
    ? (
        <Suspense
          fallback={(
            <Text size={2} color="faint" className={slots.empty()}>
              {t("nav.compliancePortalSwitcher.loading")}
            </Text>
          )}
        >
          <CompliancePortalSwitcherMenu queryRef={menuQueryRef} />
        </Suspense>
      )
    : null;

  if (compliancePortalId != null && !isNew) {
    if (queryRef == null) {
      return (
        <div className={slots.row()}>
          <div className={slots.rowTrigger()}>
            <NavPanelSwitcher
              active={false}
              onOpenChange={handleOpenChange}
              value={<NavPanelSwitcherValueSkeleton />}
            >
              {menu}
            </NavPanelSwitcher>
          </div>
        </div>
      );
    }

    return (
      <div className={slots.row()}>
        <Suspense
          fallback={(
            <div className={slots.rowTrigger()}>
              <NavPanelSwitcher
                active={false}
                onOpenChange={handleOpenChange}
                value={<NavPanelSwitcherValueSkeleton />}
              >
                {menu}
              </NavPanelSwitcher>
            </div>
          )}
        >
          <CompliancePortalSwitcherRow queryRef={queryRef} onOpenChange={handleOpenChange}>
            {menu}
          </CompliancePortalSwitcherRow>
        </Suspense>
      </div>
    );
  }

  return (
    <div className={slots.row()}>
      <div className={slots.rowTrigger()}>
        <NavPanelSwitcher
          active={isNew}
          onOpenChange={handleOpenChange}
          value={isNew
            ? <NavPanelSwitcherValue>{newLabel}</NavPanelSwitcherValue>
            : <NavPanelSwitcherValueSkeleton />}
        >
          {menu}
        </NavPanelSwitcher>
      </div>
    </div>
  );
}
