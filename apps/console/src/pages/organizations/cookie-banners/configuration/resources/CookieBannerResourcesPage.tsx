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

import { usePageTitle } from "@probo/hooks";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerResourcesPageQuery } from "#/__generated__/core/CookieBannerResourcesPageQuery.graphql";

import { CookieBannerPageHeader } from "../../_components/CookieBannerPageHeader";
import { cookieBannerPage } from "../../variants";

import { TrackerResourceList } from "./_components/TrackerResourceList";

export const cookieBannerResourcesPageQuery = graphql`
  query CookieBannerResourcesPageQuery(
    $cookieBannerId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
    $filter: TrackerResourceFilter
  ) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        ...TrackerResourceList_cookieBanner
          @arguments(first: $first, after: $after, last: $last, before: $before, filter: $filter)
      }
    }
  }
`;

interface CookieBannerResourcesPageProps {
  queryRef: PreloadedQuery<CookieBannerResourcesPageQuery>;
}

export function CookieBannerResourcesPage({
  queryRef,
}: CookieBannerResourcesPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const title = t("resourcesPage.title");
  usePageTitle(title);
  const data = usePreloadedQuery<CookieBannerResourcesPageQuery>(
    cookieBannerResourcesPageQuery,
    queryRef,
  );

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  return (
    <div className={cookieBannerPage()}>
      <CookieBannerPageHeader
        title={title}
        description={t("resourcesPage.description")}
      />
      <TrackerResourceList cookieBannerKey={data.node} />
    </div>
  );
}
