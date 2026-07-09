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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import type { SubprocessorListItem_subprocessor$key } from "./__generated__/SubprocessorListItem_subprocessor.graphql";
import { SubprocessorListItem } from "./SubprocessorListItem";

// A subprocessor edge node: the list-item fragment key plus the fields the page
// reads to group and key the list.
export type SubprocessorNode = SubprocessorListItem_subprocessor$key & {
  readonly id: string;
  readonly category: string;
};

interface SubprocessorCategorySectionProps {
  category: string;
  subprocessors: readonly SubprocessorNode[];
}

// One category group: a localized header (label + description) above a
// responsive grid of subprocessor cards.
export function SubprocessorCategorySection({ category, subprocessors }: SubprocessorCategorySectionProps) {
  const { t } = useTranslation("subprocessors");

  return (
    <section className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <Text size={3} weight="medium" color="neutral" highContrast role="heading" aria-level={2}>
          {t(`categories.${category}.label`)}
        </Text>
        <Text size={2} color="neutral">
          {t(`categories.${category}.description`)}
        </Text>
      </div>
      <div className="grid grid-cols-1 items-start gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {subprocessors.map(subprocessor => (
          <SubprocessorListItem key={subprocessor.id} subprocessorKey={subprocessor} />
        ))}
      </div>
    </section>
  );
}
