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

import { useLazyLoadQuery } from "react-relay";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannerSettingsPageQuery } from "#/__generated__/core/CookieBannerSettingsPageQuery.graphql";

import { BannerSettingsForm } from "../_components/BannerSettingsForm";
import { CategoryList } from "../_components/CategoryList";

const settingsPageQuery = graphql`
  query CookieBannerSettingsPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        name
        origin
        privacyPolicyUrl
        consentExpiryDays
        consentMode
        categories(first: 50, orderBy: { field: RANK, direction: ASC }) {
          __id
          edges {
            node {
              id
              name
              description
              required
              rank
              cookies {
                name
                duration
                description
              }
              createdAt
              updatedAt
            }
          }
        }
      }
    }
  }
`;

export default function CookieBannerSettingsPage() {
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  if (!cookieBannerId) {
    throw new Error("Missing :cookieBannerId param in route");
  }

  const data = useLazyLoadQuery<CookieBannerSettingsPageQuery>(
    settingsPageQuery,
    { cookieBannerId },
  );

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;
  const connectionId = banner.categories.__id;

  return (
    <div className="space-y-8">
      <BannerSettingsForm banner={banner} />

      <CategoryList
        cookieBannerId={banner.id}
        categories={banner.categories.edges.map(e => e.node)}
        connectionId={connectionId}
      />
    </div>
  );
}
