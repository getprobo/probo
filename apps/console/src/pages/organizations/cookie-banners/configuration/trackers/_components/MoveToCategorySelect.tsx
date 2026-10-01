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

import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { MoveToCategorySelect_cookieBanner$key } from "#/__generated__/core/MoveToCategorySelect_cookieBanner.graphql";

import { moveToCategorySelect } from "../../../variants";

const moveToCategorySelectFragment = graphql`
  fragment MoveToCategorySelect_cookieBanner on CookieBanner {
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
`;

interface MoveToCategorySelectProps {
  cookieBannerKey: MoveToCategorySelect_cookieBanner$key;
  currentCategoryId?: string;
  currentCategoryName?: string;
  onSelect: (categoryId: string) => void;
}

export function MoveToCategorySelect({
  cookieBannerKey,
  currentCategoryId,
  currentCategoryName,
  onSelect,
}: MoveToCategorySelectProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const cookieBanner = useFragment(moveToCategorySelectFragment, cookieBannerKey);
  const categories = cookieBanner.categories.edges.map(edge => edge.node);
  const { root } = moveToCategorySelect();
  const placeholder = currentCategoryName ?? "-";

  return (
    <div className={root()}>
      <Select
        value={currentCategoryId ?? null}
        onValueChange={(value: string | null) => {
          if (value != null && value !== currentCategoryId) {
            onSelect(value);
          }
        }}
      >
        <SelectTrigger
          size={1}
          variant="surface"
          placeholder={placeholder}
          aria-label={t("trackersPage.columns.category")}
        >
          {(value: string | null) => (
            value != null
              ? categories.find(category => category.id === value)?.name ?? placeholder
              : placeholder
          )}
        </SelectTrigger>
        <SelectPopup align="start">
          {categories.length === 0
            ? (
                <SelectItem value={null} disabled>
                  {t("moveToCategorySelect.empty")}
                </SelectItem>
              )
            : categories.map(category => (
                <SelectItem key={category.id} value={category.id}>
                  {category.name}
                </SelectItem>
              ))}
        </SelectPopup>
      </Select>
    </div>
  );
}
