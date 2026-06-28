// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { LockSimpleIcon } from "@phosphor-icons/react";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Button/Link";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link as RouterLink, useLocation } from "react-router";

import type { TopBar_query$key } from "./__generated__/TopBar_query.graphql";
import { TopBarUserMenu } from "./TopBarUserMenu";
import { topBar } from "./variants";

const NAV_ITEMS = [
  { to: "/documents", labelKey: "topBar.nav.documents" },
  { to: "/subprocessors", labelKey: "topBar.nav.subprocessors" },
  { to: "/updates", labelKey: "topBar.nav.updates" },
  { to: "/requests", labelKey: "topBar.nav.requests" },
] as const;

const topBarFragment = graphql`
  fragment TopBar_query on Query {
    viewer {
      ...TopBarUserMenu_identity
    }
    currentTrustCenter @required(action: THROW) {
      themedLogoUrl
      organization {
        name
      }
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
  const { pathname } = useLocation();

  const { currentTrustCenter } = data;
  const organizationName = currentTrustCenter.organization.name;
  const logoUrl = currentTrustCenter.themedLogoUrl ?? undefined;

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
            fallback={organizationName.charAt(0) || "?"}
            className={slots.logo()}
          />
          <Text size={2} weight="medium" color="neutral" highContrast>
            {organizationName}
          </Text>
          <Text size={2} color="neutral">
            {t("topBar.tagline")}
          </Text>
        </RouterLink>

        <div className={slots.spacer()} />

        <nav className={slots.nav()}>
          {NAV_ITEMS.map(item => (
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
                <Button variant="solid" color="neutral" highContrast iconStart={<LockSimpleIcon />}>
                  {t("topBar.getAccess")}
                </Button>
              )
            : <TopBarUserMenu identityKey={data.viewer} />}
        </nav>
      </div>
    </header>
  );
}
