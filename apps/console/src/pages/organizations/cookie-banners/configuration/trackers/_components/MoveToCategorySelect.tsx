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

import { Option, Select } from "@probo/ui";
import { Suspense, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { MoveToCategoryDropdownQuery } from "#/__generated__/core/MoveToCategoryDropdownQuery.graphql";

import { moveToCategoryDropdownQuery } from "./MoveToCategoryDropdown";

interface MoveToCategorySelectProps {
  currentCategoryId?: string;
  currentCategoryName?: string;
  highlight?: boolean;
  onSelect: (categoryId: string) => void;
}

export function MoveToCategorySelect({
  currentCategoryId,
  currentCategoryName,
  highlight = false,
  onSelect,
}: MoveToCategorySelectProps) {
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  const [categoryQueryRef, loadCategoryQuery]
    = useQueryLoader<MoveToCategoryDropdownQuery>(moveToCategoryDropdownQuery);

  const handleOpenChange = useCallback(
    (open: boolean) => {
      if (open && cookieBannerId) {
        loadCategoryQuery({ cookieBannerId });
      }
    },
    [loadCategoryQuery, cookieBannerId],
  );

  const handleValueChange = useCallback(
    (categoryId: string) => {
      if (categoryId !== currentCategoryId) {
        onSelect(categoryId);
      }
    },
    [currentCategoryId, onSelect],
  );

  return (
    <Select
      variant={highlight ? "editor" : "ghost"}
      className={highlight ? undefined : "px-0"}
      placeholder={currentCategoryName ?? <span className="text-txt-tertiary">-</span>}
      onValueChange={handleValueChange}
      onOpenChange={handleOpenChange}
    >
      {categoryQueryRef && (
        <Suspense>
          <MoveToCategoryOptions queryRef={categoryQueryRef} />
        </Suspense>
      )}
    </Select>
  );
}

interface MoveToCategoryOptionsProps {
  queryRef: PreloadedQuery<MoveToCategoryDropdownQuery>;
}

function MoveToCategoryOptions({ queryRef }: MoveToCategoryOptionsProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const data = usePreloadedQuery<MoveToCategoryDropdownQuery>(moveToCategoryDropdownQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    return null;
  }

  const categories = data.node.categories.edges.map(e => e.node);

  if (categories.length === 0) {
    return (
      <Option value="" disabled className="text-txt-tertiary">
        {t("moveToCategorySelect.empty")}
      </Option>
    );
  }

  return (
    <>
      {categories.map(cat => (
        <Option key={cat.id} value={cat.id}>
          {cat.name}
        </Option>
      ))}
    </>
  );
}
