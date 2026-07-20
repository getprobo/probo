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
import { Link } from "@probo/ui/src/v2/Button/Link";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link as RouterLink, useLocation } from "react-router";

import { buildRequestAllContinueUrl, redirectToInitiate } from "#/lib/auth/continueUrl";

import type { TopBar_query$key } from "./__generated__/TopBar_query.graphql";
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
      title
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

  const { currentCompliancePortal } = data;
  const title = currentCompliancePortal.title;
  const logoUrl = currentCompliancePortal.themedLogoUrl ?? undefined;

  const slots = topBar();

  return (
    <header className={slots.bar()}>
      <div className={slots.inner()}>
        <RouterLink to="/" className={slots.brand()}>
          <Avatar
            size={1}
            variant="soft"
            color="neutral"
            radius="small"
            src={logoUrl}
            fallback={title.charAt(0) || "?"}
            className={slots.logo()}
          />
          <Text size={2} weight="medium" color="neutral" highContrast className={slots.brandName()}>
            {title}
          </Text>
          <Text size={2} color="neutral" className={slots.tagline()}>
            {t("topBar.tagline")}
          </Text>
        </RouterLink>

        <div className={slots.spacer()} />

        <nav className={slots.nav()}>
          {TOP_BAR_NAV_ITEMS.map(item => (
            <Link
              key={item.to}
              to={item.to}
              variant="ghost"
              color="neutral"
              size={2}
              active={isActive(pathname, item.to)}
            >
              {t(item.labelKey)}
            </Link>
          ))}
          {data.viewer == null
            ? (
                <Button
                  variant="solid"
                  color="neutral"
                  highContrast
                  iconStart={<LockSimpleIcon />}
                  onClick={() => {
                    redirectToInitiate(buildRequestAllContinueUrl());
                  }}
                >
                  {t("topBar.getAccess")}
                </Button>
              )
            : <TopBarUserMenu identityKey={data.viewer} />}
        </nav>

        <TopBarMobileNav identityKey={data.viewer ?? null} />
      </div>
    </header>
  );
}
