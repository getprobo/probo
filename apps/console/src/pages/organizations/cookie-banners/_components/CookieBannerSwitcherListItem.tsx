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

import { Menu } from "@base-ui/react/menu";
import { CheckIcon } from "@phosphor-icons/react";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { graphql, useFragment } from "react-relay";
import { Link, useParams } from "react-router";

import type { CookieBannerSwitcherListItem_cookieBanner$key } from "#/__generated__/core/CookieBannerSwitcherListItem_cookieBanner.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { cookieBannerSwitcher } from "#/pages/iam/organizations/_components/shell/variants";

const cookieBannerSwitcherListItemFragment = graphql`
  fragment CookieBannerSwitcherListItem_cookieBanner on CookieBanner {
    id
    name
    origin
  }
`;

export interface CookieBannerSwitcherListItemProps {
  cookieBannerKey: CookieBannerSwitcherListItem_cookieBanner$key;
}

/**
 * One cookie banner in the privacy-panel switcher.
 *
 * Name and origin stack because a single truncated line cannot tell two
 * banners on similar hosts apart. The check marks the banner the URL is on.
 */
export function CookieBannerSwitcherListItem({ cookieBannerKey }: CookieBannerSwitcherListItemProps) {
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const cookieBanner = useFragment(cookieBannerSwitcherListItemFragment, cookieBannerKey);
  const selected = cookieBannerId === cookieBanner.id;
  const slots = cookieBannerSwitcher();

  return (
    <Menu.Item
      className={slots.item()}
      render={(
        <Link
          to={`/organizations/${organizationId}/privacy/cookie-banners/${cookieBanner.id}`}
        />
      )}
    >
      <span className={slots.itemBody()}>
        <Text size={2} weight="medium" color="neutral" highContrast className={slots.itemName()}>
          {cookieBanner.name}
        </Text>
        <Text size={1} color="faint" className={slots.itemOrigin()}>
          {cookieBanner.origin}
        </Text>
      </span>
      {selected && <CheckIcon className={slots.itemCheck()} />}
    </Menu.Item>
  );
}
