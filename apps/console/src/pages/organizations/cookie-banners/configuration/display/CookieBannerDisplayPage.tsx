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
        categories(first: 50, orderBy: { field: RANK, direction: ASC }, filter: { excludeKind: UNCATEGORISED })
          @connection(key: "CookieBannerDisplayPage_categories")
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
  const data = usePreloadedQuery<CookieBannerDisplayPageQuery>(cookieBannerDisplayPageQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    throw new Error("invalid type for node");
  }

  const banner = data.node;
  const connectionId = banner.categories.__id;
  const categories = banner.categories.edges.map(e => e.node);

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
