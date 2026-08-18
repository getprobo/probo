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
import { useLocation } from "react-router";

import type { CookieBannerSwitcherMenuQuery } from "#/__generated__/core/CookieBannerSwitcherMenuQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  navPanelSwitcher,
  NavPanelSwitcher,
  NavPanelSwitcherValue,
} from "#/pages/organizations/_components/NavPanelSwitcher";

import { cookieBannersBasePath } from "../_lib/cookieBannerPaths";
import type { SelectedCookieBanner } from "../_lib/useSelectedCookieBanner";

import {
  CookieBannerSwitcherMenu,
  cookieBannerSwitcherMenuQuery,
} from "./CookieBannerSwitcherMenu";
import { CookieBannerSwitcherValue } from "./CookieBannerSwitcherValue";

export interface CookieBannerSwitcherProps {
  banner: SelectedCookieBanner | null;
}

export function CookieBannerSwitcher({ banner }: CookieBannerSwitcherProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { pathname } = useLocation();
  const [menuQueryRef, loadMenuQuery] = useQueryLoader<CookieBannerSwitcherMenuQuery>(
    cookieBannerSwitcherMenuQuery,
  );

  const isNew = pathname === `${cookieBannersBasePath(organizationId)}/new`;
  const selectLabel = t("nav.cookieBannerSwitcher.select");
  const newLabel = t("nav.cookieBannerSwitcher.new");
  const triggerLabel = t("nav.cookieBannerSwitcher.label");
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
          <CookieBannerSwitcherMenu queryRef={menuQueryRef} selectedId={banner?.id ?? null} />
        </Suspense>
      )
    : null;

  if (isNew) {
    return (
      <NavPanelSwitcher
        active
        label={triggerLabel}
        onOpenChange={handleOpenChange}
        value={<NavPanelSwitcherValue>{newLabel}</NavPanelSwitcherValue>}
      >
        {menu}
      </NavPanelSwitcher>
    );
  }

  return (
    <NavPanelSwitcher
      active={false}
      label={triggerLabel}
      onOpenChange={handleOpenChange}
      value={<CookieBannerSwitcherValue fallback={selectLabel} name={banner?.name ?? null} />}
    >
      {menu}
    </NavPanelSwitcher>
  );
}
