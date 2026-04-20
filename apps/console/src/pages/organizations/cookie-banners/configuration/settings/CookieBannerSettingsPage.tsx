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

import type { CookieBannerSettingsPageQuery } from "#/__generated__/core/CookieBannerSettingsPageQuery.graphql";

import { BannerSettingsForm } from "../_components/BannerSettingsForm";
import { CategoryList } from "../_components/CategoryList";

export const cookieBannerSettingsPageQuery = graphql`
  query CookieBannerSettingsPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        ...BannerSettingsForm_cookieBanner
        ...CategoryList_cookieBanner
      }
    }
  }
`;

interface CookieBannerSettingsPageProps {
  queryRef: PreloadedQuery<CookieBannerSettingsPageQuery>;
}

export default function CookieBannerSettingsPage({
  queryRef,
}: CookieBannerSettingsPageProps) {
  const data = usePreloadedQuery(cookieBannerSettingsPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  return (
    <div className="space-y-8">
      <BannerSettingsForm cookieBannerKey={data.node} />
      <CategoryList cookieBannerKey={data.node} />
    </div>
  );
}
