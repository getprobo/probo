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

import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { TopBar_organization$key } from "#/__generated__/iam/TopBar_organization.graphql";

import { TopBarUserMenu } from "./TopBarUserMenu";
import { topBar } from "./variants";

const topBarFragment = graphql`
  fragment TopBar_organization on Organization {
    id
    name
    logo {
      downloadUrl
    }
    ...TopBarUserMenu_organization
  }
`;

interface TopBarProps {
  organizationKey: TopBar_organization$key;
}

export function TopBar({ organizationKey }: TopBarProps) {
  const { t } = useTranslation();
  const organization = useFragment(topBarFragment, organizationKey);
  const slots = topBar();
  const tagline = t("topBar.tagline");

  return (
    <header className={slots.bar()}>
      <div className={slots.inner()}>
        <Link
          to={`/${organization.id}`}
          className={slots.brand()}
          aria-label={`${organization.name} ${tagline}`}
        >
          <Avatar
            size={2}
            variant="soft"
            color="neutral"
            radius="small"
            src={organization.logo?.downloadUrl ?? undefined}
            fallback={organization.name.charAt(0) || "?"}
            className={slots.logo()}
          />
          <span className={slots.brandText()}>
            <Text size={2} weight="medium" color="neutral" highContrast className={slots.brandName()}>
              {organization.name}
            </Text>
            <Text size={1} color="neutral" className={slots.tagline()}>
              {tagline}
            </Text>
          </span>
        </Link>

        <TopBarUserMenu organizationKey={organization} />
      </div>
    </header>
  );
}
