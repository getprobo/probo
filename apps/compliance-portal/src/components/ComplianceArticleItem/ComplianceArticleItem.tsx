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
import type { ReactNode } from "react";

import { complianceArticleItem } from "./variants";

export interface ComplianceArticleItemProps {
  // Leading category icon (pass a phosphor icon).
  icon: ReactNode;
  // Primary line.
  title: ReactNode;
  // Optional accent sub-label below the title.
  eyebrow?: ReactNode;
  // Trailing metadata on desktop; stacks under the title on small screens.
  meta?: ReactNode;
}

// A single row of a compliance article list. Encodes the row typography so
// callers pass plain content; wrap rows in a bordered list container.
export function ComplianceArticleItem({ icon, title, eyebrow, meta }: ComplianceArticleItemProps) {
  const slots = complianceArticleItem();

  return (
    <div className={slots.root()}>
      <span className={slots.icon()}>{icon}</span>
      <div className={slots.content()}>
        <div className={slots.text()}>
          <Text size={2} weight="medium" color="neutral" highContrast className={slots.title()}>
            {title}
          </Text>
          {eyebrow != null && (
            <Text size={1} color="gold">
              {eyebrow}
            </Text>
          )}
        </div>
        {meta != null && (
          <Text size={1} color="faint" className={slots.meta()}>
            {meta}
          </Text>
        )}
      </div>
    </div>
  );
}
