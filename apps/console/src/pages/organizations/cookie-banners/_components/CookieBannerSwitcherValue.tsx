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

import { graphql, useFragment } from "react-relay";

import type { CookieBannerSwitcherValue_cookieBanner$key } from "#/__generated__/core/CookieBannerSwitcherValue_cookieBanner.graphql";
import type { CookieBannerSwitcherValueQuery$data } from "#/__generated__/core/CookieBannerSwitcherValueQuery.graphql";
import { NavPanelSwitcherValue } from "#/pages/organizations/_components/NavPanelSwitcher";

export const cookieBannerSwitcherValueQuery = graphql`
  query CookieBannerSwitcherValueQuery(
    $organizationId: ID!
    $cookieBannerId: ID!
    $hasCookieBannerId: Boolean!
  ) {
    selected: node(id: $cookieBannerId) @include(if: $hasCookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        ...CookieBannerSwitcherValue_cookieBanner
      }
    }
    organization: node(id: $organizationId) @skip(if: $hasCookieBannerId) {
      __typename
      ... on Organization {
        cookieBanners(first: 1, orderBy: { field: CREATED_AT, direction: DESC })
          @required(action: THROW) {
          edges {
            node {
              id
              ...CookieBannerSwitcherValue_cookieBanner
            }
          }
        }
      }
    }
  }
`;

const cookieBannerSwitcherValueFragment = graphql`
  fragment CookieBannerSwitcherValue_cookieBanner on CookieBanner {
    name
  }
`;

export interface CookieBannerSwitcherValueProps {
  cookieBannerKey: CookieBannerSwitcherValue_cookieBanner$key;
}

export function CookieBannerSwitcherValue({ cookieBannerKey }: CookieBannerSwitcherValueProps) {
  const cookieBanner = useFragment(cookieBannerSwitcherValueFragment, cookieBannerKey);
  return <NavPanelSwitcherValue>{cookieBanner.name}</NavPanelSwitcherValue>;
}

export function cookieBannerFromSwitcherValueQuery(
  data: CookieBannerSwitcherValueQuery$data,
) {
  if (data.selected != null) {
    return data.selected.__typename === "CookieBanner" ? data.selected : null;
  }
  if (data.organization == null) {
    return null;
  }
  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }
  return data.organization.cookieBanners.edges[0]?.node ?? null;
}
