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

import type { CookieBannerDisplayPageQuery } from "#/__generated__/core/CookieBannerDisplayPageQuery.graphql";

import { CategoryDialog } from "./_components/CategoryDialog";
import { CategorySection } from "./_components/CategorySection";
import { ThemePreview } from "./_components/ThemePreview";

export const cookieBannerDisplayPageQuery = graphql`
  query CookieBannerDisplayPageQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        id
        consentCategories(first: 50, orderBy: { field: RANK, direction: ASC })
          @connection(key: "CookieBannerDisplayPage_consentCategories")
          @required(action: THROW) {
          __id
          edges {
            node {
              id
              rank
              ...CategorySectionFragment
            }
          }
        }
        ...ThemePreview_cookieBanner
      }
    }
  }
`;

interface CookieBannerDisplayPageProps {
  queryRef: PreloadedQuery<CookieBannerDisplayPageQuery>;
}

export default function CookieBannerDisplayPage({
  queryRef,
}: CookieBannerDisplayPageProps) {
  const { __ } = useTranslate();
  const data = usePreloadedQuery(cookieBannerDisplayPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;
  const connectionId = banner.consentCategories.__id;
  const categories = banner.consentCategories.edges.map(e => e.node);

  const [showCreateDialog, setShowCreateDialog] = useState(false);

  return (
    <div className="space-y-8">
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {__("Organize cookies into categories and declare which cookies your site uses.")}
          </p>
          <Button variant="secondary" onClick={() => setShowCreateDialog(true)}>
            <IconPlusSmall size={16} />
            {__("Add Category")}
          </Button>
        </div>

        {categories.length === 0 && (
          <Card className="border p-8 text-center text-muted-foreground">
            {__("No categories yet. Add a category to start managing cookies.")}
          </Card>
        )}

        {categories.map(category => (
          <CategorySection
            key={category.id}
            categoryKey={category}
            connectionId={connectionId}
          />
        ))}

        {showCreateDialog && (
          <CategoryDialog
            cookieBannerId={banner.id}
            connectionId={connectionId}
            nextRank={categories.length > 0 ? categories[categories.length - 1].rank + 1 : 0}
            onOpenChange={setShowCreateDialog}
          />
        )}
      </div>

      <ThemePreview cookieBannerKey={data.node} />
    </div>
  );
}
