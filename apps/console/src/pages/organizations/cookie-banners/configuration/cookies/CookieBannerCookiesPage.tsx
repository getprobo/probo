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
import { Button, Card, IconPlusSmall } from "@probo/ui";
import { useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CookieBannerCookiesPageQuery } from "#/__generated__/core/CookieBannerCookiesPageQuery.graphql";

import { CategoryDialog } from "../_components/CategoryDialog";

import { CategorySection } from "./_components/CategorySection";

export const cookieBannerCookiesPageQuery = graphql`
  query CookieBannerCookiesPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) {
      __typename
      ... on CookieBanner {
        id
        categories(first: 50, orderBy: { field: RANK, direction: ASC }) @required(action: THROW) {
          __id
          edges {
            node {
              id
              rank
              ...CategorySectionFragment
            }
          }
        }
      }
    }
  }
`;

interface CookieBannerCookiesPageProps {
  queryRef: PreloadedQuery<CookieBannerCookiesPageQuery>;
}

export default function CookieBannerCookiesPage({
  queryRef,
}: CookieBannerCookiesPageProps) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(cookieBannerCookiesPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;
  const connectionId = banner.categories.__id;
  const categories = banner.categories.edges.map(e => e.node);
  const sorted = [...categories].sort((a, b) => a.rank - b.rank);

  const [showCreateDialog, setShowCreateDialog] = useState(false);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">{__("Cookies")}</h3>
        <Button variant="secondary" onClick={() => setShowCreateDialog(true)}>
          <IconPlusSmall size={16} />
          {__("Add Category")}
        </Button>
      </div>

      {sorted.length === 0 && (
        <Card className="border p-8 text-center text-muted-foreground">
          {__("No categories yet. Add a category to start managing cookies.")}
        </Card>
      )}

      {sorted.map(category => (
        <CategorySection key={category.id} categoryKey={category} />
      ))}

      {showCreateDialog && (
        <CategoryDialog
          cookieBannerId={banner.id}
          connectionId={connectionId}
          onOpenChange={setShowCreateDialog}
        />
      )}
    </div>
  );
}
