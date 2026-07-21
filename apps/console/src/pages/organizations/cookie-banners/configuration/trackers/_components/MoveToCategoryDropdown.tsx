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

import { DropdownItem } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MoveToCategoryDropdownQuery } from "#/__generated__/core/MoveToCategoryDropdownQuery.graphql";

export const moveToCategoryDropdownQuery = graphql`
  query MoveToCategoryDropdownQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        categories(first: 50, orderBy: { field: RANK, direction: ASC })
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

interface MoveToCategoryDropdownProps {
  queryRef: PreloadedQuery<MoveToCategoryDropdownQuery>;
  onMove: (categoryId: string) => void;
}

export function MoveToCategoryDropdown({
  queryRef,
  onMove,
}: MoveToCategoryDropdownProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const data = usePreloadedQuery<MoveToCategoryDropdownQuery>(moveToCategoryDropdownQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    return null;
  }

  const categories = data.node.categories.edges.map(e => e.node);

  if (categories.length === 0) {
    return (
      <DropdownItem className="text-sm text-txt-tertiary" disabled>
        {t("moveToCategoryDropdown.empty")}
      </DropdownItem>
    );
  }

  return (
    <>
      {categories.map(cat => (
        <DropdownItem
          className="text-sm"
          key={cat.id}
          onSelect={() => onMove(cat.id)}
        >
          {cat.name}
        </DropdownItem>
      ))}
    </>
  );
}
