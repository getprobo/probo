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
import { ListSkeleton } from "@probo/ui/src/v2/List/ListSkeleton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerTCFPageQuery } from "#/__generated__/core/CookieBannerTCFPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { GVLVendorList } from "./_components/GVLVendorList";
import { GVLVendorListFilter } from "./_components/GVLVendorListFilter";
import { GVLVendorListSearch } from "./_components/GVLVendorListSearch";
import { GVLVendorStats } from "./_components/GVLVendorStats";
import { tcfPage, tcfSection } from "./variants";

export const cookieBannerTCFPageQuery = graphql`
  query CookieBannerTCFPageQuery(
    $cookieBannerId: ID!
    $filter: CommonGVLVendorFilter
  ) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        capabilities {
          tcf
        }
        ...GVLVendorList_cookieBanner
        ...GVLVendorStats_cookieBanner
      }
    }
    ...GVLVendorList_query @arguments(first: 15, filter: $filter)
    ...GVLVendorStats_query
  }
`;

interface CookieBannerTCFPageProps {
  queryRef: PreloadedQuery<CookieBannerTCFPageQuery>;
}

export function CookieBannerTCFPage({ queryRef }: CookieBannerTCFPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const title = t("tcfPage.title");
  usePageTitle(title);
  const data = usePreloadedQuery<CookieBannerTCFPageQuery>(cookieBannerTCFPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  if (!data.node.capabilities.tcf) {
    throw new NotFoundError(t("tcfPage.notFound"));
  }

  const { root, intro, list, tools } = tcfSection();

  return (
    <div className={tcfPage()}>
      <section className={root()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {title}
          </Heading>
          <Text size={2} color="faint">
            {t("tcfPage.description")}
          </Text>
          <GVLVendorStats
            queryKey={data}
            cookieBannerKey={data.node}
          />
        </div>
        <div className={list()}>
          <div className={tools()}>
            <GVLVendorListSearch />
            <GVLVendorListFilter />
          </div>
          <Suspense fallback={<ListSkeleton count={4} />}>
            <GVLVendorList
              queryKey={data}
              cookieBannerKey={data.node}
            />
          </Suspense>
        </div>
      </section>
    </div>
  );
}
