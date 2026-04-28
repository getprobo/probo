// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerDisplayPageQuery } from "#/__generated__/core/CookieBannerDisplayPageQuery.graphql";

import { CategoryList } from "../_components/CategoryList";

import { ThemePreview } from "./_components/ThemePreview";

export const cookieBannerDisplayPageQuery = graphql`
  query CookieBannerDisplayPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        ...CategoryList_cookieBanner
        ...ThemePreview_cookieBanner
      }
    }
  }
`;

interface CookieBannerDisplayPageProps {
  queryRef: PreloadedQuery<CookieBannerDisplayPageQuery>;
}

export default function CookieBannerDisplayPage({ queryRef }: CookieBannerDisplayPageProps) {
  const data = usePreloadedQuery(cookieBannerDisplayPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-8">
      <CategoryList cookieBannerKey={data.node} />
      <ThemePreview cookieBannerKey={data.node} />
    </div>
  );
}
