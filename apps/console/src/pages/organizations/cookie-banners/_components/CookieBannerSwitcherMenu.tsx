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

import { CaretDownIcon, PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePaginationFragment, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";

import type { CookieBannerSwitcherMenu_organization$key } from "#/__generated__/core/CookieBannerSwitcherMenu_organization.graphql";
import type { CookieBannerSwitcherMenuQuery } from "#/__generated__/core/CookieBannerSwitcherMenuQuery.graphql";
import type { CookieBannerSwitcherMenuRefetchQuery } from "#/__generated__/core/CookieBannerSwitcherMenuRefetchQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navPanelSwitcher } from "#/pages/organizations/_components/NavPanelSwitcher";

import { cookieBannersBasePath } from "../_lib/cookieBannerPaths";

import { CookieBannerSwitcherListItem } from "./CookieBannerSwitcherListItem";

export const cookieBannerSwitcherMenuQuery = graphql`
  query CookieBannerSwitcherMenuQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ...CookieBannerSwitcherMenu_organization
    }
  }
`;

const cookieBannerSwitcherMenuFragment = graphql`
  fragment CookieBannerSwitcherMenu_organization on Organization
  @argumentDefinitions(
    first: { type: Int, defaultValue: 50 }
    after: { type: CursorKey, defaultValue: null }
  )
  @refetchable(queryName: "CookieBannerSwitcherMenuRefetchQuery") {
    canCreateCookieBanner: permission(action: "core:cookie-banner:create")
    cookieBanners(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
    )
      @connection(key: "CookieBannerSwitcherMenu_cookieBanners", filters: [])
      @required(action: THROW) {
      edges {
        node {
          id
          ...CookieBannerSwitcherListItem_cookieBanner
        }
      }
    }
  }
`;

export interface CookieBannerSwitcherMenuProps {
  queryRef: PreloadedQuery<CookieBannerSwitcherMenuQuery>;
  selectedId: string | null;
}

export function CookieBannerSwitcherMenu({ queryRef, selectedId }: CookieBannerSwitcherMenuProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const slots = navPanelSwitcher();

  const { organization } = usePreloadedQuery<CookieBannerSwitcherMenuQuery>(
    cookieBannerSwitcherMenuQuery,
    queryRef,
  );
  const organizationKey: CookieBannerSwitcherMenu_organization$key | null
    = organization.__typename === "Organization" ? organization : null;
  const { data, hasNext, loadNext, isLoadingNext } = usePaginationFragment<
    CookieBannerSwitcherMenuRefetchQuery,
    CookieBannerSwitcherMenu_organization$key
  >(cookieBannerSwitcherMenuFragment, organizationKey);
  if (data == null) {
    throw new Error("invalid type for node");
  }

  const banners = data.cookieBanners.edges.map(edge => edge.node);
  // The switcher value falls back to the most recent banner when the selection
  // is stale, so the check mark has to follow. A selection missing from the
  // loaded pages only counts as stale once the whole list is loaded.
  const isSelectionLoaded = banners.some(banner => banner.id === selectedId);
  const checkedId = selectedId != null && (isSelectionLoaded || hasNext)
    ? selectedId
    : banners[0]?.id;
  const newBannerHref = `${cookieBannersBasePath(organizationId)}/new`;

  return (
    <>
      <div className={slots.list()}>
        {banners.length === 0
          ? (
              <Text size={2} color="faint" className={slots.empty()}>
                {t("nav.cookieBannerSwitcher.empty")}
              </Text>
            )
          : banners.map(banner => (
              <CookieBannerSwitcherListItem
                key={banner.id}
                cookieBannerKey={banner}
                selected={checkedId === banner.id}
              />
            ))}
        {hasNext && (
          <Button
            variant="ghost"
            color="neutral"
            size={2}
            loading={isLoadingNext}
            iconStart={<CaretDownIcon />}
            className={slots.more()}
            onClick={() => {
              loadNext(50);
            }}
          >
            {t("nav.cookieBannerSwitcher.showMore")}
          </Button>
        )}
      </div>
      {data.canCreateCookieBanner && (
        <>
          <DropdownSeparator />
          <DropdownItem
            iconStart={<PlusIcon />}
            render={<Link to={newBannerHref} />}
          >
            {t("nav.cookieBannerSwitcher.create")}
          </DropdownItem>
        </>
      )}
    </>
  );
}
