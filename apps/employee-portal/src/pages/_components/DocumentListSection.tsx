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

import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { documentListSection } from "./variants";

export interface DocumentListSectionProps {
  heading: string;
  count: number;
  empty?: ReactNode;
  summary?: ReactNode;
  // True when a previous page exists; drives the prev arrow.
  hasPrevious: boolean;
  // True when a next page exists; drives the next arrow.
  hasNext: boolean;
  // Dims the list frame while a page change is pending.
  busy?: boolean;
  previousLabel: string;
  nextLabel: string;
  children: ReactNode;
  onPrevious: () => void;
  onNext: () => void;
}

export function DocumentListSection({
  heading,
  count,
  empty,
  summary,
  hasPrevious,
  hasNext,
  busy = false,
  previousLabel,
  nextLabel,
  children,
  onPrevious,
  onNext,
}: DocumentListSectionProps) {
  const { t } = useTranslation();
  const slots = documentListSection({ busy });

  return (
    <section className={slots.root()}>
      <Heading level={2} size={2} weight="medium" highContrast className={slots.heading()}>
        {heading}
      </Heading>
      {count === 0
        ? empty
        : (
            <div className={slots.body()}>
              <div className={slots.frame()} aria-busy={busy || undefined}>
                {summary}
                <Table variant="ghost" className={slots.table()}>
                  <TableHeader className={slots.header()}>
                    <TableRow>
                      <TableColumnHeaderCell>
                        {t("documents.columns.title")}
                      </TableColumnHeaderCell>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {children}
                  </TableBody>
                </Table>
              </div>
              <Pagination
                className={slots.pager()}
                variant="surface"
                showLabels
                hasPrevious={hasPrevious}
                hasNext={hasNext}
                previousLabel={previousLabel}
                nextLabel={nextLabel}
                disabled={busy}
                onPrevious={onPrevious}
                onNext={onNext}
              />
            </div>
          )}
    </section>
  );
}
