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
      <div className="grid grid-cols-3 gap-4 max-lg:grid-cols-2 max-sm:grid-cols-1">
        {subprocessors.map(subprocessor => (
          <SubprocessorListItem key={subprocessor.id} subprocessorKey={subprocessor} />
        ))}
      </div>
    </section>
  );
}
