// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
import { Button, IconPlusLarge, PageHeader, Skeleton } from "@probo/ui";

export function RisksPageSkeleton() {
  const { __ } = useTranslate();
  return (
    <div className="space-y-6">
      <PageHeader title={__("Risks")}>
        <Button icon={IconPlusLarge} disabled>
          {__("New Risk")}
        </Button>
      </PageHeader>
      <div className="grid grid-cols-2 gap-4">
        <Skeleton className="aspect-square" />
        <Skeleton className="aspect-square" />
      </div>
      <div>
        <Skeleton style={{ aspectRatio: "1280/280" }} />
      </div>
    </div>
  );
}
