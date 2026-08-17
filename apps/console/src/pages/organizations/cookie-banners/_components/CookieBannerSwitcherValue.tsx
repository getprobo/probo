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

import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type {
  CookieBannerSwitcherValueQuery$data,
  CookieBannerSwitcherValueQuery,
} from "#/__generated__/core/CookieBannerSwitcherValueQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
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
        name
        organization {
          id
        }
      }
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        cookieBanners(first: 1, orderBy: { field: CREATED_AT, direction: DESC })
          @required(action: THROW) {
          edges {
            node {
              id
              name
            }
          }
        }
      }
    }
  }
`;

export interface CookieBannerSwitcherValueProps {
  fallback: string;
  queryRef: PreloadedQuery<CookieBannerSwitcherValueQuery>;
}

export function CookieBannerSwitcherValue({ fallback, queryRef }: CookieBannerSwitcherValueProps) {
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<CookieBannerSwitcherValueQuery>(
    cookieBannerSwitcherValueQuery,
    queryRef,
  );
  const banner = cookieBannerFromSwitcherValueQuery(data, organizationId);

  return <NavPanelSwitcherValue>{banner?.name ?? fallback}</NavPanelSwitcherValue>;
}

// A remembered banner can be deleted or belong to another organization, in
// which case the organization's most recent banner takes over.
export function cookieBannerFromSwitcherValueQuery(
  data: CookieBannerSwitcherValueQuery$data,
  organizationId: string,
) {
  if (
    data.selected?.__typename === "CookieBanner"
    && data.selected.organization?.id === organizationId
  ) {
    return data.selected;
  }
  if (data.organization == null) {
    return null;
  }
  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }
  return data.organization.cookieBanners.edges[0]?.node ?? null;
}
