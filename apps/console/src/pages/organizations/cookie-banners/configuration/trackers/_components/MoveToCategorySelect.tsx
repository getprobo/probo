// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { Option, Select } from "@probo/ui";
import { Suspense, useCallback } from "react";
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
  const { __ } = useTranslate();
  const data = usePreloadedQuery<MoveToCategoryDropdownQuery>(moveToCategoryDropdownQuery, queryRef);

  if (data.node.__typename !== "CookieBanner") {
    return null;
  }

  const categories = data.node.categories.edges.map(e => e.node);

  if (categories.length === 0) {
    return (
      <Option value="" disabled className="text-txt-tertiary">
        {__("No categories")}
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
