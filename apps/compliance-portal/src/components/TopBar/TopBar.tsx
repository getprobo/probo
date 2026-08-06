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

import { LockSimpleIcon } from "@phosphor-icons/react";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link as RouterLink, useLocation } from "react-router";

import { getSafeContinueUrl, redirectToInitiate } from "#/lib/auth/continueUrl";
import { useLocalizedPath } from "#/lib/i18n/useLocale";

import type { TopBar_query$key } from "./__generated__/TopBar_query.graphql";
import { LocaleSelect } from "./LocaleSelect";
import { TOP_BAR_NAV_ITEMS } from "./navItems";
import { TopBarMobileNav } from "./TopBarMobileNav";
import { TopBarUserMenu } from "./TopBarUserMenu";
import { topBar } from "./variants";

const topBarFragment = graphql`
  fragment TopBar_query on Query {
    viewer {
      ...TopBarUserMenu_identity
      ...TopBarMobileNav_identity
    }
    currentCompliancePortal @required(action: THROW) {
      themedLogoUrl
      entityName
      rightsRequestsEnabled
      ...TopBarMobileNav_compliancePortal
    }
  }
`;

interface TopBarProps {
  queryKey: TopBar_query$key;
}

function isActive(pathname: string, to: string): boolean {
  return pathname === to || pathname.startsWith(`${to}/`);
}

export function TopBar({ queryKey }: TopBarProps) {
  const { t } = useTranslation();
  const data = useFragment(topBarFragment, queryKey);
  const location = useLocation();
  const { pathname } = location;
  const localizedPath = useLocalizedPath();

  const { currentCompliancePortal } = data;
  const entityName = currentCompliancePortal.entityName;
  const logoUrl = currentCompliancePortal.themedLogoUrl ?? undefined;
  const navItems = TOP_BAR_NAV_ITEMS.filter(
    (item) => item.to !== "/requests" || currentCompliancePortal.rightsRequestsEnabled,
  );

  const slots = topBar();
  const homePath = localizedPath("/");

  return (
    <header className={slots.bar()}>
      <div className={slots.inner()}>
        <RouterLink
          to={homePath}
          className={slots.brand()}
          aria-label={`${entityName} ${t("topBar.tagline")}`}
        >
          <Avatar
            size={2}
            variant="soft"
            color="neutral"
            radius="small"
            src={logoUrl}
            fallback={entityName.charAt(0) || "?"}
            className={slots.logo()}
          />
          <span className={slots.brandText()}>
            <Text size={2} weight="medium" color="neutral" highContrast className={slots.brandName()}>
              {entityName}
            </Text>
            <Text size={1} color="neutral" className={slots.tagline()}>
              {t("topBar.tagline")}
            </Text>
          </span>
        </RouterLink>

        <nav className={slots.nav()}>
          {navItems.map((item) => {
            const to = localizedPath(item.to);
            return (
              <ButtonLink
                key={item.to}
                to={to}
                variant="ghost"
                color="neutral"
                size={2}
                active={isActive(pathname, to)}
              >
                {t(item.labelKey)}
              </ButtonLink>
            );
          })}
          {data.viewer == null
            ? (
                <>
                  <LocaleSelect />
                  <Button
                    variant="solid"
                    color="neutral"
                    highContrast
                    iconStart={<LockSimpleIcon />}
                    onClick={() => {
                      redirectToInitiate(getSafeContinueUrl(window.location.href));
                    }}
                  >
                    {t("topBar.getAccess")}
                  </Button>
                </>
              )
            : <TopBarUserMenu identityKey={data.viewer} />}
        </nav>

        <TopBarMobileNav
          identityKey={data.viewer ?? null}
          compliancePortalKey={currentCompliancePortal}
        />
      </div>
    </header>
  );
}
