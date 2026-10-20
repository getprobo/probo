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

import { PlusIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannerTrackersPageQuery } from "#/__generated__/core/CookieBannerTrackersPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CookieBannerPageHeader } from "../../_components/CookieBannerPageHeader";
import { cookieBannerNewTrackerPath } from "../../_lib/cookieBannerPaths";
import { cookieBannerPage } from "../../variants";

import { TrackerPatternList } from "./_components/TrackerPatternList";

export const cookieBannerTrackersPageQuery = graphql`
  query CookieBannerTrackersPageQuery(
    $cookieBannerId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
    $filter: TrackerPatternFilter
    $order: TrackerPatternOrder
  ) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        canCreate: permission(action: "core:tracker-pattern:create")
        ...TrackerPatternList_cookieBanner
          @arguments(first: $first, after: $after, last: $last, before: $before, filter: $filter, order: $order)
      }
    }
  }
`;

interface CookieBannerTrackersPageProps {
  queryRef: PreloadedQuery<CookieBannerTrackersPageQuery>;
}

export function CookieBannerTrackersPage({
  queryRef,
}: CookieBannerTrackersPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const location = useLocation();
  const organizationId = useOrganizationId();
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const title = t("trackersPage.title");
  usePageTitle(title);
  const data = usePreloadedQuery<CookieBannerTrackersPageQuery>(
    cookieBannerTrackersPageQuery,
    queryRef,
  );

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  return (
    <div className={cookieBannerPage()}>
      <CookieBannerPageHeader
        title={title}
        description={t("trackersPage.description")}
        actions={data.node.canCreate && cookieBannerId != null
          ? (
              <ButtonLink
                to={{
                  pathname: cookieBannerNewTrackerPath(organizationId, cookieBannerId),
                  search: location.search,
                }}
                size={2}
                variant="solid"
                color="neutral"
                highContrast
                iconStart={<PlusIcon />}
              >
                {t("trackersPage.actions.new")}
              </ButtonLink>
            )
          : undefined}
      />
      <TrackerPatternList cookieBannerKey={data.node} />
    </div>
  );
}
