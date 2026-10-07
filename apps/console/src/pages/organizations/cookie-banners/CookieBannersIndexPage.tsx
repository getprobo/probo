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

import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannersIndexPageQuery } from "#/__generated__/core/CookieBannersIndexPageQuery.graphql";

import { CookieBannerPageHeader } from "./_components/CookieBannerPageHeader";
import { cookieBannerPage } from "./variants";

export const cookieBannersIndexPageQuery = graphql`
  query CookieBannersIndexPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateCookieBanner: permission(action: "core:cookie-banner:create")
        cookieBanners(first: 1, orderBy: { field: CREATED_AT, direction: DESC })
          @required(action: THROW) {
          edges {
            node {
              id
            }
          }
        }
      }
    }
  }
`;

interface CookieBannersIndexPageProps {
  queryRef: PreloadedQuery<CookieBannersIndexPageQuery>;
}

export function CookieBannersIndexPage({ queryRef }: CookieBannersIndexPageProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { organization } = usePreloadedQuery<CookieBannersIndexPageQuery>(
    cookieBannersIndexPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const cookieBannerId = organization.cookieBanners.edges[0]?.node.id;
  if (cookieBannerId == null) {
    if (!organization.canCreateCookieBanner) {
      return (
        <div className={cookieBannerPage()}>
          <CookieBannerPageHeader
            title={t("cookieBannersIndex.empty.title")}
            description={t("cookieBannersIndex.empty.description")}
          />
        </div>
      );
    }
    return <Navigate to="new" replace />;
  }

  return <Navigate to={`${cookieBannerId}/configure`} replace />;
}
