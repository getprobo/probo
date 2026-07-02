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
import type { ReactNode } from "react";

import { complianceArticleItem } from "./variants";

export interface ComplianceArticleItemProps {
  // Leading category icon (pass a phosphor icon).
  icon: ReactNode;
  // Primary line.
  title: ReactNode;
  // Optional accent sub-label below the title.
  eyebrow?: ReactNode;
  // Right-aligned metadata (e.g. relative time).
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
        <Text size={2} weight="medium" color="neutral" highContrast>
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
  );
}
