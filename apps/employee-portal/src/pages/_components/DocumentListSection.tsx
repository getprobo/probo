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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { List } from "@probo/ui/src/v2/List/List";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import type { ReactNode } from "react";

import { documentListSection } from "./variants";

export interface DocumentListSectionProps {
  heading: string;
  count: number;
  empty?: ReactNode;
  summary?: ReactNode;
  hasNext?: boolean;
  isLoadingNext?: boolean;
  loadMoreLabel: string;
  children: ReactNode;
  onLoadMore?: () => void;
}

export function DocumentListSection({
  heading,
  count,
  empty,
  summary,
  hasNext = false,
  isLoadingNext = false,
  loadMoreLabel,
  children,
  onLoadMore,
}: DocumentListSectionProps) {
  const slots = documentListSection();

  return (
    <section className={slots.root()}>
      <Heading level={2} size={2} weight="medium" highContrast className={slots.heading()}>
        {heading}
      </Heading>
      {count === 0
        ? empty
        : (
            <>
              <div className={slots.frame()}>
                {summary}
                <List className={slots.list()}>
                  {children}
                </List>
              </div>
              {hasNext && onLoadMore != null && (
                <div className={slots.more()}>
                  <Button
                    variant="ghost"
                    color="neutral"
                    loading={isLoadingNext}
                    onClick={onLoadMore}
                  >
                    {loadMoreLabel}
                  </Button>
                </div>
              )}
            </>
          )}
    </section>
  );
}
