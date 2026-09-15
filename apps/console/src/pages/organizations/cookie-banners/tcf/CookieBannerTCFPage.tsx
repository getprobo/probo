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

import { ListSkeleton } from "@probo/ui/src/v2/List/ListSkeleton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerTCFPage_cookieBanner$key } from "#/__generated__/core/CookieBannerTCFPage_cookieBanner.graphql";
import type { CookieBannerTCFPageQuery } from "#/__generated__/core/CookieBannerTCFPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { GVLVendorList } from "./_components/GVLVendorList";
import { GVLVendorListSearch } from "./_components/GVLVendorListSearch";
import { tcfPage, tcfSection } from "./variants";

export const cookieBannerTCFPageQuery = graphql`
  query CookieBannerTCFPageQuery($cookieBannerId: ID!, $query: String) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        capabilities {
          tcf
        }
        ...CookieBannerTCFPage_cookieBanner
        ...GVLVendorList_cookieBanner
      }
    }
    ...GVLVendorList_query @arguments(first: 25, query: $query)
  }
`;

const cookieBannerFragment = graphql`
  fragment CookieBannerTCFPage_cookieBanner on CookieBanner {
    gvlVendors(first: 500) {
      totalCount
    }
  }
`;

interface CookieBannerTCFPageProps {
  queryRef: PreloadedQuery<CookieBannerTCFPageQuery>;
}

export function CookieBannerTCFPage({ queryRef }: CookieBannerTCFPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const data = usePreloadedQuery<CookieBannerTCFPageQuery>(cookieBannerTCFPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  if (!data.node.capabilities.tcf) {
    throw new NotFoundError(t("tcfPage.notFound"));
  }

  const banner = useFragment<CookieBannerTCFPage_cookieBanner$key>(
    cookieBannerFragment,
    data.node,
  );
  const selectedCount = banner.gvlVendors?.totalCount ?? 0;
  const { root, intro, tools } = tcfSection();

  return (
    <div className={tcfPage()}>
      <section className={root()}>
        <div className={intro()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("tcfPage.title")}
          </Heading>
          <Text size={2} color="faint">
            {t("tcfPage.description")}
          </Text>
          <Text size={2} color="neutral">
            {t("tcfPage.vendorCount", { count: selectedCount })}
          </Text>
          <div className={tools()}>
            <GVLVendorListSearch />
          </div>
        </div>
        <Suspense fallback={<ListSkeleton count={4} />}>
          <GVLVendorList
            queryKey={data}
            cookieBannerKey={data.node}
          />
        </Suspense>
      </section>
    </div>
  );
}
