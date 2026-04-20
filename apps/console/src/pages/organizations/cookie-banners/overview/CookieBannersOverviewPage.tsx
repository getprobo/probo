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

import { useTranslate } from "@probo/i18n";
import { Badge, Button, Card } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { CookieBannersOverviewPageQuery } from "#/__generated__/core/CookieBannersOverviewPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CookieBannerEmptyState } from "./_components/CookieBannerEmptyState";

export const cookieBannersOverviewPageQuery = graphql`
  query CookieBannersOverviewPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        cookieBanners(first: 50, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              name
              origin
              state
              createdAt
            }
          }
        }
      }
    }
  }
`;

interface CookieBannersOverviewPageProps {
  queryRef: PreloadedQuery<CookieBannersOverviewPageQuery>;
}

export function CookieBannersOverviewPage({ queryRef }: CookieBannersOverviewPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const { organization } = usePreloadedQuery(cookieBannersOverviewPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const banners = organization.cookieBanners.edges.map(e => e.node);
  const newBannerHref = `/organizations/${organizationId}/cookie-banners/new`;

  if (banners.length === 0) {
    return (
      <CookieBannerEmptyState>
        <Button to={newBannerHref}>{__("Create your first banner")}</Button>
      </CookieBannerEmptyState>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Button to={newBannerHref}>{__("Create Banner")}</Button>
      </div>

      <Card className="divide-y divide-border rounded-lg border">
        {banners.map(banner => (
          <Link
            key={banner.id}
            to={`/organizations/${organizationId}/cookie-banners/${banner.id}`}
            className="flex items-center justify-between gap-4 p-4 hover:bg-muted/50 transition-colors"
          >
            <div className="min-w-0 flex-1">
              <div className="font-medium">{banner.name}</div>
              <div className="text-sm text-muted-foreground truncate">{banner.origin}</div>
            </div>
            <div className="flex items-center gap-3">
              <Badge variant={banner.state === "ACTIVE" ? "success" : "danger"}>
                {banner.state === "ACTIVE" ? __("Active") : __("Inactive")}
              </Badge>
              <span className="text-xs text-muted-foreground">
                {new Date(banner.createdAt).toLocaleDateString()}
              </span>
            </div>
          </Link>
        ))}
      </Card>
    </div>
  );
}
