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

import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { complianceArticleItem } from "./variants";

// Loading placeholder paired with ComplianceArticleItem: same row layout with
// a pulse icon and skeleton text.
export function ComplianceArticleItemSkeleton() {
  const slots = complianceArticleItem();

  return (
    <div className={slots.root()} aria-hidden>
      <div className={slots.iconPlaceholder()} />
      <div className={slots.content()}>
        <TextSkeleton size={2} className="w-48" />
        <TextSkeleton size={1} className="w-16" />
      </div>
      <TextSkeleton size={1} className="w-20 shrink-0" />
    </div>
  );
}
