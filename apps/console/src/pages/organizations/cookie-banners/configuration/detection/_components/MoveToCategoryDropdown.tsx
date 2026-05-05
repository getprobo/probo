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
import { DropdownItem } from "@probo/ui";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MoveToCategoryDropdownQuery } from "#/__generated__/core/MoveToCategoryDropdownQuery.graphql";

export const moveToCategoryDropdownQuery = graphql`
  query MoveToCategoryDropdownQuery($cookieBannerId: ID!) {
    node(id: $cookieBannerId) @required(action: THROW) {
      __typename
      ... on CookieBanner {
        consentCategories(first: 50, orderBy: { field: RANK, direction: ASC })
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
  const { __ } = useTranslate();
  const data = usePreloadedQuery(moveToCategoryDropdownQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    return null;
  }

  const categories = data.node.consentCategories.edges.map(e => e.node);

  if (categories.length === 0) {
    return (
      <DropdownItem className="text-sm text-txt-tertiary" disabled>
        {__("No categories")}
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
