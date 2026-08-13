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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useQueryLoader } from "react-relay";
import { useLocation, useParams } from "react-router";

import type { CookieBannerSwitcherMenuQuery } from "#/__generated__/core/CookieBannerSwitcherMenuQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { cookieBannerSwitcher } from "#/pages/iam/organizations/_components/shell/variants";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  CookieBannerSwitcherMenu,
  cookieBannerSwitcherMenuQuery,
} from "./CookieBannerSwitcherMenu";
import { CookieBannerSwitcherValue } from "./CookieBannerSwitcherValue";

/**
 * Privacy-panel control that picks a cookie banner instead of linking to a
 * list.
 *
 * Lives next to the cookie-banner routes (not in the IAM shell) so Relay
 * compiles the query against the core schema. The list is fetched on open:
 * the panel is visible for every privacy page, and most of those visits never
 * open this menu. The selected banner's name is a separate query so the
 * trigger can show it without loading the list. CoreRelayProvider is local
 * because the surrounding chrome runs on the IAM environment. The group label
 * around this control belongs to NavPanelGroup.
 */
export function CookieBannerSwitcher() {
  return (
    <CoreRelayProvider>
      <CookieBannerSwitcherInner />
    </CoreRelayProvider>
  );
}

function CookieBannerSwitcherInner() {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { pathname } = useLocation();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<CookieBannerSwitcherMenuQuery>(
    cookieBannerSwitcherMenuQuery,
  );

  const prefix = `/organizations/${organizationId}/privacy/cookie-banners/`;
  const active = pathname.startsWith(prefix);
  const isNew = pathname === `${prefix}new`;
  const selectLabel = t("nav.cookieBannerSwitcher.select");
  const newLabel = t("nav.cookieBannerSwitcher.new");

  const handleOpenChange = useCallback((open: boolean) => {
    if (open) {
      loadQuery({ organizationId });
    }
  }, [loadQuery, organizationId]);

  const slots = cookieBannerSwitcher();

  return (
    <Dropdown onOpenChange={handleOpenChange}>
      <DropdownTrigger
        render={(
          <Button
            variant={active ? "soft" : "outline"}
            color={active ? "gold" : "neutral"}
            size={2}
            active={active}
            iconEnd={<CaretDownIcon />}
            className={slots.trigger()}
          >
            <span className={slots.value()}>
              {cookieBannerId != null
                ? (
                    <Suspense fallback={selectLabel}>
                      <CookieBannerSwitcherValue fallback={selectLabel} />
                    </Suspense>
                  )
                : (isNew ? newLabel : selectLabel)}
            </span>
          </Button>
        )}
      />
      <DropdownPopup side="bottom" align="start" className={slots.popup()}>
        {queryRef != null && (
          <Suspense
            fallback={(
              <Text size={2} color="faint" className={slots.empty()}>
                {t("nav.cookieBannerSwitcher.loading")}
              </Text>
            )}
          >
            <CookieBannerSwitcherMenu queryRef={queryRef} />
          </Suspense>
        )}
      </DropdownPopup>
    </Dropdown>
  );
}
