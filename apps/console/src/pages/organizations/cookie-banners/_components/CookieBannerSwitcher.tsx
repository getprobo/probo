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
import { type ReactNode, Suspense, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { useLocation } from "react-router";

import type { CookieBannerSwitcherMenuQuery } from "#/__generated__/core/CookieBannerSwitcherMenuQuery.graphql";
import type { CookieBannerSwitcherValueQuery } from "#/__generated__/core/CookieBannerSwitcherValueQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  navPanelSwitcher,
  NavPanelSwitcher,
  NavPanelSwitcherValue,
  NavPanelSwitcherValueSkeleton,
} from "#/pages/organizations/_components/NavPanelSwitcher";

import { useSelectedCookieBannerId } from "../_lib/useSelectedCookieBannerId";

import {
  CookieBannerSwitcherMenu,
  cookieBannerSwitcherMenuQuery,
} from "./CookieBannerSwitcherMenu";
import {
  cookieBannerFromSwitcherValueQuery,
  CookieBannerSwitcherValue,
  cookieBannerSwitcherValueQuery,
} from "./CookieBannerSwitcherValue";

export interface CookieBannerSwitcherProps {
  queryRef: PreloadedQuery<CookieBannerSwitcherValueQuery> | null;
}

export function CookieBannerSwitcher({ queryRef }: CookieBannerSwitcherProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { pathname } = useLocation();
  const selectedId = useSelectedCookieBannerId();
  const [menuQueryRef, loadMenuQuery] = useQueryLoader<CookieBannerSwitcherMenuQuery>(
    cookieBannerSwitcherMenuQuery,
  );

  const prefix = `/organizations/${organizationId}/privacy/cookie-banners/`;
  const isNew = pathname === `${prefix}new`;
  const selectLabel = t("nav.cookieBannerSwitcher.select");
  const newLabel = t("nav.cookieBannerSwitcher.new");
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
              {t("nav.cookieBannerSwitcher.loading")}
            </Text>
          )}
        >
          <CookieBannerSwitcherMenu queryRef={menuQueryRef} selectedId={selectedId} />
        </Suspense>
      )
    : null;

  if (isNew) {
    return (
      <NavPanelSwitcher
        active
        onOpenChange={handleOpenChange}
        value={<NavPanelSwitcherValue>{newLabel}</NavPanelSwitcherValue>}
      >
        {menu}
      </NavPanelSwitcher>
    );
  }

  if (queryRef == null) {
    return (
      <NavPanelSwitcher
        active={false}
        onOpenChange={handleOpenChange}
        value={<NavPanelSwitcherValueSkeleton />}
      >
        {menu}
      </NavPanelSwitcher>
    );
  }

  return (
    <Suspense
      fallback={(
        <NavPanelSwitcher
          active={false}
          onOpenChange={handleOpenChange}
          value={<NavPanelSwitcherValueSkeleton />}
        >
          {menu}
        </NavPanelSwitcher>
      )}
    >
      <CookieBannerSwitcherSelected
        fallback={selectLabel}
        onOpenChange={handleOpenChange}
        queryRef={queryRef}
      >
        {menu}
      </CookieBannerSwitcherSelected>
    </Suspense>
  );
}

interface CookieBannerSwitcherSelectedProps {
  fallback: string;
  onOpenChange: (open: boolean) => void;
  queryRef: PreloadedQuery<CookieBannerSwitcherValueQuery>;
  children?: ReactNode;
}

function CookieBannerSwitcherSelected({
  fallback,
  onOpenChange,
  queryRef,
  children,
}: CookieBannerSwitcherSelectedProps) {
  const data = usePreloadedQuery<CookieBannerSwitcherValueQuery>(
    cookieBannerSwitcherValueQuery,
    queryRef,
  );
  const banner = cookieBannerFromSwitcherValueQuery(data);

  return (
    <NavPanelSwitcher
      active={false}
      onOpenChange={onOpenChange}
      value={banner != null
        ? <CookieBannerSwitcherValue cookieBannerKey={banner} />
        : <NavPanelSwitcherValue>{fallback}</NavPanelSwitcherValue>}
    >
      {children}
    </NavPanelSwitcher>
  );
}
